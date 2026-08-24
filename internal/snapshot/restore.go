package snapshot

import (
	"fmt"
	"path/filepath"
)

// Restore loads the snapshot generation recorded by the cursor.
func (m *Manager) Restore(snapshotDir string) error {
	gen, err := m.CurrentGeneration()
	if err != nil {
		return err
	}
	if gen == 0 {
		return fmt.Errorf("snapshot: no snapshot generation recorded")
	}
	path := filepath.Join(snapshotDir, fmt.Sprintf("gen-%d", gen), "snapshot.json")
	if err := m.st.Load(path); err != nil {
		return fmt.Errorf("snapshot: restore generation %d: %w", gen, err)
	}
	return m.audit.Note("restore", "", "", fmt.Sprintf("generation %d", gen))
}
