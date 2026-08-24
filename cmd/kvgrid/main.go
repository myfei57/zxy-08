// Command kvgrid starts a KVGrid cluster process with an embedded console.
package main

import (
	"flag"
	"fmt"
	"log"

	"kvgrid/internal/cluster"
	"kvgrid/internal/config"
)

func main() {
	var (
		addr    = flag.String("addr", "", "HTTP listen address")
		data    = flag.String("data", "", "data directory")
		shards  = flag.Int("shards", 0, "number of hash shards")
		quota   = flag.Int64("quota", 0, "capacity quota in bytes")
		restore = flag.Bool("restore", false, "restore from the newest snapshot before serving")
	)
	flag.Parse()
	cfg := config.Default()
	if *addr != "" {
		cfg.HTTPAddr = *addr
	}
	if *data != "" {
		cfg.DataDir = *data
	}
	if *shards > 0 {
		cfg.ShardCount = *shards
	}
	if *quota > 0 {
		cfg.QuotaBytes = *quota
	}
	cl, err := cluster.Build(cfg)
	if err != nil {
		log.Fatalf("kvgrid: build: %v", err)
	}
	defer func() {
		if err := cl.Close(); err != nil {
			log.Printf("kvgrid: close: %v", err)
		}
	}()
	if *restore {
		if err := cl.Recover(); err != nil {
			log.Fatalf("kvgrid: recover: %v", err)
		}
	}
	fmt.Printf("kvgrid listening on %s (data %s)\n", cfg.HTTPAddr, cfg.DataDir)
	if err := cl.Start(); err != nil {
		log.Fatal(err)
	}
}
