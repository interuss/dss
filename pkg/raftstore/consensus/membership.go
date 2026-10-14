package consensus

import (
	"bytes"
	"encoding/gob"
	"maps"
	"slices"
	"sync"

	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
)

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
