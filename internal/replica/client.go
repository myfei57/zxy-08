package replica

import (
	"encoding/json"
	"os"
	"strings"

	"kvgrid/internal/store"
)

// Client applies operations received from a primary to a target store.
type Client struct {
	target *store.Store
}

// ApplyRemote applies a single replicated operation.
func (c *Client) ApplyRemote(op store.Op) error {
	return c.target.Apply(op)
}

// readJournal parses a primary journal file into operations.
func readJournal(path string) ([]store.Op, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ops []store.Op
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var op store.Op
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}
