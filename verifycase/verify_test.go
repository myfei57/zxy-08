package verifycase

import (
	"path/filepath"
	"testing"

	"kvgrid/internal/replica"
	"kvgrid/internal/store"
)

func TestResyncResumesFromAppliedOffset(t *testing.T) {
	dir := t.TempDir()
	primary, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "primary")})
	if err != nil {
		t.Fatal(err)
	}
	defer primary.Close()
	op1, err := primary.Set("k", []byte("a"), 0)
	if err != nil {
		t.Fatal(err)
	}
	op2, err := primary.Set("k", []byte("b"), 0)
	if err != nil {
		t.Fatal(err)
	}
	replicaStore, err := store.NewStore(store.Options{DataDir: filepath.Join(dir, "replica")})
	if err != nil {
		t.Fatal(err)
	}
	defer replicaStore.Close()
	if err := replicaStore.Apply(op1); err != nil {
		t.Fatal(err)
	}
	if err := replicaStore.Apply(op2); err != nil {
		t.Fatal(err)
	}
	if _, err := primary.Set("k", []byte("c"), 0); err != nil {
		t.Fatal(err)
	}
	manager := replica.NewManager(replicaStore)
	if err := manager.Resync(primary.JournalPath()); err != nil {
		t.Fatal(err)
	}
	got, ok := replicaStore.Get("k")
	if !ok {
		t.Fatal("key missing after resync")
	}
	if string(got) != "c" {
		t.Fatalf("resync must not overwrite newer values with stale ones: got %q", got)
	}
}
