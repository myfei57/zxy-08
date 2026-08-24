package replica

import "fmt"

// Resync brings the replica up to date from the primary journal.
func (m *Manager) Resync(primaryJournal string) error {
	ops, err := readJournal(primaryJournal)
	if err != nil {
		return fmt.Errorf("replica: read primary journal: %w", err)
	}
	client := &Client{target: m.st}
	// The whole log is replayed from the beginning; already-applied operations
	// are committed a second time and can overwrite newer values.
	for _, op := range ops {
		if err := client.ApplyRemote(op); err != nil {
			return fmt.Errorf("replica: apply %s: %w", op.Key, err)
		}
	}
	return nil
}
