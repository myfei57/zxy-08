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
	next := own.Table().Clone()
	next.Replace(shardID, left, right, newOwner)
	if err := s.audit.Record(audit.Event{Type: "split-start", Key: shardID, Detail: newOwner}); err != nil {
		return err
	}
	if err := own.SaveTable(next); err != nil {
		return err
	}
	return s.audit.Record(audit.Event{Type: "split-done", Key: shardID, Detail: left.ID + "," + right.ID})
}

// splitChildren cuts a shard into two children at the given hash point.
func splitChildren(src Shard, point uint32) (Shard, Shard) {
	left := Shard{ID: src.ID + "-l", Start: src.Start, End: point}
	right := Shard{ID: src.ID + "-r", Start: point, End: src.End}
	return left, right
}
