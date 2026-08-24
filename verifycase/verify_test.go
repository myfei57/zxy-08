package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"kvgrid/internal/audit"
	"kvgrid/internal/evict"
	"kvgrid/internal/quota"
	"kvgrid/internal/store"
)

func TestEvictReleasesQuotaBeforeKeyRemoval(t *testing.T) {
	dir := t.TempDir()
	st, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "data")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.Set("k", []byte("payload"), 0); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	quotaManager := quota.NewManager(1024, filepath.Join(blocker, "ledger.log"))
	if err := quotaManager.Account("k", 7); err != nil {
		t.Fatal(err)
	}
	evictor := evict.NewEvictor(st, quotaManager, audit.NewLogger(filepath.Join(dir, "audit.log")))
	if err := evictor.Run("k"); err == nil {
		t.Fatal("eviction must fail when the quota release is not durable")
	}
	if !st.Has("k") {
		t.Fatal("key must remain in the index when the quota release fails")
	}
}
