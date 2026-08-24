package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"kvgrid/internal/audit"
	"kvgrid/internal/snapshot"
	"kvgrid/internal/store"
)

func TestRestoreRejectsMixedGenerations(t *testing.T) {
	dir := t.TempDir()
	snapDir := filepath.Join(dir, "snapshots")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	st1, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "s1")})
	if err != nil {
		t.Fatal(err)
	}
	defer st1.Close()
	if _, err := st1.Set("k", []byte("v1"), 0); err != nil {
		t.Fatal(err)
	}
	if err := st1.Dump(filepath.Join(snapDir, "gen-1", "snapshot.json")); err != nil {
		t.Fatal(err)
	}
	st2, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "s2")})
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	if _, err := st2.Set("k", []byte("v2"), 0); err != nil {
		t.Fatal(err)
	}
	if err := st2.Dump(filepath.Join(snapDir, "gen-2", "snapshot.json")); err != nil {
		t.Fatal(err)
	}
	target, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "target")})
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	snap := snapshot.NewManager(target, snapDir, filepath.Join(dir, "cursor.json"), audit.NewLogger(filepath.Join(dir, "audit.log")))
	if err := snap.WriteCursor(2); err != nil {
		t.Fatal(err)
	}
	if err := snap.Restore(snapDir); err != nil {
		t.Fatal(err)
	}
	got, ok := target.Get("k")
	if !ok {
		t.Fatal("key missing after restore")
	}
	if string(got) != "v2" {
		t.Fatalf("restore must load only the current generation: got %q", got)
	}
}
