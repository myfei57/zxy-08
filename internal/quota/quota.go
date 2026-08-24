// Package quota tracks cluster capacity and releases capacity durably on eviction.
package quota

import "sync"

// Manager enforces a hard capacity bound across the cluster.
type Manager struct {
	mu         sync.Mutex
	capacity   int64
	used       int64
	sizes      map[string]int64
	ledgerPath string
}

// NewManager creates a quota manager with the given capacity.
func NewManager(capacity int64, ledgerPath string) *Manager {
	return &Manager{capacity: capacity, sizes: make(map[string]int64), ledgerPath: ledgerPath}
}

// Capacity returns the configured capacity.
func (m *Manager) Capacity() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.capacity
}

// Used returns the currently occupied capacity.
func (m *Manager) Used() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.used
}

// Size returns the recorded size of a key.
func (m *Manager) Size(key string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	size, ok := m.sizes[key]
	return size, ok
}
