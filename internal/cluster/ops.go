package cluster

import (
	"fmt"
	"path/filepath"
	"time"

	"kvgrid/internal/audit"
	"kvgrid/internal/config"
	"kvgrid/internal/console"
	"kvgrid/internal/replica"
	"kvgrid/internal/shard"
)

// Write stores a key with an optional TTL in milliseconds.
func (c *Cluster) Write(key string, value []byte, ttlMs int64) error {
	expires := int64(0)
	if ttlMs > 0 {
		expires = time.Now().UnixNano() + ttlMs*int64(time.Millisecond)
	}
	return c.Router.Set(key, value, expires)
}

// Read fetches a key from its owner.
func (c *Cluster) Read(key string) ([]byte, bool) {
	return c.Router.Get(key)
}

// Delete removes a key from its owner.
func (c *Cluster) Delete(key string) error {
	owner, err := c.Router.Locate(key)
	if err != nil {
		return err
	}
	st := c.Stores[owner]
	if st == nil {
		return fmt.Errorf("cluster: no store for %s", owner)
	}
	if !st.Has(key) {
		return fmt.Errorf("cluster: key %s not found on %s", key, owner)
	}
	if err := st.Delete(key); err != nil {
		return err
	}
	return c.Audit.Note("delete", owner, key, "")
}

// Topology returns the node and shard layout for the console.
func (c *Cluster) Topology() console.Topology {
	var nodes []console.NodeInfo
	counts := c.Own.Table().CountByOwner()
	for _, id := range c.Reg.ActiveIDs() {
		n, _ := c.Reg.Get(id)
		nodes = append(nodes, console.NodeInfo{
			ID:      n.ID,
			Address: n.Address,
			State:   string(n.State),
			Keys:    c.Stores[id].Count(),
			Shards:  counts[id],
		})
	}
	var shards []console.ShardInfo
	owners := map[string]string{}
	for _, sh := range c.Own.Table().Shards {
		owner, _ := c.Own.Table().OwnerOf(sh.ID)
		owners[sh.ID] = owner
		shards = append(shards, console.ShardInfo{ID: sh.ID, Start: sh.Start, End: sh.End, Owner: owner})
	}
	return console.Topology{Nodes: nodes, Shards: shards, Owners: owners}
}

// Keys lists stored keys across every node.
func (c *Cluster) Keys(prefix string) []console.KeyInfo {
	var out []console.KeyInfo
	for _, id := range nodeIDs {
		st := c.Stores[id]
		if st == nil {
			continue
		}
		for _, key := range st.Keys(prefix) {
			v, _ := st.Entry(key)
			out = append(out, console.KeyInfo{
				Key:     key,
				Bytes:   len(v.Value),
				Version: v.Version,
				Expires: v.Expires,
			})
		}
	}
	return out
}

// KeyDetail returns the routing and stored value of one key.
func (c *Cluster) KeyDetail(key string) (console.KeyDetail, error) {
	info, err := c.Router.RouteInfo(key)
	if err != nil {
		return console.KeyDetail{}, err
	}
	value, ok := c.Read(key)
	if !ok {
		return console.KeyDetail{}, fmt.Errorf("cluster: key %s not found", key)
	}
	return console.KeyDetail{
		Key:   info.Key,
		Hash:  info.Hash,
		Shard: info.ShardID,
		Owner: info.Owner,
		Value: string(value),
	}, nil
}

// Snapshots lists the snapshot generations on disk.
func (c *Cluster) Snapshots() []console.SnapshotInfo {
	names := c.Snapshot.SnapshotFiles()
	out := make([]console.SnapshotInfo, 0, len(names))
	for _, name := range names {
		out = append(out, console.SnapshotInfo{
			Generation: name,
			Path:       filepath.Join(config.SnapshotDir(c.Cfg.DataDir), name),
		})
	}
	return out
}

// AuditEntries returns the newest audit events.
func (c *Cluster) AuditEntries(limit int) []audit.Event {
	return c.Audit.Entries(limit)
}

// AuditCount returns the number of events recorded in this process.
func (c *Cluster) AuditCount() int {
	return c.Audit.Count()
}

// AuditCounts returns event counts grouped by type.
func (c *Cluster) AuditCounts() map[string]int {
	return c.Audit.Counts()
}

// QuotaStatus returns the capacity usage for the console.
func (c *Cluster) QuotaStatus() console.QuotaInfo {
	return console.QuotaInfo{Capacity: c.Quota.Capacity(), Used: c.Quota.Used()}
}

// TakeSnapshot snapshots the local store and commits a new generation.
func (c *Cluster) TakeSnapshot() error {
	return c.Snapshot.Take()
}

// Restore reloads the newest snapshot generation.
func (c *Cluster) Restore() error {
	return c.Snapshot.Restore(config.SnapshotDir(c.Cfg.DataDir))
}

// Rebalance drains the source node and completes the migration.
func (c *Cluster) Rebalance(from string, to string) error {
	plan, err := shard.BuildPlan(c.Reg, c.Own, from, to)
	if err != nil {
		return err
	}
	if err := c.Migrator.StartDrain(plan); err != nil {
		return err
	}
	return c.Migrator.Finish()
}

// SplitShard splits a shard in half under a new owner.
func (c *Cluster) SplitShard(shardID string, owner string) error {
	sh, ok := c.Own.Table().ByID(shardID)
	if !ok {
		return fmt.Errorf("cluster: unknown shard %s", shardID)
	}
	mid := sh.Start + (sh.End-sh.Start)/2
	return c.Splitter.Split(c.Own, shardID, mid, owner)
}

// RunExpire scans and deletes one batch of expired keys.
func (c *Cluster) RunExpire() (int, error) {
	return c.Scanner.Scan(time.Now().UnixNano())
}

// RunEvict evicts the oldest keys until capacity is released.
func (c *Cluster) RunEvict() (int, error) {
	picked := c.Evictor.Pick(10)
	n := 0
	for _, key := range picked {
		if err := c.Evictor.Run(key); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// Resync replays the primary journal into the given replica node.
func (c *Cluster) Resync(nodeID string) error {
	primary, ok := c.Reg.FirstActive()
	if !ok {
		return fmt.Errorf("cluster: no active primary")
	}
	if primary == nodeID {
		return fmt.Errorf("cluster: %s is already the primary", nodeID)
	}
	journal := c.Stores[primary].JournalPath()
	return replica.NewManager(c.Stores[nodeID]).Resync(journal)
}
