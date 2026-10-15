package consensus

import (
	"bytes"
	"encoding/gob"
	"maps"
	"slices"
	"sync"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
)

// minVoters is the minimum number of voters a Raft group must keep.
// Below it, the group can no longer tolerate the loss of a single node without losing its quorum.
const minVoters = 3

type membership struct {
	sync.RWMutex

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

// restore replaces the members and the removed IDs.
func (m *membership) restore(members map[uint64]string, removedIDs []uint64) {
	m.Lock()
	defer m.Unlock()
	m.members = maps.Clone(members)
	if m.members == nil {
		m.members = map[uint64]string{}
	}
	m.removedIDs = slices.Clone(removedIDs)
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

// snapshot wraps the application snapshot with the membership state that the raft library does not track itself,
// so that a node catching up via a snapshot fully recovers it, including which IDs must never be reused.
// TODO: Use protobuf instead of gob as a follow-up to https://github.com/interuss/dss/pull/1765
type snapshot struct {
	// Members maps the ID of every member of the group to its peer URL.
	Members map[uint64]string
	// RemovedIDs are the IDs of the nodes permanently removed from the group.
	RemovedIDs []uint64
	// AppData is the application snapshot data.
	AppData []byte
}

// encodeSnapshot encodes the application snapshot with the current membership state into the snapshot data.
// TODO: Use protobuf instead of gob as a follow-up to https://github.com/interuss/dss/pull/1765
func (c *Consensus) encodeSnapshot() ([]byte, error) {
	appData, err := c.getAppSnapshot()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get application snapshot data")
	}

	membershipClone := c.membership.clone()
	env := snapshot{
		Members:    membershipClone.members,
		RemovedIDs: membershipClone.removedIDs,
		AppData:    appData,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(env); err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode snapshot")
	}
	return buf.Bytes(), nil
}

// decodeSnapshot decodes the snapshot data produced by encodeSnapshot.
// TODO: Use protobuf instead of gob as a follow-up to https://github.com/interuss/dss/pull/1765
func decodeSnapshot(data []byte) (snapshot, error) {
	var env snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&env); err != nil {
		return snapshot{}, stacktrace.Propagate(err, "failed to decode snapshot")
	}
	return env, nil
}

// restoreMembership replaces the membership state and updates the transport peers accordingly.
func (c *Consensus) restoreMembership(members map[uint64]string, removedIDs []uint64) {
	c.membership.restore(members, removedIDs)
	for _, id := range removedIDs {
		c.transport.RemovePeer(types.ID(id))
	}
	for id, peerURL := range members {
		if id != c.nodeID && peerURL != "" {
			c.transport.AddPeer(types.ID(id), []string{peerURL})
		}
	}
}
