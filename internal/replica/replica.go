// Package replica implements acknowledgement and resync of replicated writes.
package replica

import "kvgrid/internal/store"

// Manager resynchronizes a replica store from the primary journal.
type Manager struct {
	st *store.Store
}

// NewManager wraps the replica's local store.
func NewManager(st *store.Store) *Manager {
	return &Manager{st: st}
}
