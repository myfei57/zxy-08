package shard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Ownership persists the shard table and publishes consistent snapshots.
type Ownership struct {
	path  string
	table *Table
}

// NewOwnership creates a handle to the ownership file.
func NewOwnership(path string) *Ownership {
	return &Ownership{path: path}
}

// Load reads the persisted table, if any.
func (o *Ownership) Load() error {
	data, err := os.ReadFile(o.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("ownership: read %s: %w", o.path, err)
	}
	var t Table
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("ownership: parse %s: %w", o.path, err)
	}
	o.table = &t
	return nil
}

// Table returns the current ownership view.
func (o *Ownership) Table() *Table {
	if o.table == nil {
		return &Table{Owners: map[string]string{}}
	}
	return o.table
}

// SaveTable atomically publishes a complete table.
func (o *Ownership) SaveTable(t *Table) error {
	if err := os.MkdirAll(filepath.Dir(o.path), 0o755); err != nil {
		return fmt.Errorf("ownership: mkdir: %w", err)
	}
	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("ownership: encode: %w", err)
	}
	tmp := o.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("ownership: write tmp: %w", err)
	}
	if err := os.Rename(tmp, o.path); err != nil {
		return fmt.Errorf("ownership: publish: %w", err)
	}
	o.table = t.Clone()
	return nil
}

// Move reassigns one shard to a new owner and publishes the change.
func (o *Ownership) Move(shardID string, to string) error {
	next := o.Table().Clone()
	if _, ok := next.OwnerOf(shardID); !ok {
		return fmt.Errorf("ownership: unknown shard %s", shardID)
	}
	next.Owners[shardID] = to
	return o.SaveTable(next)
}
