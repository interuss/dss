package consensus

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	proposalpb "github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	"github.com/interuss/stacktrace"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// errSelfRemoved is returned by consumer.run when this node applies its own removal from the group.
var errSelfRemoved = stacktrace.NewError("this node was removed from the Raft group")

// confChangeContext is used as the raftpb.ConfChangeV2.Context.
// It carries the information needed to apply the config change.
type confChangeContext struct {
	// ID is used to track the proposal until it is applied.
	ID string `json:"id"`
	member
}

// consumer drives the Raft Ready loop and applies committed entries and snapshots
// to local storage, membership, and the application state machine.
type consumer struct {
	logger *zap.Logger
	nodeID uint64
	node   *node.NodeState

	tickInterval     time.Duration
	snapshotInterval uint64

	storage        *storage
	transport      *peerTransport
	membership     *membership
	tracker        *proposalsTracker
	readTracker    *readIndexTracker
	leader         *leaderRoutines
	metrics        *metrics
	getAppSnapshot func() (*proposalpb.Snapshot, error)
	commitC        chan<- EntryCommit

	readyOnce sync.Once
	readyC    chan struct{} // closed once this node is a member of the group and knows its leader
	stopC     chan struct{} // when closed, signals shutdown to the raft updates consumer goroutine
	doneC     chan struct{} // closed when the raft updates consumer exits

	confState     *raftpb.ConfState
	snapshotIndex atomic.Uint64
	appliedIndex  uint64
}

func newConsumer(
	logger *zap.Logger,
	nodeID uint64,
	node *node.NodeState,
	tickInterval time.Duration,
	snapshotInterval uint64,
	storage *storage,
	transport *peerTransport,
	membership *membership,
	tracker *proposalsTracker,
	readTracker *readIndexTracker,
	leader *leaderRoutines,
	getAppSnapshot func() (*proposalpb.Snapshot, error),
	commitC chan<- EntryCommit,
	stopC chan struct{},
) (*consumer, error) {
	c := &consumer{
		logger:           logger,
		nodeID:           nodeID,
		node:             node,
		tickInterval:     tickInterval,
		snapshotInterval: snapshotInterval,
		storage:          storage,
		transport:        transport,
		membership:       membership,
		tracker:          tracker,
		readTracker:      readTracker,
		leader:           leader,
		getAppSnapshot:   getAppSnapshot,
		commitC:          commitC,
		readyC:           make(chan struct{}),
		stopC:            stopC,
		doneC:            make(chan struct{}),
	}

	snap, err := storage.Snapshot()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get snapshot from storage")
	}

	if len(snap.Data) > 0 {
		initSnapshot, err := decodeSnapshot(snap.Data)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to decode snapshot")
		}
		err = membership.restore(initSnapshot.Members, initSnapshot.Removed)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to restore membership")
		}
		commitC <- EntryCommit{Snapshot: initSnapshot.AppSnapshot}
	}

	c.confState = snap.GetMetadata().GetConfState()
	c.snapshotIndex.Store(snap.GetMetadata().GetIndex())
	c.appliedIndex = snap.GetMetadata().GetIndex()

	return c, nil
}

func (c *consumer) entriesToApply(entries []*raftpb.Entry) ([]*raftpb.Entry, error) {
	if len(entries) == 0 {
		return entries, nil
	}

	result := make([]*raftpb.Entry, 0)

	firstIdx := entries[0].GetIndex()
	if firstIdx > c.appliedIndex+1 {
		return nil, stacktrace.NewError("unexpected gap: first committed entry index %d > applied index %d + 1", firstIdx, c.appliedIndex)
	}

	// Skip entries that have already been applied.
	if skip := c.appliedIndex + 1 - firstIdx; skip < uint64(len(entries)) {
		result = entries[skip:]
	}

	return result, nil
}

// startRaftUpdatesConsumer starts a goroutine that processes the Ready channel of the Raft node and applies committed entries to the state machine
// TODO: restart the store when this goroutine stops on an error
func (c *consumer) startRaftUpdatesConsumer() {
	raftNode := c.node.RaftNode()
	go func() {
		defer close(c.doneC)
		defer c.leader.set(false)

		ticker := time.NewTicker(c.tickInterval)
		defer ticker.Stop()

		for {
			select {
			case <-c.stopC:
				return
			case <-ticker.C:
				raftNode.Tick()
			case raftUpdate, ok := <-raftNode.Ready():
				if !ok {
					c.logger.Error("could not read from Ready()")
					return
				}

				if softState := raftUpdate.SoftState; softState != nil {
					if softState.Lead != raft.None {
						c.readyOnce.Do(func() {
							c.logger.Info("ready to serve requests", zap.Uint64("leaderID", softState.Lead))
							close(c.readyC)
						})
					}
					c.leader.set(softState.RaftState == raft.StateLeader)
				}

				err := c.storage.handleReceivedState(raftUpdate.Snapshot, raftUpdate.HardState, raftUpdate.Entries)
				if err != nil {
					c.logger.Error("failed to handle received snapshot", zap.Error(err))
					return
				}

				if !raft.IsEmptySnap(raftUpdate.Snapshot) {
					if raftUpdate.Snapshot.GetMetadata().GetIndex() <= c.appliedIndex {
						c.logger.Error("snapshot index shall be greater than current applied index",
							zap.Uint64("snapshotIndex", raftUpdate.Snapshot.GetMetadata().GetIndex()), zap.Uint64("appliedIndex", c.appliedIndex))
						return
					}

					env, err := decodeSnapshot(raftUpdate.Snapshot.GetData())
					if err != nil {
						c.logger.Error("failed to decode received snapshot", zap.Error(err))
						return
					}

					err = c.membership.restore(env.Members, env.Removed)
					if err != nil {
						c.logger.Error("failed to restore membership", zap.Error(err))
						return
					}
					c.commitC <- EntryCommit{Snapshot: env.AppSnapshot}

					c.confState = raftUpdate.Snapshot.GetMetadata().GetConfState()
					c.snapshotIndex.Store(raftUpdate.Snapshot.GetMetadata().GetIndex())
					c.appliedIndex = raftUpdate.Snapshot.GetMetadata().GetIndex()
				}

				c.updateSnapshotConfState(raftUpdate.Messages)
				c.transport.send(raftUpdate.Messages)

				entries, err := c.entriesToApply(raftUpdate.CommittedEntries)
				if err != nil {
					c.logger.Error("failed to get entries to apply", zap.Error(err))
					return
				}

				err = c.publishEntries(entries)
				if errors.Is(err, errSelfRemoved) {
					c.logger.Warn("this node was removed from the Raft group, stopping")
					c.node.SetRemoved()
					// TODO: let the store do this through Stop, which needs a signal from here
					c.transport.stopPeers()
					raftNode.Stop()
					// to start again as a waiting node
					if err := c.storage.archive(c.nodeID); err != nil {
						c.logger.Error("failed to archive the local state", zap.Error(err))
					}
					return
				}
				if err != nil {
					c.logger.Error("failed to publish entries", zap.Error(err))
					return
				}

				for _, rs := range raftUpdate.ReadStates {
					c.readTracker.setIndex(string(rs.RequestCtx), rs.Index, c.appliedIndex)
				}
				// A read's ReadState may have arrived in an earlier Ready() batch than the
				// entries it depends on, so re-check pending reads on every apply too.
				c.readTracker.releaseUpTo(c.appliedIndex)

				raftNode.Advance()
			case err := <-c.transport.errorC():
				// rafthttp reports critical errors here (e.g. this node was removed) and asks to stop the raft node and transport
				c.logger.Error("transport error", zap.Error(err))
				// TODO: same as above
				c.transport.stopPeers()
				raftNode.Stop()
				return
			}
		}
	}()
}

func (c *consumer) publishEntries(entries []*raftpb.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	c.logger.Debug("publishing entries", zap.Int("numEntries", len(entries)), zap.Uint64("firstIndex", entries[0].GetIndex()), zap.Uint64("lastIndex", entries[len(entries)-1].GetIndex()))

	var wg sync.WaitGroup
	for _, entry := range entries {
		switch entry.GetType() {
		case raftpb.EntryNormal:
			if err := c.submitNormalEntryToStorage(entry.GetData(), &wg); err != nil {
				return stacktrace.Propagate(err, "failed to process normal entry")
			}
		case raftpb.EntryConfChange:
			if err := c.applyConfigChange(entry.GetData()); err != nil {
				return stacktrace.Propagate(err, "failed to process config change entry")
			}
		case raftpb.EntryConfChangeV2:
			if err := c.applyConfigChangeV2(entry.GetData()); err != nil {
				return stacktrace.Propagate(err, "failed to process config change v2 entry")
			}
		}
	}

	// Wait for all entries to be applied before updating the applied index and potentially triggering a snapshot.
	wg.Wait()
	c.appliedIndex = entries[len(entries)-1].GetIndex()
	if c.appliedIndex-c.snapshotIndex.Load() < c.snapshotInterval {
		return nil
	}

	start := time.Now()
	members, removed := c.membership.snapshot()
	data, err := encodeSnapshot(members, removed, c.getAppSnapshot)
	if err != nil {
		return stacktrace.Propagate(err, "failed to encode snapshot")
	}

	if err := c.storage.triggerSnapshot(c.appliedIndex, c.confState, data); err != nil {
		return stacktrace.Propagate(err, "failed to trigger snapshot")
	}

	c.metrics.recordSnapshot(context.Background(), start)
	c.snapshotIndex.Store(c.appliedIndex)
	return nil
}

// submitNormalEntryToStorage passes the proposal to the store and waits for the result to be returned before untracking it.
func (c *consumer) submitNormalEntryToStorage(data []byte, wg *sync.WaitGroup) error {
	if len(data) == 0 {
		return nil
	}

	proposal, err := decodeProposal(data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to decode committed proposal")
	}

	applyDoneC := make(chan ProposalResult, 1)
	wg.Go(func() {
		// Ensure that the tracker is cleaned up.
		c.tracker.untrack(proposal.ID, <-applyDoneC)
	})

	c.commitC <- EntryCommit{Prop: proposal, Done: applyDoneC}
	return nil
}

// raftpb.ConfChange is still used internally by Raft to bootstrap the group; we apply the change to the node
// and record the URL of the bootstrap members. Changes requested by clients are processed by applyConfigChangeV2.
func (c *consumer) applyConfigChange(data []byte) error {
	cc := &raftpb.ConfChange{}
	if err := proto.Unmarshal(data, cc); err != nil {
		return stacktrace.Propagate(err, "failed to unmarshal config change data")
	}

	c.confState = c.node.RaftNode().ApplyConfChange(cc)

	if cc.GetType() == raftpb.ConfChangeAddNode {
		var m member
		if err := json.Unmarshal(cc.GetContext(), &m); err != nil {
			return stacktrace.Propagate(err, "failed to unmarshal context of config change adding node %d", cc.GetNodeId())
		}
		c.membership.addNode(cc.GetNodeId(), m)
	}
	return nil
}

// applyConfigChangeV2 applies a committed membership change requested by a client.
// Invalid changes are rejected and leave the configuration unchanged.
// It returns errSelfRemoved if this node was removed from the group.
func (c *consumer) applyConfigChangeV2(data []byte) error {
	cc := &raftpb.ConfChangeV2{}
	err := proto.Unmarshal(data, cc)
	if err != nil {
		return stacktrace.Propagate(err, "failed to unmarshal config change data")
	}

	var cctx confChangeContext
	if len(cc.GetContext()) > 0 {
		if err := json.Unmarshal(cc.GetContext(), &cctx); err != nil {
			return stacktrace.Propagate(err, "failed to unmarshal config change context")
		}
	}

	if len(cc.GetChanges()) != 1 {
		err = stacktrace.NewErrorWithCode(dsserr.BadRequest, "expected exactly one change, got %d", len(cc.GetChanges()))
	} else {
		err = c.membership.validateConfChange(cc.GetChanges()[0], cctx.member)
	}
	if err != nil {
		c.logger.Warn("rejecting config change", zap.String("id", cctx.ID), zap.Error(err))
		// A rejected change is treated as a no-op.
		c.tracker.untrack(cctx.ID, ProposalResult{Error: err})
		return nil
	}

	change := cc.GetChanges()[0]
	c.confState = c.node.RaftNode().ApplyConfChange(cc)

	selfRemoved := false
	switch change.GetType() {
	case raftpb.ConfChangeAddNode:
		c.membership.addNode(change.GetNodeId(), cctx.member)

	case raftpb.ConfChangeRemoveNode:
		c.membership.removeNode(change.GetNodeId())
		selfRemoved = change.GetNodeId() == c.nodeID
	}

	c.logger.Info("applied config change", zap.String("id", cctx.ID), zap.Stringer("type", change.GetType()), zap.Uint64("nodeID", change.GetNodeId()))
	c.tracker.untrack(cctx.ID, ProposalResult{})

	if selfRemoved {
		return errSelfRemoved
	}
	return nil
}

// updateSnapshotConfState replaces the ConfState of the snapshots sent to followers with the current one.
func (c *consumer) updateSnapshotConfState(msgs []*raftpb.Message) {
	for i := range msgs {
		if msgs[i].GetType() == raftpb.MsgSnap {
			msgs[i].Snapshot.Metadata.ConfState = c.confState
		}
	}
}
