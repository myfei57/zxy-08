// Package cluster wires nodes, shards, storage, replication and the console.
package cluster

import (
	"fmt"
	"os"

	"kvgrid/internal/audit"
	"kvgrid/internal/config"
	"kvgrid/internal/console"
	"kvgrid/internal/evict"
	"kvgrid/internal/expire"
	"kvgrid/internal/node"
	"kvgrid/internal/quota"
	"kvgrid/internal/route"
	"kvgrid/internal/shard"
	"kvgrid/internal/snapshot"
	"kvgrid/internal/store"
)

// nodeIDs are the fixed member identities of the cluster.
var nodeIDs = []string{"node-a", "node-b", "node-c"}

// Cluster is the assembled KVGrid process.
type Cluster struct {
	Cfg      config.Config
	Reg      *node.Registry
	Own      *shard.Ownership
	Stores   map[string]*store.Store
	Quota    *quota.Manager
	Router   *route.Router
	Migrator *shard.Migrator
	Splitter *shard.Splitter
	Snapshot *snapshot.Manager
	Scanner  *expire.Scanner
	Evictor  *evict.Evictor
	Audit    *audit.Logger
	Server   *console.Server
}

// Build creates the cluster and its on-disk state.
func Build(cfg config.Config) (*Cluster, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("cluster: mkdir data: %w", err)
	}
	reg := node.NewRegistry()
	for _, id := range nodeIDs {
		if err := reg.Register(node.New(id, "127.0.0.1:0")); err != nil {
			return nil, err
		}
	}
	stores := make(map[string]*store.Store)
	for _, id := range nodeIDs {
		opts := store.Options{
			DataDir: config.NodeDataDir(cfg.DataDir, id),
			MetaDir: config.MetaDir(cfg.DataDir, id),
		}
		if err := os.MkdirAll(opts.MetaDir, 0o755); err != nil {
			return nil, err
		}
		st, err := store.NewStore(opts)
		if err != nil {
			return nil, err
		}
		stores[id] = st
	}
	own := shard.NewOwnership(config.OwnershipPath(cfg.DataDir))
	if err := own.Load(); err != nil {
		return nil, err
	}
	if len(own.Table().Shards) == 0 {
		if err := own.SaveTable(shard.BuildInitial(cfg.ShardCount, nodeIDs)); err != nil {
			return nil, err
		}
	}
	auditLogger := audit.NewLogger(config.AuditPath(cfg.DataDir))
	quotaManager := quota.NewManager(cfg.QuotaBytes, config.QuotaLedgerPath(cfg.DataDir))
	migrator := shard.NewMigrator(reg, own, stores, auditLogger)
	router := route.NewRouter(own, stores, quotaManager, reg, migrator, auditLogger)
	snapshotManager := snapshot.NewManager(
		stores[cfg.NodeID],
		config.SnapshotDir(cfg.DataDir),
		config.SnapshotCursorPath(cfg.DataDir),
		auditLogger,
	)
	scanner := expire.NewScanner(stores[cfg.NodeID], config.ExpireCursorPath(cfg.DataDir), 100, auditLogger)
	evictor := evict.NewEvictor(stores[cfg.NodeID], quotaManager, auditLogger)
	c := &Cluster{
		Cfg:      cfg,
		Reg:      reg,
		Own:      own,
		Stores:   stores,
		Quota:    quotaManager,
		Router:   router,
		Migrator: migrator,
		Splitter: shard.NewSplitter(auditLogger),
		Snapshot: snapshotManager,
		Scanner:  scanner,
		Evictor:  evictor,
		Audit:    auditLogger,
	}
	c.Server = console.NewServer(c)
	return c, nil
}

// Start serves the console until the process exits.
func (c *Cluster) Start() error {
	_ = c.Audit.Note("start", c.Cfg.NodeID, "", c.Cfg.HTTPAddr)
	return c.Server.Start(c.Cfg.HTTPAddr)
}

// Close flushes every node store.
func (c *Cluster) Close() error {
	for _, id := range nodeIDs {
		if st := c.Stores[id]; st != nil {
			if err := st.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
