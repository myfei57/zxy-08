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

func TestRebalanceKeepsInflightWrites(t *testing.T) {
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
	if err := own.SaveTable(shard.BuildInitial(2, []string{"node-a"})); err != nil {
		t.Fatal(err)
	}
	auditLogger := audit.NewLogger(filepath.Join(dir, "audit.log"))
	migrator := shard.NewMigrator(reg, own, stores, auditLogger)
	plan, err := shard.BuildPlan(reg, own, "node-a", "node-b")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.StartDrain(plan); err != nil {
		t.Fatal(err)
	}
	router := route.NewRouter(own, stores, quota.NewManager(1<<20, ""), reg, migrator, auditLogger)
	key := keyInShard(t, own.Table(), "shard-0")
	if err := router.Set(key, []byte("inflight"), 0); err != nil {
		t.Fatal(err)
	}
	if err := migrator.Finish(); err != nil {
		t.Fatal(err)
	}
	if !stores["node-b"].Has(key) {
		t.Fatal("in-flight write must be replayed onto the new owner after migration")
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
