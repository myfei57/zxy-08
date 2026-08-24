package verifycase

import (
	"fmt"
	"path/filepath"
	"testing"

	"kvgrid/internal/audit"
	"kvgrid/internal/node"
	"kvgrid/internal/quota"
	"kvgrid/internal/route"
	"kvgrid/internal/shard"
	"kvgrid/internal/store"
)

func TestRouteUsesCurrentOwnershipAfterRebalance(t *testing.T) {
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
	if err := own.SaveTable(shard.BuildInitial(4, []string{"node-a"})); err != nil {
		t.Fatal(err)
	}
	auditLogger := audit.NewLogger(filepath.Join(dir, "audit.log"))
	migrator := shard.NewMigrator(reg, own, stores, auditLogger)
	router := route.NewRouter(own, stores, quota.NewManager(1<<20, ""), reg, migrator, auditLogger)
	if err := own.Move("shard-1", "node-b"); err != nil {
		t.Fatal(err)
	}
	key := keyInShard(t, own.Table(), "shard-1")
	if err := router.Set(key, []byte("v"), 0); err != nil {
		t.Fatal(err)
	}
	if !stores["node-b"].Has(key) {
		t.Fatal("write must land on the current owner after rebalance")
	}
	if stores["node-a"].Has(key) {
		t.Fatal("key must not remain on the stale owner")
	}
}

func keyInShard(t *testing.T, table *shard.Table, shardID string) string {
	t.Helper()
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("route-key-%d", i)
		sh, ok := table.ShardFor(shard.HashOf(key))
		if ok && sh.ID == shardID {
			return key
		}
	}
	t.Fatalf("no key found for shard %s", shardID)
	return ""
}
