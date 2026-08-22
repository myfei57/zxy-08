// Package evict frees capacity by removing keys after their quota is released.
package evict

import (
	"kvgrid/internal/audit"
	"kvgrid/internal/quota"
	"kvgrid/internal/store"
)

// Evictor picks keys and coordinates quota release with index removal.
type Evictor struct {
	st    *store.Store
	quota *quota.Manager
	audit *audit.Logger
}

// NewEvictor creates an evictor bound to a store and quota manager.
func NewEvictor(st *store.Store, q *quota.Manager, audit *audit.Logger) *Evictor {
	return &Evictor{st: st, quota: q, audit: audit}
}
