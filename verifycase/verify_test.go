package verifycase

import (
	"path/filepath"
	"testing"

	"kvgrid/internal/audit"
	"kvgrid/internal/expire"
	"kvgrid/internal/store"
)

func TestExpireCursorAdvancesAfterDeleteDurable(t *testing.T) {
	dir := t.TempDir()
	st, err := store.NewStore(store.Options{
		DataDir:        filepath.Join(dir, "data"),
		DeleteMetaPath: filepath.Join(dir, "missing-meta", "delete.meta"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.Set("expired", []byte("v"), 100); err != nil {
		t.Fatal(err)
	}
	cursorPath := filepath.Join(dir, "expire-cursor.json")
	scanner := expire.NewScanner(st, cursorPath, 10, audit.NewLogger(filepath.Join(dir, "audit.log")))
	if _, err := scanner.Scan(1000); err == nil {
		t.Fatal("expiry scan must fail when the deletes are not durable")
	}
	cur, err := scanner.ReadCursor()
	if err != nil {
		t.Fatal(err)
	}
	if cur != 0 {
		t.Fatalf("cursor advanced to %d although the delete batch was not durable", cur)
	}
}
