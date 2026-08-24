// Package route locates keys, gates writes on capacity and coordinates drains.
package route

import (
	"kvgrid/internal/audit"
	"kvgrid/internal/node"
	"kvgrid/internal/quota"
	"kvgrid/internal/shard"
	"kvgrid/internal/store"
)

// Router dispatches reads and writes to the owning node of each key.
type Router struct {
	own    *shard.Ownership
	stores map[string]*store.Store
	quota  *quota.Manager
	reg    *node.Registry
	mig    *shard.Migrator
	audit  *audit.Logger
	cached map[string]string
}

// NewRouter creates a router bound to the cluster state.
func NewRouter(
	own *shard.Ownership,
	stores map[string]*store.Store,
	q *quota.Manager,
	reg *node.Registry,
	mig *shard.Migrator,
	audit *audit.Logger,
) *Router {
	cached := make(map[string]string)
	for _, sh := range own.Table().Shards {
		if owner, ok := own.Table().OwnerOf(sh.ID); ok {
			cached[sh.ID] = owner
		}
	}
	return &Router{own: own, stores: stores, quota: q, reg: reg, mig: mig, audit: audit, cached: cached}
}
