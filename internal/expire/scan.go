package expire

import "fmt"

// Scan durably deletes one batch of expired keys and then advances the cursor.
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
	if err := s.WriteCursor(s.st.LastSeq()); err != nil {
		return 0, fmt.Errorf("expire: cursor: %w", err)
	}
	cursor, err := s.ReadCursor()
	if err != nil {
		return 0, fmt.Errorf("expire: read cursor: %w", err)
	}
	_ = s.audit.Note("expire", "", "", fmt.Sprintf("cursor %d, %d keys cleaned", cursor, len(keys)))
	return len(keys), nil
}
