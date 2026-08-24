package shard

import (
	"fmt"

	"kvgrid/internal/node"
)

// PlanEntry describes one shard moving from one node to another.
type PlanEntry struct {
	ShardID string
	To      string
}

// Plan is the ordered migration plan produced when a node drains.
type Plan struct {
	Entries []PlanEntry
	From    string
	To      string
}

// BuildPlan computes the shards that move when the source node drains.
func BuildPlan(reg *node.Registry, own *Ownership, from string, to string) (*Plan, error) {
	if _, ok := reg.Get(from); !ok {
		return nil, fmt.Errorf("plan: unknown source node %s", from)
	}
	if _, ok := reg.Get(to); !ok {
		return nil, fmt.Errorf("plan: unknown target node %s", to)
	}
	plan := &Plan{From: from, To: to}
	for _, shard := range own.Table().Shards {
		owner, _ := own.Table().OwnerOf(shard.ID)
		if owner == from {
			plan.Entries = append(plan.Entries, PlanEntry{ShardID: shard.ID, To: to})
		}
	}
	if len(plan.Entries) == 0 {
		return nil, fmt.Errorf("plan: node %s owns no shards", from)
	}
	return plan, nil
}
