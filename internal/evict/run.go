package evict

import "fmt"

// Run evicts a single key: the quota release must be durable first.
func (e *Evictor) Run(key string) error {
	if _, ok := e.quota.Size(key); !ok {
		return fmt.Errorf("evictor: %s has no quota record", key)
	}
	if err := e.quota.Release(key); err != nil {
		return fmt.Errorf("evictor: release quota: %w", err)
	}
	if err := e.st.Delete(key); err != nil {
		return fmt.Errorf("evictor: delete key: %w", err)
	}
	return e.audit.Note("evict", "", key, "quota released before removal")
}
