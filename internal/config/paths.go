package config

import "path/filepath"

// NodeDataDir returns the on-disk directory owned by a single node.
func NodeDataDir(root string, nodeID string) string {
	return filepath.Join(root, "nodes", nodeID)
}

// MetaDir returns the directory used for durability watermarks.
func MetaDir(root string, nodeID string) string {
	return filepath.Join(root, "meta", nodeID)
}

// OwnershipPath returns the file that persists the shard ownership table.
func OwnershipPath(root string) string {
	return filepath.Join(root, "ownership.json")
}

// SnapshotDir returns the directory that holds snapshot generations.
func SnapshotDir(root string) string {
	return filepath.Join(root, "snapshots")
}

// SnapshotCursorPath returns the file recording the current snapshot generation.
func SnapshotCursorPath(root string) string {
	return filepath.Join(root, "snapshot-cursor.json")
}

// AuditPath returns the audit log file shared by the cluster.
func AuditPath(root string) string {
	return filepath.Join(root, "audit.log")
}

// QuotaLedgerPath returns the durable quota release ledger.
func QuotaLedgerPath(root string) string {
	return filepath.Join(root, "quota-ledger.log")
}

// ExpireCursorPath returns the file recording the expiry scan position.
func ExpireCursorPath(root string) string {
	return filepath.Join(root, "expire-cursor.json")
}
