package consensus

import (
	"maps"
	"slices"
	"sync"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
)

// minVoters is the minimum number of voters a Raft group must keep.
// Below it, the group can no longer tolerate the loss of a single node without losing its quorum.
const minVoters = 3

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

// removeNode removes nodeID from the members and retires its ID.
func (m *membership) removeNode(nodeID uint64) {
	m.Lock()
	defer m.Unlock()
	delete(m.members, nodeID)
	m.removedIDs = append(m.removedIDs, nodeID)
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

// validateConfChange checks that applying change keeps the membership valid.
// NB: This method must remain deterministic as it is called by every node when the change is applied.
// It uses members and not the ConfState, which can be more recent than the members after restoring
// a snapshot received from the leader (see Consensus.updateSnapshotConfState).
// A rejected change is reported as a dsserr.BadRequest error.
func (m *membership) validateConfChange(change raftpb.ConfChangeSingle, peerURL string) error {
	m.RLock()
	defer m.RUnlock()

	if change.NodeID == raft.None {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node ID %d is invalid", change.NodeID)
	}

	if slices.Contains(m.removedIDs, change.NodeID) {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node ID %d was removed from the group and cannot be reused", change.NodeID)
	}

	_, isMember := m.members[change.NodeID]
	switch change.Type {
	case raftpb.ConfChangeAddNode:
		if isMember {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node %d is already a member of the group", change.NodeID)
		}
		if _, err := types.NewURLs([]string{peerURL}); err != nil {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "invalid URL %q for node %d: %v", peerURL, change.NodeID, err)
		}

	case raftpb.ConfChangeRemoveNode:
		if !isMember {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node %d is not a member of the group", change.NodeID)
		}
		if len(m.members)-1 < minVoters {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "removing node %d would leave %d voters, below the minimum of %d", change.NodeID, len(m.members)-1, minVoters)
		}

	default:
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "unsupported config change type %s", change.Type)
	}

	return nil
}
