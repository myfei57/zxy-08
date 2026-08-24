// Package snapshot manages generation-consistent snapshot persistence and restore.
package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"kvgrid/internal/audit"
	"kvgrid/internal/store"
)

// Manager owns the snapshot directory and the cursor that marks the durable
// generation.
type Manager struct {
	st          *store.Store
	snapshotDir string
	cursorPath  string
	audit       *audit.Logger
}

// NewManager creates a snapshot manager for a store.
func NewManager(st *store.Store, snapshotDir string, cursorPath string, audit *audit.Logger) *Manager {
	return &Manager{st: st, snapshotDir: snapshotDir, cursorPath: cursorPath, audit: audit}
}

// Take writes a new snapshot generation.
func (m *Manager) Take() error {
	gen, err := m.CurrentGeneration()
	if err != nil {
		return err
	}
	gen++
	target := filepath.Join(m.snapshotDir, fmt.Sprintf("gen-%d", gen), "snapshot.json")
	// The snapshot data file is flushed and published before the cursor is
	// advanced, so a crash in between leaves recovery pointing at the previous
	// (complete) generation instead of trusting a half-written snapshot.
	if err := m.st.Dump(target); err != nil {
		return fmt.Errorf("snapshot: dump generation %d: %w", gen, err)
	}
	if err := m.WriteCursor(gen); err != nil {
		return fmt.Errorf("snapshot: cursor: %w", err)
	}
	return m.audit.Note("snapshot", "", "", fmt.Sprintf("generation %d", gen))
}

// SnapshotFiles lists the snapshot generations present on disk.
func (m *Manager) SnapshotFiles() []string {
	entries, err := os.ReadDir(m.snapshotDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[:4] == "gen-" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}
