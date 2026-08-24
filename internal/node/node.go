// Package node models cluster members and their lifecycle states.
package node

// State is the lifecycle state of a cluster node.
type State string

const (
	Active   State = "active"
	Draining State = "draining"
	Down     State = "down"
)

// Node is a single cluster member.
type Node struct {
	ID      string
	Address string
	State   State
}

// New creates an active node with the given identity.
func New(id string, address string) *Node {
	return &Node{ID: id, Address: address, State: Active}
}

// MarkDraining moves an active node into the draining state.
func (n *Node) MarkDraining() error {
	return n.transition(Draining)
}

// MarkDown moves a draining node into the down state.
func (n *Node) MarkDown() error {
	return n.transition(Down)
}

func (n *Node) transition(to State) error {
	if !validTransition(n.State, to) {
		return newTransitionError(n.ID, n.State, to)
	}
	n.State = to
	return nil
}
