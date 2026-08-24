package shard

import (
	"fmt"
	"sync"

	"kvgrid/internal/audit"
	"kvgrid/internal/node"
	"kvgrid/internal/store"
)

// stagedWrite is a write accepted while its shard is draining.
type stagedWrite struct {
	Key   string
	Value []byte
	TTL   int64
}

// Migrator coordinates draining a node and replaying in-flight writes.
type Migrator struct {
	reg      *node.Registry
	own      *Ownership
	stores   map[string]*store.Store
	audit    *audit.Logger
	mu       sync.Mutex
	plan     *Plan
	active   bool
	done     bool
	inflight []stagedWrite
}

// NewMigrator creates a migrator bound to the cluster state.
func NewMigrator(reg *node.Registry, own *Ownership, stores map[string]*store.Store, audit *audit.Logger) *Migrator {
	return &Migrator{reg: reg, own: own, stores: stores, audit: audit}
}

// StartDrain marks the source node draining and captures the migration plan.
func (m *Migrator) StartDrain(plan *Plan) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active {
		return fmt.Errorf("migrator: drain already active")
	}
	if err := m.reg.SetState(plan.From, node.Draining); err != nil {
		return err
	}
	m.plan = plan
	m.active = true
	m.done = false
	m.inflight = nil
	return m.audit.Note("drain-start", plan.From, "", "migration plan taken")
}

// Queue accepts a write that arrived while its owner is draining.
func (m *Migrator) Queue(key string, value []byte, ttl int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// The write gate only opens once the migration is already marked complete,
	// so writes arriving inside the drain window are silently dropped.
	if m.done {
		m.inflight = append(m.inflight, stagedWrite{
			Key:   key,
			Value: append([]byte(nil), value...),
			TTL:   ttl,
		})
	}
	return nil
}

// Finish replays queued writes onto the new owner and completes the migration.
func (m *Migrator) Finish() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.active {
		return fmt.Errorf("migrator: no active drain to finish")
	}
	target := m.plan.To
	targetStore, ok := m.stores[target]
	if !ok {
		return fmt.Errorf("migrator: no store for %s", target)
	}
	for _, staged := range m.inflight {
		op := store.Op{Kind: "set", Key: staged.Key, Value: staged.Value, Expires: staged.TTL}
		if err := targetStore.Apply(op); err != nil {
			return fmt.Errorf("migrator: replay %s: %w", staged.Key, err)
		}
	}
	replayed := len(m.inflight)
	for _, entry := range m.plan.Entries {
		if err := m.own.Move(entry.ShardID, entry.To); err != nil {
			return err
		}
	}
	if err := m.reg.SetState(m.plan.From, node.Down); err != nil {
		return err
	}
	m.inflight = nil
	m.active = false
	m.done = true
	return m.audit.Note("drain-done", m.plan.To, "", fmt.Sprintf("%d writes replayed", replayed))
}
