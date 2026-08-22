package replica

import "fmt"

// Resync brings the replica up to date from the primary journal.
func (m *Manager) Resync(primaryJournal string) error {
	ops, err := readJournal(primaryJournal)
	if err != nil {
		return fmt.Errorf("replica: read primary journal: %w", err)
	}
	offset, err := m.st.AppliedOffset()
	if err != nil {
		return fmt.Errorf("replica: applied offset: %w", err)
	}
	client := &Client{target: m.st}
	for _, op := range ops {
		if op.Seq <= offset {
			continue
		}
		if err := client.ApplyRemote(op); err != nil {
			return fmt.Errorf("replica: apply %s: %w", op.Key, err)
		}
	}
	return nil
}
