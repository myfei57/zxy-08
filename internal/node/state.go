package node

import "fmt"

// validTransition reports whether moving from one node state to another is
// allowed by the lifecycle machine.
func validTransition(from State, to State) bool {
	switch from {
	case Active:
		return to == Draining || to == Down
	case Draining:
		return to == Down || to == Active
	case Down:
		return to == Active
	default:
		return false
	}
}

func newTransitionError(id string, from State, to State) error {
	return fmt.Errorf("node %s cannot transition from %s to %s", id, from, to)
}
