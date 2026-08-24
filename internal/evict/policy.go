package evict

import "sort"

// Pick selects up to limit keys for eviction, oldest writes first.
func (e *Evictor) Pick(limit int) []string {
	keys := e.st.Keys("")
	sort.Slice(keys, func(i, j int) bool {
		a, aok := e.st.Entry(keys[i])
		b, bok := e.st.Entry(keys[j])
		if !aok {
			return false
		}
		if !bok {
			return true
		}
		return a.Seq < b.Seq
	})
	if len(keys) > limit {
		keys = keys[:limit]
	}
	return keys
}
