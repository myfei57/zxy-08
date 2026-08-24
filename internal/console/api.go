package console

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Topology is the cluster layout shown by the topology page.
type Topology struct {
	Nodes  []NodeInfo        `json:"nodes"`
	Shards []ShardInfo       `json:"shards"`
	Owners map[string]string `json:"owners"`
}

// NodeInfo describes one node for the console.
type NodeInfo struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	State   string `json:"state"`
	Keys    int    `json:"keys"`
	Shards  int    `json:"shards"`
}

// ShardInfo describes one shard for the console.
type ShardInfo struct {
	ID    string `json:"id"`
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
	Owner string `json:"owner"`
}

// KeyInfo describes one stored key for the console.
type KeyInfo struct {
	Key     string `json:"key"`
	Bytes   int    `json:"bytes"`
	Version uint64 `json:"version"`
	Expires int64  `json:"expires"`
}

// KeyDetail describes the routing and stored value of one key.
type KeyDetail struct {
	Key   string `json:"key"`
	Hash  uint32 `json:"hash"`
	Shard string `json:"shard"`
	Owner string `json:"owner"`
	Value string `json:"value"`
}

// SnapshotInfo describes one snapshot generation for the console.
type SnapshotInfo struct {
	Generation string `json:"generation"`
	Path       string `json:"path"`
}

// QuotaInfo describes the current capacity usage.
type QuotaInfo struct {
	Capacity int64 `json:"capacity"`
	Used     int64 `json:"used"`
}

type putKeyRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   int64  `json:"ttl"`
}

type deleteKeyRequest struct {
	Key string `json:"key"`
}

type rebalanceRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type splitRequest struct {
	Shard string `json:"shard"`
	Owner string `json:"owner"`
}

type resyncRequest struct {
	Node string `json:"node"`
}

func (s *Server) handleTopology(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.backend.Topology())
}

func (s *Server) handleKeys(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	writeJSON(w, http.StatusOK, s.backend.Keys(prefix))
}

func (s *Server) handleGetKey(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	detail, err := s.backend.KeyDetail(key)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.backend.Snapshots())
}

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	limit := 100
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   s.backend.AuditCount(),
		"by_type": s.backend.AuditCounts(),
		"entries": s.backend.AuditEntries(limit),
	})
}

func (s *Server) handleQuota(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.backend.QuotaStatus())
}

func (s *Server) handlePutKey(w http.ResponseWriter, r *http.Request) {
	var req putKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.Write(req.Key, []byte(req.Value), req.TTL); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stored"})
}

func (s *Server) handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	var req deleteKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.Delete(req.Key); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleTakeSnapshot(w http.ResponseWriter, r *http.Request) {
	if err := s.backend.TakeSnapshot(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "snapshot taken"})
}

func (s *Server) handleRebalance(w http.ResponseWriter, r *http.Request) {
	var req rebalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.Rebalance(req.From, req.To); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "rebalanced"})
}

func (s *Server) handleSplit(w http.ResponseWriter, r *http.Request) {
	var req splitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.SplitShard(req.Shard, req.Owner); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "split"})
}

func (s *Server) handleExpire(w http.ResponseWriter, r *http.Request) {
	n, err := s.backend.RunExpire()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"cleaned": n})
}

func (s *Server) handleEvict(w http.ResponseWriter, r *http.Request) {
	n, err := s.backend.RunEvict()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"evicted": n})
}

func (s *Server) handleResync(w http.ResponseWriter, r *http.Request) {
	var req resyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.Resync(req.Node); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resynced"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
