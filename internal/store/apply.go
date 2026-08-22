package store

import "fmt"

// Apply applies a remote operation to this store and records the applied offset.
func (s *Store) Apply(op Op) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if op.Kind == "delete" {
		delete(s.index, op.Key)
	} else {
		version := op.Version
		if version == 0 {
			version = uint64(1)
			if prev, ok := s.index[op.Key]; ok {
				version = prev.Version + 1
			}
		}
		s.index[op.Key] = &Value{
			Key:     op.Key,
			Value:   append([]byte(nil), op.Value...),
			Version: version,
			Expires: op.Expires,
			Seq:     op.Seq,
		}
	}
	if op.Seq > s.applied {
		s.applied = op.Seq
	}
	if err := writeMetaAtomic(s.metaPath("applied.meta"), metaSeq{Seq: s.applied}); err != nil {
		return fmt.Errorf("store: write applied meta: %w", err)
	}
	return nil
}

// AppliedOffset returns the highest operation seq applied to this store.
func (s *Store) AppliedOffset() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applied, nil
}
