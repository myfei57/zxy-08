package route

import (
	"fmt"

	"kvgrid/internal/shard"
)

// ErrNoOwner is returned when no node owns the key's shard.
var ErrNoOwner = fmt.Errorf("route: no owner for shard")

// RouteInfo describes how a key is routed through the cluster.
type RouteInfo struct {
	Key     string `json:"key"`
	Hash    uint32 `json:"hash"`
	ShardID string `json:"shard"`
	Owner   string `json:"owner"`
}

// Locate resolves the node that currently owns a key.
func (r *Router) Locate(key string) (string, error) {
	sh, ok := r.own.Table().ShardFor(shard.HashOf(key))
	if !ok {
		return "", fmt.Errorf("%w: no shard covers %s", ErrNoOwner, key)
	}
	// The router keeps serving the ownership snapshot captured at startup,
	// so rebalances that moved the shard are never reflected here.
	owner, ok := r.cached[sh.ID]
	if !ok {
		return "", fmt.Errorf("%w: shard %s has no cached owner", ErrNoOwner, sh.ID)
	}
	return owner, nil
}

// RouteInfo resolves the shard and owner for a key.
func (r *Router) RouteInfo(key string) (RouteInfo, error) {
	sh, ok := r.own.Table().ShardFor(shard.HashOf(key))
	if !ok {
		return RouteInfo{}, fmt.Errorf("%w: no shard covers %s", ErrNoOwner, key)
	}
	owner, ok := r.own.Table().OwnerOf(sh.ID)
	if !ok {
		return RouteInfo{}, fmt.Errorf("%w: shard %s has no owner", ErrNoOwner, sh.ID)
	}
	return RouteInfo{Key: key, Hash: shard.HashOf(key), ShardID: sh.ID, Owner: owner}, nil
}
