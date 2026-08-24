package verifycase

import (
	"path/filepath"
	"testing"

	"kvgrid/internal/replica"
	"kvgrid/internal/store"
)

func TestReplicaAckAfterDurableWrite(t *testing.T) {
	dir := t.TempDir()
	st, err := store.NewStore(store.Options{
		DataDir: filepath.Join(dir, "data"),
		MetaDir: filepath.Join(dir, "missing-meta"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	op, err := st.Set("k", []byte("v"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := replica.Ack(st, op); err == nil {
		t.Fatal("replicated ack must fail when the write is not durable")
	}
}
