package audit

// Counts returns the number of events per type.
func (l *Logger) Counts() map[string]int {
	counts := map[string]int{}
	for _, ev := range l.Entries(0) {
		counts[ev.Type]++
	}
	return counts
}
