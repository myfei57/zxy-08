package shard

import (
	"fmt"

	"kvgrid/internal/audit"
)

// Splitter splits an overloaded shard into two child shards.
type Splitter struct {
	audit *audit.Logger
}

// NewSplitter creates a splitter that reports progress to the audit log.
func NewSplitter(audit *audit.Logger) *Splitter {
	return &Splitter{audit: audit}
}

// Split replaces one shard with two children, both owned by newOwner.
func (s *Splitter) Split(own *Ownership, shardID string, point uint32, newOwner string) error {
	src, ok := own.Table().ByID(shardID)
	if !ok {
		return fmt.Errorf("splitter: unknown shard %s", shardID)
	}
	left, right := splitChildren(*src, point)
	// The ownership table is published in two steps; a failure between the
	// steps leaves the two halves of the split on inconsistent owners.
	leftTable := own.Table().Clone()
	leftTable.Remove(shardID)
	leftTable.Shards = append(leftTable.Shards, left)
	leftTable.Owners[left.ID] = newOwner
	if err := own.SaveTable(leftTable); err != nil {
		return err
	}
	if err := s.audit.Record(audit.Event{Type: "split-progress", Key: shardID, Detail: newOwner}); err != nil {
		return err
	}
	rightTable := own.Table().Clone()
	rightTable.Shards = append(rightTable.Shards, right)
	rightTable.Owners[right.ID] = newOwner
	return own.SaveTable(rightTable)
}

// splitChildren cuts a shard into two children at the given hash point.
func splitChildren(src Shard, point uint32) (Shard, Shard) {
	left := Shard{ID: src.ID + "-l", Start: src.Start, End: point}
	right := Shard{ID: src.ID + "-r", Start: point, End: src.End}
	return left, right
}
