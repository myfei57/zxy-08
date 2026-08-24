// Package console exposes the KVGrid HTTP API and embedded control pages.
package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"kvgrid/internal/audit"
	"kvgrid/internal/web"
)

// Backend is the cluster surface the console renders.
type Backend interface {
	Write(key string, value []byte, ttl int64) error
	Read(key string) ([]byte, bool)
	Delete(key string) error
	Topology() Topology
	Keys(prefix string) []KeyInfo
	KeyDetail(key string) (KeyDetail, error)
	Snapshots() []SnapshotInfo
	AuditEntries(limit int) []audit.Event
	AuditCount() int
	AuditCounts() map[string]int
	QuotaStatus() QuotaInfo
	TakeSnapshot() error
	Restore() error
	Rebalance(from string, to string) error
	SplitShard(shardID string, owner string) error
	RunExpire() (int, error)
	RunEvict() (int, error)
	Resync(nodeID string) error
}

// Server serves the console pages and JSON API.
type Server struct {
	backend Backend
}

// NewServer creates a console server backed by a cluster.
func NewServer(backend Backend) *Server {
	return &Server{backend: backend}
}

// Handler builds the chi router.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", s.handleIndex)
	r.Get("/topology", s.handlePage(web.TopologyHTML))
	r.Get("/keys", s.handlePage(web.KeysHTML))
	r.Get("/snapshots", s.handlePage(web.SnapshotsHTML))
	r.Get("/audit", s.handlePage(web.AuditHTML))
	r.Get("/api/health", s.handleHealth)
	r.Get("/api/topology", s.handleTopology)
	r.Get("/api/keys", s.handleKeys)
	r.Get("/api/keys/{key}", s.handleGetKey)
	r.Get("/api/snapshots", s.handleSnapshots)
	r.Get("/api/audit", s.handleAudit)
	r.Get("/api/quota", s.handleQuota)
	r.Post("/api/keys", s.handlePutKey)
	r.Delete("/api/keys", s.handleDeleteKey)
	r.Post("/api/snapshots", s.handleTakeSnapshot)
	r.Post("/api/rebalance", s.handleRebalance)
	r.Post("/api/split", s.handleSplit)
	r.Post("/api/expire", s.handleExpire)
	r.Post("/api/evict", s.handleEvict)
	r.Post("/api/resync", s.handleResync)
	return r
}

// Start serves the console until the process exits.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
