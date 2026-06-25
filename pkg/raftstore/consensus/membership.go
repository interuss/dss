package consensus

import (
	"context"
	"encoding/json"
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/raft/v3/raftpb"
)

var errSelfRemoved = stacktrace.NewError("this node was removed from the cluster")

// snapshotEnvelope wraps the application's payload with the current node ID -> peer URL mapping
// and the set of node IDs ever removed from the cluster. Allows any node that catches up via a
// snapshot to fully recover the membership state, including which IDs must never be reused.
type snapshotEnvelope struct {
	Members    map[uint64]string `json:"members"`
	RemovedIDs []uint64          `json:"removed_ids,omitempty"`
	AppData    []byte            `json:"app_data,omitempty"`
}

// reconcileMembers replaces the view of the cluster from newMembers and merges newRemovedIDs into
// the locally known set of permanently removed node IDs (removals are never undone).
func (c *Consensus) reconcileMembers(newMembers map[uint64]string, newRemovedIDs []uint64) error {
	c.membersMu.Lock()

	c.transport.RemoveAllPeers()
	for id, peerURL := range newMembers {
		if id == c.nodeID {
			continue
		}

		c.transport.AddPeer(types.ID(id), []string{peerURL})
	}

	for _, id := range newRemovedIDs {
		c.removedIDs[id] = struct{}{}
	}

	c.members = maps.Clone(newMembers)
	membersSnapshot := maps.Clone(c.members)
	c.membersMu.Unlock()

	if err := c.storage.saveMembers(membersSnapshot); err != nil {
		return stacktrace.Propagate(err, "failed to persist member list recovered from snapshot")
	}

	return nil
}

// membersSnapshot returns a thread-safe copy of the current node ID -> peer URL table.
func (c *Consensus) membersSnapshot() map[uint64]string {
	c.membersMu.Lock()
	defer c.membersMu.Unlock()
	return maps.Clone(c.members)
}

// removedIDsSnapshot returns a thread-safe copy of the set of node IDs ever removed from the
// cluster, as a sorted slice.
func (c *Consensus) removedIDsSnapshot() []uint64 {
	c.membersMu.Lock()
	ids := make([]uint64, 0, len(c.removedIDs))
	for id := range c.removedIDs {
		ids = append(ids, id)
	}
	c.membersMu.Unlock()

	slices.Sort(ids)
	return ids
}

// confChangeContext is used as the raftpb.ConfChangeV2.Context.
// It carries the information needed to apply the config change.
type confChangeContext struct {
	ID     string `json:"id"`
	NodeID uint64 `json:"node_id"`
	URL    string `json:"url,omitempty"`
}

// ProposeConfChange proposes a single membership change
// and blocks until it is applied or ctx is cancelled.
// peerURL is required for raftpb.ConfChangeAddNode, raftpb.ConfChangeAddLearnerNode and raftpb.ConfChangeUpdateNode
// and ignored for raftpb.ConfChangeRemoveNode.
func (c *Consensus) ProposeConfChange(ctx context.Context, changeType raftpb.ConfChangeType, nodeID uint64, peerURL string) (*raftpb.ConfState, error) {
	cctx := confChangeContext{
		ID:     uuid.NewString(),
		NodeID: nodeID,
		URL:    peerURL,
	}

	ctxBuf, err := json.Marshal(cctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal config change context")
	}

	cc := raftpb.ConfChangeV2{
		Changes: []raftpb.ConfChangeSingle{{Type: changeType, NodeID: nodeID}},
		Context: ctxBuf,
	}

	applied := c.tracker.track(cctx.ID)

	err = c.node.ProposeConfChange(ctx, cc)
	if err != nil {
		c.tracker.untrack(cctx.ID, ProposalResult{Error: err})
		return nil, stacktrace.Propagate(err, "failed to propose config change to Raft")
	}

	select {
	case res := <-applied:
		if res.Error != nil {
			return nil, res.Error
		}

		confState, ok := res.Result.(*raftpb.ConfState)
		if !ok {
			return nil, stacktrace.NewError("failed to assert config state type from applied result")
		}

		return confState, nil

	case <-ctx.Done():
		c.tracker.untrack(cctx.ID, ProposalResult{Error: ctx.Err()})
		return nil, ctx.Err()
	}
}

// applyConfigChangeV2Entry applies a committed membership change.
// It updates the config state, the rafthttp transport peers and persists the new member list to disk.
// Returns true if the caller should trigger a snapshot to let a newly added node catch up immediately.
func (c *Consensus) applyConfigChangeV2Entry(data []byte) (bool, error) {
	var cc raftpb.ConfChangeV2
	err := cc.Unmarshal(data)
	if err != nil {
		return false, stacktrace.Propagate(err, "failed to unmarshal config change data")
	}

	var cctx confChangeContext
	if len(cc.Context) > 0 {
		if err := json.Unmarshal(cc.Context, &cctx); err != nil {
			return false, stacktrace.Propagate(err, "failed to unmarshal config change context")
		}
	}

	if len(cc.Changes) != 1 {
		return false, stacktrace.NewError("expected exactly one change in ConfChangeV2, got %d", len(cc.Changes))
	}
	change := cc.Changes[0]

	confState := c.node.ApplyConfChange(cc)
	if confState == nil {
		return false, stacktrace.NewError("failed to apply config change")
	}

	c.confState = *confState

	c.membersMu.Lock()

	triggerSnapshot := false
	selfRemoved := false
	switch change.Type {
	case raftpb.ConfChangeAddNode:
		c.transport.AddPeer(types.ID(change.NodeID), []string{cctx.URL})
		c.members[change.NodeID] = cctx.URL
		triggerSnapshot = true

	case raftpb.ConfChangeUpdateNode:
		c.transport.UpdatePeer(types.ID(change.NodeID), []string{cctx.URL})
		c.members[change.NodeID] = cctx.URL

	case raftpb.ConfChangeRemoveNode:
		if change.NodeID == c.nodeID {
			selfRemoved = true
		} else {
			c.transport.RemovePeer(types.ID(change.NodeID))
		}
		delete(c.members, change.NodeID)
		c.removedIDs[change.NodeID] = struct{}{}
	default:
		c.membersMu.Unlock()
		return false, stacktrace.NewError("unsupported config change type %d", change.Type)
	}
	membersSnapshot := maps.Clone(c.members)

	c.membersMu.Unlock()

	if err := c.storage.saveMembers(membersSnapshot); err != nil {
		return false, stacktrace.Propagate(err, "failed to persist updated member list")
	}

	c.tracker.untrack(cctx.ID, ProposalResult{Result: &c.confState})

	if selfRemoved {
		c.logger.Warn("this node was removed from the cluster, shutting down")
		return false, errSelfRemoved
	}

	return triggerSnapshot, nil
}
