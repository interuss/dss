package admin

import (
	"context"
	"maps"
	"sync"

	"go.etcd.io/raft/v3/raftpb"
)

// MembershipManager is to be implemented by any store that supports raft membership changes.
type MembershipManager interface {
	ProposeConfChange(ctx context.Context, changeType raftpb.ConfChangeType, nodeID uint64, peerURL string) (*raftpb.ConfState, error)
}

type storeRegistry struct {
	mu     sync.Mutex
	stores map[string]MembershipManager
}

func (r *storeRegistry) register(name string, store MembershipManager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stores[name] = store
}

func (r *storeRegistry) clone() map[string]MembershipManager {
	r.mu.Lock()
	defer r.mu.Unlock()

	return maps.Clone(r.stores)
}
