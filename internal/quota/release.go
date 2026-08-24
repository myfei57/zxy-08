package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type releaseEntry struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

// Release durably records the quota release before freeing the capacity.
func (m *Manager) Release(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	size, ok := m.sizes[key]
	if !ok {
		return fmt.Errorf("quota: no usage recorded for %s", key)
	}
	entry := releaseEntry{Key: key, Size: size}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if m.ledgerPath != "" {
		if err := os.MkdirAll(filepath.Dir(m.ledgerPath), 0o755); err != nil {
			return fmt.Errorf("quota: ledger dir: %w", err)
		}
		file, err := os.OpenFile(m.ledgerPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("quota: open ledger: %w", err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			file.Close()
			return fmt.Errorf("quota: ledger write: %w", err)
		}
		if err := file.Sync(); err != nil {
			file.Close()
			return fmt.Errorf("quota: ledger sync: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("quota: ledger close: %w", err)
		}
	}
	delete(m.sizes, key)
	m.used -= size
	return nil
}
