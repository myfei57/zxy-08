package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"kvgrid/internal/audit"
	"kvgrid/internal/snapshot"
	"kvgrid/internal/store"
)

func TestSnapshotCursorAfterDataDurable(t *testing.T) {
	dir := t.TempDir()
	st, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "data")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.Set("k", []byte("v"), 0); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "snapshots")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	snap := snapshot.NewManager(st, blocker, filepath.Join(dir, "cursor.json"), audit.NewLogger(filepath.Join(dir, "audit.log")))
	if err := snap.Take(); err == nil {
		t.Fatal("snapshot must fail when the snapshot data cannot be dumped")
	}
	gen, err := snap.CurrentGeneration()
	if err != nil {
		t.Fatal(err)
	}
	if gen != 0 {
		t.Fatalf("cursor advanced to generation %d although the snapshot data was not durable", gen)
	}
}
