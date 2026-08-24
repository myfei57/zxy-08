package evict

import "fmt"

// Run evicts a single key.
func (e *Evictor) Run(key string) error {
	if _, ok := e.quota.Size(key); !ok {
		return fmt.Errorf("evictor: %s has no quota record", key)
	}
	// The key is dropped from the index before the quota release is durable;
	// a failed release then leaves the capacity permanently missing.
	if err := e.st.Delete(key); err != nil {
		return fmt.Errorf("evictor: delete key: %w", err)
	}
	if err := e.quota.Release(key); err != nil {
		return fmt.Errorf("evictor: release quota: %w", err)
	}
	return e.audit.Note("evict", "", key, "quota released before removal")
}
