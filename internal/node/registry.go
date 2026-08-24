package node

import "fmt"

// Registry tracks the members of a cluster.
type Registry struct {
	nodes map[string]*Node
}

// NewRegistry returns an empty node registry.
func NewRegistry() *Registry {
	return &Registry{nodes: make(map[string]*Node)}
}

// Register adds a node to the registry.
func (r *Registry) Register(n *Node) error {
	if n == nil || n.ID == "" {
		return fmt.Errorf("node registry: cannot register an empty node")
	}
	if _, exists := r.nodes[n.ID]; exists {
		return fmt.Errorf("node registry: node %s already registered", n.ID)
	}
	r.nodes[n.ID] = n
	return nil
}

// Get returns the node with the given id.
func (r *Registry) Get(id string) (*Node, bool) {
	n, ok := r.nodes[id]
	return n, ok
}

// SetState validates and applies a state transition.
func (r *Registry) SetState(id string, state State) error {
	n, ok := r.nodes[id]
	if !ok {
		return fmt.Errorf("node registry: unknown node %s", id)
	}
	if state == Active {
		n.State = Active
		return nil
	}
	if state == Draining {
		return n.MarkDraining()
	}
	if state == Down {
		return n.MarkDown()
	}
	return fmt.Errorf("node registry: invalid state %q", state)
}

// State returns the current state of a node.
func (r *Registry) State(id string) (State, bool) {
	n, ok := r.nodes[id]
	if !ok {
		return "", false
	}
	return n.State, true
}

// ActiveIDs lists the ids of nodes that are not down.
func (r *Registry) ActiveIDs() []string {
	var ids []string
	for _, n := range r.nodes {
		if n.State != Down {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

// FirstActive returns the first non-down node id.
func (r *Registry) FirstActive() (string, bool) {
	for _, n := range r.nodes {
		if n.State != Down {
			return n.ID, true
		}
	}
	return "", false
}
