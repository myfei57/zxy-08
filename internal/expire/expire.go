// Package expire scans keys whose TTL elapsed and advances the cleanup cursor.
package expire

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"kvgrid/internal/audit"
	"kvgrid/internal/store"
)

type cursor struct {
	Seq uint64 `json:"seq"`
}

// Scanner removes expired keys in batches and persists a cursor.
type Scanner struct {
	st         *store.Store
	cursorPath string
	batchSize  int
	audit      *audit.Logger
}

// NewScanner creates a scanner that commits batches before moving its cursor.
func NewScanner(st *store.Store, cursorPath string, batchSize int, audit *audit.Logger) *Scanner {
	return &Scanner{st: st, cursorPath: cursorPath, batchSize: batchSize, audit: audit}
}

// ReadCursor returns the last committed scan position.
func (s *Scanner) ReadCursor() (uint64, error) {
	data, err := os.ReadFile(s.cursorPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var c cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return 0, fmt.Errorf("expire: parse cursor: %w", err)
	}
	return c.Seq, nil
}

// WriteCursor persists the scan position.
func (s *Scanner) WriteCursor(seq uint64) error {
	if err := os.MkdirAll(filepath.Dir(s.cursorPath), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cursor{Seq: seq})
	if err != nil {
		return err
	}
	tmp := s.cursorPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.cursorPath)
}
