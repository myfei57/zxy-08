package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type cursorFile struct {
	Generation uint64 `json:"generation"`
}

// CurrentGeneration returns the last committed snapshot generation.
func (m *Manager) CurrentGeneration() (uint64, error) {
	data, err := os.ReadFile(m.cursorPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var c cursorFile
	if err := json.Unmarshal(data, &c); err != nil {
		return 0, fmt.Errorf("snapshot: parse cursor: %w", err)
	}
	return c.Generation, nil
}

// WriteCursor durably records the current snapshot generation.
func (m *Manager) WriteCursor(gen uint64) error {
	if err := os.MkdirAll(filepath.Dir(m.cursorPath), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cursorFile{Generation: gen})
	if err != nil {
		return err
	}
	tmp := m.cursorPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, m.cursorPath)
}
