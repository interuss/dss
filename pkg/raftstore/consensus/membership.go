package consensus

import (
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

// addNode records nodeID as a member described by mem and adds it as a peer of the transport.
func (m *membership) addNode(nodeID uint64, mem member) {
	m.Lock()
	m.members[nodeID] = mem
	m.Unlock()

	if nodeID != m.nodeID {
		m.transport.AddPeer(types.ID(nodeID), []string{mem.URL})
	}
}

// removeNode removes nodeID from the members, retires its ID and removes it from the peers of the transport.
func (m *membership) removeNode(nodeID uint64) {
	m.Lock()
	delete(m.members, nodeID)
	m.removedIDs = append(m.removedIDs, nodeID)
	m.Unlock()

	if nodeID != m.nodeID {
		m.transport.RemovePeer(types.ID(nodeID))
	}
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

	// It is safe to do this outside the lock because the transport methods as well as addNode and removeNode
	// are only called by the consumer goroutine. So all transport updates happen one at a time. The lock only protects
	// the readers on other goroutine (isIDRemoved, validateConfChange etc.) who never call the transport methods.
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

// validateConfChange checks that applying change keeps the membership valid.
// NB: This method must remain deterministic as it is called by every node when the change is applied.
// It uses members and not the ConfState, which can be more recent than the members after restoring
// a snapshot received from the leader (see Consensus.updateSnapshotConfState).
// A rejected change is reported as a dsserr.BadRequest error.
func (m *membership) validateConfChange(change *raftpb.ConfChangeSingle, peerURL string) error {
	m.RLock()
	defer m.RUnlock()

	if change.GetNodeId() == raft.None {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node ID %d is invalid", change.GetNodeId())
	}

	if slices.Contains(m.removedIDs, change.GetNodeId()) {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node ID %d was removed from the group and cannot be reused", change.GetNodeId())
	}

	_, isMember := m.members[change.GetNodeId()]
	switch change.GetType() {
	case raftpb.ConfChangeAddNode:
		if isMember {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node %d is already a member of the group", change.GetNodeId())
		}
		if _, err := types.NewURLs([]string{peerURL}); err != nil {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "invalid URL %q for node %d: %v", peerURL, change.GetNodeId(), err)
		}

	case raftpb.ConfChangeRemoveNode:
		if !isMember {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node %d is not a member of the group", change.GetNodeId())
		}
		if len(m.members)-1 < minVoters {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "removing node %d would leave %d voters, below the minimum of %d", change.GetNodeId(), len(m.members)-1, minVoters)
		}

	default:
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "unsupported config change type %s", change.GetType())
	}

	return nil
}
