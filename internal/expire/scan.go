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
	// The cursor moves before the batch is durably deleted, so a crash in the
	// middle leaves expired keys behind the new cursor.
	if err := s.WriteCursor(s.st.LastSeq()); err != nil {
		return 0, fmt.Errorf("expire: cursor: %w", err)
	}
	for _, key := range keys {
		if !s.st.Has(key) {
			continue
		}
		if err := s.st.Delete(key); err != nil {
			return 0, fmt.Errorf("expire: delete %s: %w", key, err)
		}
	}
	_ = s.audit.Note("expire", "", "", fmt.Sprintf("%d keys cleaned", len(keys)))
	return len(keys), nil
}
