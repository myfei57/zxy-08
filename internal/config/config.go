// Package config defines the runtime configuration of a KVGrid process.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the settings needed to start one KVGrid cluster process.
type Config struct {
	HTTPAddr   string
	DataDir    string
	NodeID     string
	ShardCount int
	QuotaBytes int64
}

// Default returns a development-friendly configuration.
func Default() Config {
	return Config{
		HTTPAddr:   "127.0.0.1:7788",
		DataDir:    filepath.Join(os.TempDir(), "kvgrid"),
		NodeID:     "node-a",
		ShardCount: 8,
		QuotaBytes: 1 << 20,
	}
}

// Validate checks that the configuration can produce a runnable cluster.
func (c Config) Validate() error {
	if c.DataDir == "" {
		return fmt.Errorf("kvgrid: data dir is required")
	}
	if c.NodeID == "" {
		return fmt.Errorf("kvgrid: node id is required")
	}
	if c.ShardCount < 1 {
		return fmt.Errorf("kvgrid: shard count must be positive")
	}
	if c.QuotaBytes < 0 {
		return fmt.Errorf("kvgrid: quota bytes cannot be negative")
	}
	return nil
}
