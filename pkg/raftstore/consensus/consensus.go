package consensus

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/interuss/dss/pkg/logging"
	params "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	v2stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

// snapshotProvider is a function that returns the snapshot data to be included in the Raft snapshot.
type snapshotProvider func() ([]byte, error)

type Consensus struct {
	logger *zap.Logger

	nodeID uint64
	node   raft.Node

	transport *rafthttp.Transport

	storage        *storage
	getAppSnapshot func() ([]byte, error)
	commitC        chan<- EntryCommit

	once            sync.Once
	shutdownTimeout time.Duration
	serverErrC      chan error    // errors of the shared http server
	stopC           chan struct{} // when closed, signals shutdown to the raft updates consumer goroutine

	tracker     *proposalsTracker
	readTracker *readIndexTracker

	membership membership
	confState  raftpb.ConfState

	snapshotIndex uint64
	appliedIndex  uint64
}

func NewConsensus(ctx context.Context, logger *zap.Logger, connectParams params.ConnectParameters, provider snapshotProvider, commitC chan<- EntryCommit) (*Consensus, error) {
	storage, old, err := newStorage(ctx, logger.With(zap.String("component", "storage")), connectParams.DataDir, connectParams.NodeID, connectParams.SnapshotCatchupEntries)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize storage")
	}

	peers, err := connectParams.PeerMap()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to parse peer map")
	}

	nodeUrl, ok := peers[connectParams.NodeID]
	if !ok {
		return nil, stacktrace.NewError("node ID %d not found in peers map", connectParams.NodeID)
	}

	var node raft.Node
	config := connectParams.RaftConfig(storage)
	if old {
		logger.Info("restarting raft node", zap.String("address", nodeUrl.String()))
		node = raft.RestartNode(config)
	} else {
		logger.Info("starting new raft node", zap.String("address", nodeUrl.String()))
		raftPeers, err := peersList(peers)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to build the list of bootstrap peers")
		}
		node = raft.StartNode(config, raftPeers)
	}

	consensus := &Consensus{
		logger: logging.WithValuesFromContext(ctx, logger),

		nodeID:         connectParams.NodeID,
		node:           node,
		storage:        storage,
		getAppSnapshot: provider,
		commitC:        commitC,

		shutdownTimeout: 2 * connectParams.ElectionInterval(),
		serverErrC:      make(chan error, 1),
		stopC:           make(chan struct{}),

		tracker:     newProposalsTracker(),
		readTracker: newReadIndexTracker(),

		membership: membership{
			nodeID:  connectParams.NodeID,
			members: map[uint64]member{},
		},
	}

	err = consensus.initTransport(ctx, connectParams.NodeID, connectParams.GroupClusterID, peers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize transport")
	}

	snap, err := consensus.storage.Snapshot()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get snapshot from storage")
	}

	consensus.confState = snap.Metadata.ConfState
	consensus.snapshotIndex = snap.Metadata.Index
	consensus.appliedIndex = snap.Metadata.Index

	if len(snap.Data) > 0 {
		initSnapshot, err := decodeSnapshot(snap.Data)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to decode snapshot")
		}
		err = consensus.membership.restore(initSnapshot.Members, initSnapshot.RemovedIDs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to restore membership")
		}
		consensus.commitC <- EntryCommit{SnapshotData: initSnapshot.AppData}
	}

	consensus.startRaftUpdatesConsumer(connectParams.TickInterval, connectParams.SnapshotIntervalEntries)

	return consensus, nil
}

// TODO: ctx is currently ignored (see issue: https://github.com/interuss/dss/issues/1610)
func (c *Consensus) Stop(ctx context.Context) {
	// TODO: remove once (see issue: https://github.com/interuss/dss/issues/1610)
	c.once.Do(func() {
		c.logger.Info("stopping consensus")
		close(c.stopC)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), c.shutdownTimeout)
		defer cancel()
		if shutdownErr := unregisterGroup(shutdownCtx, c.transport.ClusterID); shutdownErr != nil {
			c.logger.Error("failed to shutdown http server", zap.Error(shutdownErr))
		} else {
			c.logger.Info("unregistered from http server")
		}

		c.transport.Stop()
		c.logger.Info("transport stopped")
		c.node.Stop()
		c.logger.Info("raft node stopped")
	})
}

// RequestType identifies a kind of client request and the concrete type its result is expected to have
type RequestType[Result any] string

// HandleReadRequest blocks until the read-only proposal is served via ReadIndex or until ctx is cancelled.
// ReadIndex confirms (via a quorum-backed leader check) the current commit index, waits for this
// node's own appliedIndex to catch up to it, and only then executes the read against local state.
// This gives the same linearizability guarantee as proposing through the log, without writing an
// entry that every node would otherwise have to persist and replicate.
func (c *Consensus) HandleReadRequest[Result any](ctx context.Context, requestType RequestType[Result], value []byte) (Result, error) {
	var zero Result

	proposal := c.newProposal(ctx, string(requestType), value, true)

	ready := c.readTracker.track(proposal.ID)

	err := c.node.ReadIndex(ctx, []byte(proposal.ID))
	if err != nil {
		c.readTracker.untrack(proposal.ID)
		return zero, stacktrace.Propagate(err, "failed to request read index from Raft")
	}

	select {
	case <-ready:
	case <-ctx.Done():
		c.readTracker.untrack(proposal.ID)
		return zero, ctx.Err()
	}

	applyDoneC := make(chan ProposalResult, 1)
	select {
	case c.commitC <- EntryCommit{Prop: proposal, Done: applyDoneC}:
	case <-ctx.Done():
		return zero, ctx.Err()
	}

	select {
	case res := <-applyDoneC:
		if res.Error != nil {
			return zero, res.Error
		}
		if res.Result == nil {
			return zero, nil
		}
		result, ok := res.Result.(Result)
		if !ok {
			return zero, stacktrace.NewError("unexpected result type for %q: got %T", requestType, res.Result)
		}
		return result, nil

	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

// HandleWriteRequest blocks until the proposal is committed and applied / dropped or until ctx is cancelled.
func (c *Consensus) HandleWriteRequest[Result any](ctx context.Context, requestType RequestType[Result], value []byte) (Result, error) {
	var zero Result

	proposal := c.newProposal(ctx, string(requestType), value, false)

	buf, err := json.Marshal(proposal)
	if err != nil {
		return zero, stacktrace.Propagate(err, "failed to marshal proposal")
	}

	applied := c.tracker.track(proposal.ID)

	err = c.node.Propose(ctx, buf)
	if err != nil {
		c.tracker.untrack(proposal.ID, ProposalResult{Error: err})
		return zero, stacktrace.Propagate(err, "failed to propose value to Raft")
	}

	select {
	case res := <-applied:
		if res.Error != nil {
			return zero, res.Error
		}
		if res.Result == nil {
			return zero, nil
		}
		result, ok := res.Result.(Result)
		if !ok {
			return zero, stacktrace.NewError("unexpected result type for %q: got %T", requestType, res.Result)
		}
		return result, nil

	case <-ctx.Done():
		c.tracker.untrack(proposal.ID, ProposalResult{Error: ctx.Err()})
		return zero, ctx.Err()
	}
}

// member describes a member of the group.
// It is passed as the context of the config changes adding a node.
type member struct {
	URL string `json:"url"`
}

func peersList(peers map[uint64]*url.URL) ([]raft.Peer, error) {
	result := make([]raft.Peer, 0, len(peers))
	for id, peerURL := range peers {
		peerCtx, err := json.Marshal(member{URL: peerURL.String()})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to marshal context of peer %d", id)
		}

		// We pass member as the Context so that we can recover the peer URL when applying a config change.
		result = append(result, raft.Peer{ID: id, Context: peerCtx})
	}
	return result, nil
}

func (c *Consensus) initTransport(ctx context.Context, nodeID uint64, clusterID uint64, peers map[uint64]*url.URL) error {
	nodeIDStr := fmt.Sprintf("%d", nodeID)

	transport := &rafthttp.Transport{
		Logger:      logging.WithValuesFromContext(ctx, c.logger.With(zap.String("component", "transport"))),
		ID:          types.ID(nodeID),
		ClusterID:   types.ID(clusterID),
		Raft:        c,
		ServerStats: v2stats.NewServerStats(nodeIDStr, nodeIDStr),
		LeaderStats: v2stats.NewLeaderStats(c.logger, nodeIDStr),
		// ErrorC is buffered to make sure rafthttp can send errors even if the raft updates consumer is not listening yet.
		// The buffer size of 1 is sufficient because the transport will only send one error before stopping.
		ErrorC: make(chan error, 1),
	}

	err := transport.Start()
	if err != nil {
		return stacktrace.Propagate(err, "failed to start transport")
	}

	var listeningAddr string
	for peerID, peerURL := range peers {
		if peerID == nodeID {
			listeningAddr = ":" + peerURL.Port()
			continue
		}

		transport.AddPeer(types.ID(peerID), []string{peerURL.String()})
	}

	if listeningAddr == "" {
		return stacktrace.NewError("node ID %d not found in peers map", nodeID)
	}

	c.transport = transport
	c.membership.transport = transport

	err = registerGroup(ctx, c.logger.With(zap.String("component", "listener")), listeningAddr, transport.ClusterID, transport.Handler(), c.serverErrC)
	if err != nil {
		return stacktrace.Propagate(err, "failed to register the transport to the http server")
	}

	return nil
}

// startRaftUpdatesConsumer starts a goroutine that processes the Ready channel of the Raft node and applies committed entries to the state machine
func (c *Consensus) startRaftUpdatesConsumer(tickInterval time.Duration, snapshotInterval uint64) {
	go func() {
		// TODO: this shouldn't be triggered from inside the consensus instance, removing it will allow removing the once.
		// (see issue: https://github.com/interuss/dss/issues/1610)
		defer c.Stop(context.Background())

		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()

		for {
			select {
			case <-c.stopC:
				return
			case <-ticker.C:
				c.node.Tick()
			case raftUpdate, ok := <-c.node.Ready():
				if !ok {
					c.logger.Error("could not read from Ready()")
					return
				}

				err := c.storage.handleReceivedState(raftUpdate.Snapshot, raftUpdate.HardState, raftUpdate.Entries)
				if err != nil {
					c.logger.Error("failed to handle received snapshot", zap.Error(err))
					return
				}

				if !raft.IsEmptySnap(raftUpdate.Snapshot) {
					if raftUpdate.Snapshot.Metadata.Index <= c.appliedIndex {
						c.logger.Error("snapshot index shall be greater than current applied index",
							zap.Uint64("snapshotIndex", raftUpdate.Snapshot.Metadata.Index), zap.Uint64("appliedIndex", c.appliedIndex))
						return
					}

					env, err := decodeSnapshot(raftUpdate.Snapshot.Data)
					if err != nil {
						c.logger.Error("failed to decode received snapshot", zap.Error(err))
						return
					}

					err = c.membership.restore(env.Members, env.RemovedIDs)
					if err != nil {
						c.logger.Error("failed to restore membership", zap.Error(err))
						return
					}
					c.commitC <- EntryCommit{SnapshotData: env.AppData}

					c.confState = raftUpdate.Snapshot.Metadata.ConfState
					c.snapshotIndex = raftUpdate.Snapshot.Metadata.Index
					c.appliedIndex = raftUpdate.Snapshot.Metadata.Index
				}

				c.updateSnapshotConfState(raftUpdate.Messages)
				c.transport.Send(raftUpdate.Messages)

				entries, err := c.entriesToApply(raftUpdate.CommittedEntries)
				if err != nil {
					c.logger.Error("failed to get entries to apply", zap.Error(err))
					return
				}

				err = c.publishEntries(entries, snapshotInterval)
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

				c.node.Advance()
			case err := <-c.transport.ErrorC:
				c.logger.Error("transport error", zap.Error(err))
				return
			case err := <-c.serverErrC:
				c.logger.Error("http server error", zap.Error(err))
				return
			}
		}
	}()
}

func (c *Consensus) publishEntries(entries []raftpb.Entry, snapshotInterval uint64) error {
	if len(entries) == 0 {
		return nil
	}

	c.logger.Info("publishing entries", zap.Int("numEntries", len(entries)), zap.Uint64("firstIndex", entries[0].Index), zap.Uint64("lastIndex", entries[len(entries)-1].Index))

	var triggerSnapshot bool
	var err error
	var wg sync.WaitGroup
	for _, entry := range entries {
		switch entry.Type {
		case raftpb.EntryNormal:
			err := c.submitNormalEntryToStorage(entry.Data, &wg)
			if err != nil {
				return stacktrace.Propagate(err, "failed to process normal entry")
			}
		case raftpb.EntryConfChange:
			err := c.applyConfigChange(entry.Data)
			if err != nil {
				return stacktrace.Propagate(err, "failed to process config change entry")
			}
		case raftpb.EntryConfChangeV2:
			triggerSnapshot, err = c.applyConfigChangeV2(entry.Data)
			if err != nil {
				return stacktrace.Propagate(err, "failed to process config change v2 entry")
			}
		}
	}

	// wait for all entries to be applied before updating the applied index and potentially triggering a snapshot
	wg.Wait()
	c.appliedIndex = entries[len(entries)-1].Index

	if triggerSnapshot || c.appliedIndex-c.snapshotIndex >= snapshotInterval {
		data, err := c.encodeSnapshot()
		if err != nil {
			return stacktrace.Propagate(err, "failed to encode snapshot")
		}

		err = c.storage.triggerSnapshot(c.appliedIndex, &c.confState, data)
		if err != nil {
			return stacktrace.Propagate(err, "failed to trigger snapshot")
		}

		c.snapshotIndex = c.appliedIndex
	}

	return nil
}

// submitNormalEntryToStorage passes the proposal to the store and waits for the result to be returned before untracking it.
func (c *Consensus) submitNormalEntryToStorage(data []byte, wg *sync.WaitGroup) error {
	if len(data) <= 0 {
		return nil
	}

	var proposal Proposal
	err := json.Unmarshal(data, &proposal)
	if err != nil {
		return stacktrace.Propagate(err, "failed to unmarshal committed proposal")
	}

	applyDoneC := make(chan ProposalResult, 1)
	wg.Go(func() {
		// Ensure that the tracker is cleaned up
		c.tracker.untrack(proposal.ID, <-applyDoneC)
	})

	c.commitC <- EntryCommit{Prop: proposal, Done: applyDoneC}
	return nil
}

// raftpb.ConfChange is still used internally by Raft to bootstrap the group, we just need to apply the change to the node
// and record the URL of the bootstrap members. Changes requested by clients are processed by applyConfigChangeV2.
func (c *Consensus) applyConfigChange(data []byte) error {
	var cc raftpb.ConfChange
	err := cc.Unmarshal(data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to unmarshal config change data")
	}

	c.confState = *c.node.ApplyConfChange(cc)

	if cc.Type == raftpb.ConfChangeAddNode {
		var m member
		err = json.Unmarshal(cc.Context, &m)
		if err != nil {
			return stacktrace.Propagate(err, "failed to unmarshal context of config change adding node %d", cc.NodeID)
		}
		c.membership.addNode(cc.NodeID, m)
	}
	return nil
}

func (c *Consensus) applyConfigChangeV2(data []byte) (bool, error) { //nolint:unparam
	var cc raftpb.ConfChangeV2
	err := cc.Unmarshal(data)
	if err != nil {
		return false, stacktrace.Propagate(err, "failed to unmarshal config change data")
	}

	c.confState = *c.node.ApplyConfChange(cc)

	// TODO - implement config changes when triggered by a client
	return false, nil
}

func (c *Consensus) entriesToApply(entries []raftpb.Entry) ([]raftpb.Entry, error) {
	if len(entries) == 0 {
		return entries, nil
	}

	result := make([]raftpb.Entry, 0)

	firstIdx := entries[0].Index
	if firstIdx > c.appliedIndex+1 {
		return nil, stacktrace.NewError("unexpected gap: first committed entry index %d > applied index %d + 1", firstIdx, c.appliedIndex)
	}

	// Skip entries that have already been applied.
	if skip := c.appliedIndex + 1 - firstIdx; skip < uint64(len(entries)) {
		result = entries[skip:]
	}

	return result, nil
}

// updateSnapshotConfState updates the ConfState in the snapshot
// of messages that contain one as it could be outdated.
func (c *Consensus) updateSnapshotConfState(msgs []raftpb.Message) {
	for i := range msgs {
		if msgs[i].Type == raftpb.MsgSnap {
			msgs[i].Snapshot.Metadata.ConfState = c.confState
		}
	}
}

// snapshot wraps the application snapshot with the membership state that the raft library does not track itself,
// so that a node catching up via a snapshot fully recovers it, including which IDs must never be reused.
// TODO: Use protobuf instead of gob as a follow-up to https://github.com/interuss/dss/pull/1765
type snapshot struct {
	// Members maps the ID of every member of the group to its description.
	Members map[uint64]member
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

	env := snapshot{
		Members:    c.membership.members,
		RemovedIDs: c.membership.removedIDs,
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

// Process implements the rafthttp.Raft interface.
func (c *Consensus) Process(ctx context.Context, m raftpb.Message) error {
	return c.node.Step(ctx, m)
}

// IsIDRemoved implements the rafthttp.Raft interface.
func (c *Consensus) IsIDRemoved(id uint64) bool {
	return c.membership.isIDRemoved(id)
}

// ReportUnreachable implements the rafthttp.Raft interface.
func (c *Consensus) ReportUnreachable(id uint64) {
	c.node.ReportUnreachable(id)
}

// ReportSnapshot implements the rafthttp.Raft interface.
func (c *Consensus) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	c.node.ReportSnapshot(id, status)
}
