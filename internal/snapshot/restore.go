package snapshot

import (
	"fmt"
	"path/filepath"
)

// Restore loads snapshot generations without validating the recorded one.
func (m *Manager) Restore(snapshotDir string) error {
	gen, err := m.CurrentGeneration()
	if err != nil {
		return err
	}
	if gen == 0 {
		return fmt.Errorf("snapshot: no snapshot generation recorded")
	}
	// Every generation down to the first is loaded, so an older generation can
	// overwrite values that the newer generation already restored.
	for g := gen; g >= 1; g-- {
		path := filepath.Join(snapshotDir, fmt.Sprintf("gen-%d", g), "snapshot.json")
		if err := m.st.Load(path); err != nil {
			return fmt.Errorf("snapshot: restore generation %d: %w", g, err)
		}
	}
	return m.audit.Note("restore", "", "", fmt.Sprintf("generation %d", gen))
}
