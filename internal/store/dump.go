package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Dump writes the current index to a snapshot file.
func (s *Store) Dump(targetPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("store: snapshot dir: %w", err)
	}
	entries := make([]Value, 0, len(s.index))
	for _, v := range s.index {
		entries = append(entries, *v)
	}
	data, err := json.Marshal(snapshotFile{Entries: entries})
	if err != nil {
		return err
	}
	tmp := targetPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("store: snapshot write: %w", err)
	}
	if err := syncPath(tmp); err != nil {
		return err
	}
	if err := os.Rename(tmp, targetPath); err != nil {
		return fmt.Errorf("store: snapshot publish: %w", err)
	}
	return nil
}
