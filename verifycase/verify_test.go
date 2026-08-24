package verifycase

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"kvgrid/internal/audit"
	"kvgrid/internal/node"
	"kvgrid/internal/quota"
	"kvgrid/internal/route"
	"kvgrid/internal/shard"
	"kvgrid/internal/store"
)

func TestSplitExposesSingleOwnerPerLookup(t *testing.T) {
	dir := t.TempDir()
	reg := node.NewRegistry()
	if err := reg.Register(node.New("node-a", "a:1")); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(node.New("node-b", "b:2")); err != nil {
		t.Fatal(err)
	}
	stores := map[string]*store.Store{}
	for _, id := range []string{"node-a", "node-b"} {
		st, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, id)})
		if err != nil {
			t.Fatal(err)
		}
		defer st.Close()
		stores[id] = st
	}
	own := shard.NewOwnership(filepath.Join(dir, "ownership.json"))
	table := shard.BuildInitial(2, []string{"node-a"})
	if err := own.SaveTable(table); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "audit-blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	blockedAudit := audit.NewLogger(filepath.Join(blocker, "audit.log"))
	auditLogger := audit.NewLogger(filepath.Join(dir, "audit.log"))
	migrator := shard.NewMigrator(reg, own, stores, auditLogger)
	router := route.NewRouter(own, stores, quota.NewManager(1<<20, ""), reg, migrator, auditLogger)
	splitter := shard.NewSplitter(blockedAudit)
	src := table.Shards[0]
	mid := src.Start + (src.End-src.Start)/2
	leftKey := keyWithHashBetween(t, src.Start, mid)
	rightKey := keyWithHashBetween(t, mid, src.End)
	if err := splitter.Split(own, src.ID, mid, "node-b"); err == nil {
		t.Fatal("split must fail when the progress audit cannot be recorded")
	}
	ownerLeft, errL := router.Locate(leftKey)
	ownerRight, errR := router.Locate(rightKey)
	if errL != nil || errR != nil {
		t.Fatalf("lookups after the failed split failed: %v %v", errL, errR)
	}
	if ownerLeft != ownerRight {
		t.Fatalf("split must expose one consistent owner per lookup: left=%s right=%s", ownerLeft, ownerRight)
	}
}

func keyWithHashBetween(t *testing.T, lo, hi uint32) string {
	t.Helper()
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("split-key-%d", i)
		h := shard.HashOf(key)
		if h >= lo && h < hi {
			return key
		}
	}
	t.Fatalf("no key found in hash range %d..%d", lo, hi)
	return ""
}
