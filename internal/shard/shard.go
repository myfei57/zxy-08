// Package shard implements hash-range shards, the ownership table and the
// rebalance/split coordination of the KVGrid cluster.
package shard

import (
	"fmt"
	"hash/fnv"
)

// Shard is a contiguous slice of the 32-bit hash space.
type Shard struct {
	ID    string `json:"id"`
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

// Contains reports whether a hash value falls inside the shard range.
func (s Shard) Contains(hash uint32) bool {
	return hash >= s.Start && hash < s.End
}

// HashOf returns the fnv32a hash used for key placement.
func HashOf(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

// Table is an immutable snapshot of the shard layout and ownership.
type Table struct {
	Shards []Shard           `json:"shards"`
	Owners map[string]string `json:"owners"`
}

// Clone returns a deep copy of the table.
func (t *Table) Clone() *Table {
	next := &Table{
		Shards: append([]Shard(nil), t.Shards...),
		Owners: make(map[string]string, len(t.Owners)),
	}
	for id, owner := range t.Owners {
		next.Owners[id] = owner
	}
	return next
}

// OwnerOf returns the node currently owning a shard.
func (t *Table) OwnerOf(shardID string) (string, bool) {
	owner, ok := t.Owners[shardID]
	return owner, ok
}

// ShardFor finds the shard covering a hash value.
func (t *Table) ShardFor(hash uint32) (*Shard, bool) {
	for i := range t.Shards {
		if t.Shards[i].Contains(hash) {
			return &t.Shards[i], true
		}
	}
	return nil, false
}

// ByID finds the shard with the given id.
func (t *Table) ByID(id string) (*Shard, bool) {
	for i := range t.Shards {
		if t.Shards[i].ID == id {
			return &t.Shards[i], true
		}
	}
	return nil, false
}

// Remove deletes a shard from the table.
func (t *Table) Remove(shardID string) {
	filtered := t.Shards[:0]
	for _, s := range t.Shards {
		if s.ID != shardID {
			filtered = append(filtered, s)
		}
	}
	t.Shards = filtered
	delete(t.Owners, shardID)
}

// Replace swaps one shard for two child shards owned by a single node.
func (t *Table) Replace(remove string, left Shard, right Shard, owner string) {
	t.Remove(remove)
	t.Shards = append(t.Shards, left, right)
	t.Owners[left.ID] = owner
	t.Owners[right.ID] = owner
}

// CountByOwner returns how many shards each node owns.
func (t *Table) CountByOwner() map[string]int {
	counts := map[string]int{}
	for _, owner := range t.Owners {
		counts[owner]++
	}
	return counts
}

// BuildInitial creates a balanced table across the given nodes.
func BuildInitial(shardCount int, nodeIDs []string) *Table {
	t := &Table{Owners: make(map[string]string)}
	span := ^uint32(0) / uint32(shardCount)
	for i := 0; i < shardCount; i++ {
		id := fmt.Sprintf("shard-%d", i)
		start := uint32(i) * span
		end := start + span
		if i == shardCount-1 {
			end = ^uint32(0)
		}
		t.Shards = append(t.Shards, Shard{ID: id, Start: start, End: end})
		t.Owners[id] = nodeIDs[i%len(nodeIDs)]
	}
	return t
}
