package consensus

import (
	"sync"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
)

// Membership describes a Raft group.
type Membership struct {
	// ClusterID identifies the group within the cluster. It is identical on all the members of the group.
	ClusterID uint64
	// LeaderID is the ID in Nodes of the leader of the group, 0 if it is unknown.
	LeaderID uint64
	// Nodes maps the ID of every node that is or was a member of the group to the node.
	// A node's state tells whether it is a member or was removed.
	// Removed IDs can never be reused.
	Nodes map[uint64]node.Info
}

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
	members memberMap
	// removed maps the ID of every node removed from the group to its description when it was removed.
	// Removed IDs can never be reused.
	removed memberMap
}

func (m *membership) isIDRemoved(id uint64) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.removed[id]
	return ok
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

// removeNode moves nodeID from the members to the removed nodes, which retires its ID, and removes it from the peers
// of the transport.
func (m *membership) removeNode(nodeID uint64) {
	m.Lock()
	m.removed[nodeID] = m.members[nodeID]
	delete(m.members, nodeID)
	m.Unlock()

	if nodeID != m.nodeID {
		m.transport.RemovePeer(types.ID(nodeID))
	}
}

// restore replaces the members and the removed nodes and updates the transport peers accordingly.
func (m *membership) restore(members map[uint64]member, removed map[uint64]member) error {
	for id := range removed {
		if _, ok := members[id]; ok {
			return stacktrace.NewError("removed ID %d is still present in the members map", id)
		}
	}
	if members == nil {
		members = map[uint64]member{}
	}
	if removed == nil {
		removed = map[uint64]member{}
	}

	m.Lock()
	m.members = members
	m.removed = removed
	m.Unlock()

	// It is safe to do this outside the lock because the transport methods as well as addNode and removeNode
	// are only called by the consumer goroutine. So all transport updates happen one at a time. The lock only protects
	// the readers on other goroutine (isIDRemoved, validateConfChange etc.) who never call the transport methods.
	for id := range removed {
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
// When adding a node, mem describes it.
// NB: This method must remain deterministic as it is called by every node when the change is applied.
// It uses members and not the ConfState, which can be more recent than the members after restoring
// a snapshot received from the leader (see Consensus.updateSnapshotConfState).
// A rejected change is reported as a dsserr.BadRequest error, or a dsserr.AlreadyExists error if the URL of the node
// is already used by a member.
func (m *membership) validateConfChange(change *raftpb.ConfChangeSingle, mem member) error {
	m.RLock()
	defer m.RUnlock()

	if change.GetNodeId() == raft.None {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node ID %d is invalid", change.GetNodeId())
	}

	if _, ok := m.removed[change.GetNodeId()]; ok {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node ID %d was removed from the group and cannot be reused", change.GetNodeId())
	}

	existing, isMember := m.members[change.GetNodeId()]
	switch change.GetType() {
	case raftpb.ConfChangeAddNode:
		if isMember && existing.URL != mem.URL {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "the ID %d of the node at %q is already used by member %q: restart the node with an empty data directory to generate a new ID", change.GetNodeId(), mem.URL, existing.URL)
		}
		if isMember {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "node %d is already a member of the group", change.GetNodeId())
		}
		if _, err := types.NewURLs([]string{mem.URL}); err != nil {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "invalid URL %q for node %d: %v", mem.URL, change.GetNodeId(), err)
		}
		for id, other := range m.members {
			if other.URL == mem.URL {
				return stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "the URL %q of node %d is already used by member %d", mem.URL, change.GetNodeId(), id)
			}
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
