package consensus

import (
	"slices"
	"sync"

	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
)

type membership struct {
	sync.RWMutex

	// nodeID is the ID of this node, which is never added as a peer of the transport.
	nodeID uint64
	// transport is kept in sync with the members of the group.
	transport *rafthttp.Transport

	// members maps the ID of every member of the group to its description.
	members map[uint64]member
	// removedIDs are the IDs of the nodes removed from the group
	// removed IDs can never be reused.
	removedIDs []uint64
}

func (m *membership) isIDRemoved(id uint64) bool {
	m.RLock()
	defer m.RUnlock()
	return slices.Contains(m.removedIDs, id)
}

// addNode records nodeID as a member described by mem.
func (m *membership) addNode(nodeID uint64, mem member) {
	m.Lock()
	defer m.Unlock()
	m.members[nodeID] = mem
}

// restore replaces the members and the removed IDs and updates the transport peers accordingly.
func (m *membership) restore(members map[uint64]member, removedIDs []uint64) error {
	for _, id := range removedIDs {
		if _, ok := members[id]; ok {
			return stacktrace.NewError("removed ID %d is still present in the members map", id)
		}
	}
	if members == nil {
		members = map[uint64]member{}
	}

	m.Lock()
	m.members = members
	m.removedIDs = removedIDs
	m.Unlock()

	for _, id := range removedIDs {
		m.transport.RemovePeer(types.ID(id))
	}
	for id, mem := range members {
		if id != m.nodeID && mem.URL != "" {
			m.transport.AddPeer(types.ID(id), []string{mem.URL})
		}
	}
	return nil
}
