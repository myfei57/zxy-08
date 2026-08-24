package store

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load replaces the current index with the contents of one snapshot file.
func (s *Store) Load(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("store: read snapshot %s: %w", path, err)
	}
	var snap snapshotFile
	if err := json.Unmarshal(data, &snap); err != nil {
		return fmt.Errorf("store: parse snapshot %s: %w", path, err)
	}
	for _, v := range snap.Entries {
		if v.Seq > s.seq {
			s.seq = v.Seq
		}
		s.index[v.Key] = &Value{
			Key:     v.Key,
			Value:   append([]byte(nil), v.Value...),
			Version: v.Version,
			Expires: v.Expires,
			Seq:     v.Seq,
		}
	}
	return nil
}
