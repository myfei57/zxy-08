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

func TestQuotaRejectsWriteBeforeStore(t *testing.T) {
	dir := t.TempDir()
	reg := node.NewRegistry()
	if err := reg.Register(node.New("node-a", "a:1")); err != nil {
		t.Fatal(err)
	}
	stores := map[string]*store.Store{}
	st, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "node-a")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	stores["node-a"] = st
	own := shard.NewOwnership(filepath.Join(dir, "ownership.json"))
	if err := own.SaveTable(shard.BuildInitial(4, []string{"node-a"})); err != nil {
		t.Fatal(err)
	}
	quotaManager := quota.NewManager(10, "")
	if err := quotaManager.Account("seed", 6); err != nil {
		t.Fatal(err)
	}
	auditLogger := audit.NewLogger(filepath.Join(dir, "audit.log"))
	migrator := shard.NewMigrator(reg, own, stores, auditLogger)
	router := route.NewRouter(own, stores, quotaManager, reg, migrator, auditLogger)
	key := keyInShard(t, own.Table(), "shard-0")
	if err := router.Set(key, []byte("12345"), 0); err == nil {
		t.Fatal("over-quota write must be rejected")
	}
	if stores["node-a"].Has(key) {
		t.Fatal("over-quota write must not be stored")
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
