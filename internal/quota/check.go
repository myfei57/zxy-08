package quota

import "fmt"

// Check rejects a write that would exceed the capacity.
func (m *Manager) Check(key string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.used+size > m.capacity {
		return fmt.Errorf("quota: capacity exceeded (%d + %d > %d)", m.used, size, m.capacity)
	}
	return nil
}

// Account records a key's size after a successful write.
func (m *Manager) Account(key string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if size < 0 {
		return fmt.Errorf("quota: negative size for %s", key)
	}
	m.sizes[key] = size
	m.used += size
	return nil
}
