package expire

import "fmt"

// Scan removes one batch of expired keys.
func (s *Scanner) Scan(now int64) (int, error) {
	keys := s.st.ExpiredKeys(now)
	if len(keys) > s.batchSize {
		keys = keys[:s.batchSize]
	}
	if len(keys) == 0 {
		return 0, nil
	}
	for _, key := range keys {
		if !s.st.Has(key) {
			continue
		}
		if err := s.st.Delete(key); err != nil {
			return 0, fmt.Errorf("expire: delete %s: %w", key, err)
		}
	}
	// Advance the cursor only after every deletion in the batch is durable.
	// Moving it earlier leaves the cursor past keys whose tombstones have not
	// been written yet, so a crash mid-batch would skip them on resume.
	if err := s.WriteCursor(s.st.LastSeq()); err != nil {
		return 0, fmt.Errorf("expire: cursor: %w", err)
	}
	_ = s.audit.Note("expire", "", "", fmt.Sprintf("%d keys cleaned", len(keys)))
	return len(keys), nil
}
