package store

import "sort"

// Get returns the value stored for a key that has not expired.
func (s *Store) Get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.index[key]
	if !ok {
		return nil, false
	}
	if v.Expires != 0 && v.Expires <= nowNanos() {
		return nil, false
	}
	return append([]byte(nil), v.Value...), true
}

// Has reports whether a key currently exists in the index.
func (s *Store) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.index[key]
	return ok
}

// Entry returns the full stored value for a key.
func (s *Store) Entry(key string) (Value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.index[key]
	if !ok {
		return Value{}, false
	}
	return *v, true
}

// Keys lists keys with the given prefix, sorted.
func (s *Store) Keys(prefix string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var keys []string
	for key := range s.index {
		if len(prefix) == 0 || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// ExpiredKeys returns keys whose expiry has passed by the given time.
func (s *Store) ExpiredKeys(now int64) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var keys []string
	for key, v := range s.index {
		if v.Expires != 0 && v.Expires <= now {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// LastSeq returns the highest operation sequence seen by the store.
func (s *Store) LastSeq() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seq
}

// Count returns the number of keys in the index.
func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.index)
}

// JournalPath returns the path of the write-ahead journal.
func (s *Store) JournalPath() string {
	return s.journal.Name()
}
