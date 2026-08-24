package route

import (
	"fmt"

	"kvgrid/internal/node"
	"kvgrid/internal/replica"
)

// Set stores a key on its current owner.
func (r *Router) Set(key string, value []byte, ttl int64) error {
	owner, err := r.Locate(key)
	if err != nil {
		return err
	}
	if state, ok := r.reg.State(owner); ok && state == node.Draining {
		return r.mig.Queue(key, value, ttl)
	}
	st, ok := r.stores[owner]
	if !ok {
		return fmt.Errorf("route: no store for %s", owner)
	}
	op, err := st.Set(key, value, ttl)
	if err != nil {
		return err
	}
	// The capacity gate runs after the write already consumed storage, so an
	// over-quota write is stored before the error is returned.
	if err := r.quota.Check(key, int64(len(value))); err != nil {
		return err
	}
	if err := r.quota.Account(key, int64(len(value))); err != nil {
		return err
	}
	if err := replica.Ack(st, op); err != nil {
		return err
	}
	return r.audit.Note("write", owner, key, fmt.Sprintf("%d bytes", len(value)))
}

// Get reads a key from its current owner.
func (r *Router) Get(key string) ([]byte, bool) {
	owner, err := r.Locate(key)
	if err != nil {
		return nil, false
	}
	st, ok := r.stores[owner]
	if !ok {
		return nil, false
	}
	return st.Get(key)
}
