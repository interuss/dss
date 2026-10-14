package consensus

import (
	"maps"
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

	// members maps the ID of every member of the group to its peer URL.
	members map[uint64]string
	// removedIDs are the IDs of the nodes removed from the group
	// removed IDs can never be reused.
	removedIDs []uint64
}

func (m *membership) isIDRemoved(id uint64) bool {
	m.RLock()
	defer m.RUnlock()
	return slices.Contains(m.removedIDs, id)
}

// addNode records nodeID as a member reachable at peerURL.
func (m *membership) addNode(nodeID uint64, peerURL string) {
	m.Lock()
	defer m.Unlock()
	m.members[nodeID] = peerURL
}

func (m *membership) clone() membership {
	m.RLock()
	defer m.RUnlock()
	return membership{
		members:    maps.Clone(m.members),
		removedIDs: slices.Clone(m.removedIDs),
	}
}

// restore replaces the members and the removed IDs and updates the transport peers accordingly.
func (m *membership) restore(members map[uint64]string, removedIDs []uint64) error {
	m.Lock()
	m.members = members
	if m.members == nil {
		m.members = map[uint64]string{}
	}
	m.removedIDs = removedIDs
	m.Unlock()

	for _, id := range removedIDs {
		_, ok := m.members[id]
		if ok {
			return stacktrace.NewError("removed ID %d is still present in the members map", id)
		}
		m.transport.RemovePeer(types.ID(id))
	}
	for id, peerURL := range m.members {
		if id != m.nodeID && peerURL != "" {
			m.transport.AddPeer(types.ID(id), []string{peerURL})
		}
	}
	return nil
}
