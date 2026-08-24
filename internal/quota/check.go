package quota

import "fmt"

// Check rejects a write that would exceed the capacity. It must run before the
// write touches storage; otherwise an over-quota value is stored before the
// error is returned. An overwrite is gated only on its net growth, so the key's
// existing footprint is credited back before comparing against the capacity.
func (m *Manager) Check(key string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	old := m.sizes[key]
	if m.used-old+size > m.capacity {
		return fmt.Errorf("quota: capacity exceeded (%d - %d + %d > %d)", m.used, old, size, m.capacity)
	}
	return nil
}

// Account records a key's size after a successful write, crediting back any
// prior footprint so an overwrite is not double-counted.
func (m *Manager) Account(key string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if size < 0 {
		return fmt.Errorf("quota: negative size for %s", key)
	}
	if old, ok := m.sizes[key]; ok {
		m.used -= old
	}
	m.sizes[key] = size
	m.used += size
	return nil
}
