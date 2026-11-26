package consensus

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"uuid"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	proposalpb "github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	params "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

type Consensus struct {
	logger *zap.Logger

	nodeID    uint64
	publicURL string
	// config is used to start the raft node.
	config *raft.Config

	tickInterval time.Duration

	node      *node.NodeState
	transport *peerTransport
	consumer  *consumer

	storage *storage
	commitC chan<- EntryCommit

	stopC chan struct{} // when closed, signals shutdown to the raft updates consumer goroutine

	leader leaderRoutines

	tracker     *proposalsTracker
	readTracker *readIndexTracker

	// confChangeC holds a token while a config change proposed by this node is in flight.
	confChangeC chan struct{}

	membership *membership
	metrics    *metrics
}

func NewConsensus(ctx context.Context, logger *zap.Logger, connectParams params.ConnectParameters, provider func() (*proposalpb.Snapshot, error), commitC chan<- EntryCommit) (*Consensus, error) {
	storage, nodeID, hasState, err := newStorage(ctx, logger.With(zap.String("component", "storage")), connectParams.DataDir, connectParams.SnapshotCatchupEntries)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize storage")
	}

	nodeUrl, err := connectParams.RaftPeerUrl()
	if err != nil {
		_ = storage.wal.Close()
		return nil, stacktrace.Propagate(err, "failed to get the URL of the node")
	}

	config := connectParams.RaftConfig(nodeID, storage)
	config.Logger = &ZapRaftLogger{sugar: logger.WithOptions(zap.AddCallerSkip(1)).Sugar()}

	consensus := &Consensus{
		logger: logging.WithValuesFromContext(ctx, logger).With(zap.Uint64("nodeID", nodeID)),

		nodeID:       nodeID,
		publicURL:    nodeUrl.String(),
		config:       config,
		tickInterval: connectParams.TickInterval,
		node:         node.NewNodeState(),

		storage: storage,
		commitC: commitC,

		stopC: make(chan struct{}),

		tracker:     newProposalsTracker(),
		readTracker: newReadIndexTracker(),

		confChangeC: make(chan struct{}, 1),
	}

	consensus.transport, err = newPeerTransport(ctx, consensus.logger, nodeID, connectParams.GroupClusterID, nodeUrl, NewTLSInfo(connectParams, logger), consensus, consensus.Node)
	if err != nil {
		_ = storage.wal.Close()
		return nil, stacktrace.Propagate(err, "failed to initialize transport")
	}

	consensus.membership = &membership{
		nodeID:    nodeID,
		transport: consensus.transport.raw(),
		members:   map[uint64]member{},
		removed:   map[uint64]member{},
	}

	consensus.consumer, err = newConsumer(
		consensus.logger,
		nodeID,
		consensus.node,
		connectParams.TickInterval,
		connectParams.SnapshotIntervalEntries,
		consensus.storage,
		consensus.transport,
		consensus.membership,
		consensus.tracker,
		consensus.readTracker,
		&consensus.leader,
		provider,
		commitC,
		consensus.stopC,
	)
	if err != nil {
		_ = storage.wal.Close()
		return nil, stacktrace.Propagate(err, "failed to initialize consumer")
	}

	consensus.metrics, err = newMetrics(
		connectParams.Group,
		nodeID,
		consensus.node,
		consensus.storage,
		&consensus.consumer.snapshotIndex,
		consensus.tracker,
		consensus.readTracker,
		consensus.transport.leaderStats(),
	)
	if err != nil {
		_ = storage.wal.Close()
		return nil, stacktrace.Propagate(err, "failed to register raft metrics")
	}

	switch {
	case consensus.membership.isIDRemoved(nodeID):
		consensus.logger.Warn("this node was removed from the Raft group")
		consensus.node.SetRemoved()
	case hasState:
		consensus.logger.Info("restarting raft node", zap.String("address", nodeUrl.String()))
		consensus.startNode(func() raft.Node { return raft.RestartNode(consensus.config) })
	default:
		consensus.logger.Info("waiting to be added to a Raft group or to bootstrap one", zap.String("address", nodeUrl.String()))
	}

	return consensus, nil
}

// startNode makes this node a member of the group with the raft node returned by newNode and starts the consumer of
// its updates, if this node is waiting. It returns whether the raft node was started, and the state of this node.
func (c *Consensus) startNode(newNode func() raft.Node) (bool, node.State) {
	started, state := c.node.Start(newNode)
	if started {
		c.consumer.startRaftUpdatesConsumer()
	}
	return started, state
}

// Stop stops the Raft node and closes its WAL. Call it only once.
func (c *Consensus) Stop(ctx context.Context) error {
	c.logger.Info("stopping consensus")
	if err := c.metrics.unregister(); err != nil {
		c.logger.Error("failed to unregister raft metrics", zap.Error(err))
	}
	close(c.stopC)

	c.transport.close()

	var err error
	// only a member still runs a raft node and its consumer
	if raftNode, state := c.node.Get(); state == node.Member {
		select {
		case <-c.consumer.doneC:
		case <-ctx.Done():
			err = stacktrace.Propagate(ctx.Err(), "the raft updates consumer did not exit")
		}
		raftNode.Stop()
		c.logger.Info("raft node stopped")
	}

	if walErr := c.storage.wal.Close(); walErr != nil {
		return errors.Join(err, stacktrace.Propagate(walErr, "failed to close the WAL"))
	}
	c.logger.Info("WAL closed")
	return err
}

// HandleReadPayload blocks until the read-only proposal is served via ReadIndex or until ctx is cancelled.
// ReadIndex confirms (via a quorum-backed leader check) the current commit index, waits for this
// node's own appliedIndex to catch up to it, and only then executes the read against local state.
// This gives the same linearizability guarantee as proposing through the log, without writing an
// entry that every node would otherwise have to persist and replicate.
func (c *Consensus) HandleReadPayload[Result any](ctx context.Context, payload *proposalpb.Payload) (Result, error) {
	defer c.metrics.recordProposal(ctx, time.Now(), "read")

	var zero Result
	proposal := newProposal(ctx, c.nodeID, payload)

	raftNode, err := c.readyNode(ctx)
	if err != nil {
		return zero, err
	}

	ready := c.readTracker.track(proposal.ID)

	err = raftNode.ReadIndex(ctx, []byte(proposal.ID))
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
	case c.commitC <- EntryCommit{Prop: proposal, ReadOnly: true, Done: applyDoneC}:
	case <-ctx.Done():
		return zero, ctx.Err()
	}

	select {
	case res := <-applyDoneC:
		return unwrapProposalResult[Result](proposal.Name(), res)
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

// HandleWritePayload blocks until the proposal is committed and applied / dropped or until ctx is cancelled.
func (c *Consensus) HandleWritePayload[Result any](ctx context.Context, payload *proposalpb.Payload) (Result, error) {
	defer c.metrics.recordProposal(ctx, time.Now(), "write")

	var zero Result
	proposal := newProposal(ctx, c.nodeID, payload)

	raftNode, err := c.readyNode(ctx)
	if err != nil {
		return zero, err
	}

	buf, err := proposal.encode()
	if err != nil {
		return zero, stacktrace.Propagate(err, "failed to encode proposal")
	}

	applied := c.tracker.track(proposal.ID)

	err = raftNode.Propose(ctx, buf)
	if err != nil {
		c.tracker.untrack(proposal.ID, ProposalResult{Error: err})
		return zero, stacktrace.Propagate(err, "failed to propose value to Raft")
	}

	select {
	case res := <-applied:
		return unwrapProposalResult[Result](proposal.Name(), res)
	case <-ctx.Done():
		c.tracker.untrack(proposal.ID, ProposalResult{Error: ctx.Err()})
		return zero, ctx.Err()
	}
}

// AddNode proposes adding the node reachable by its peers at peerURL as a voter of the Raft group.
// The ID of the node is read from the node itself, which must be waiting.
// It blocks until the change is applied or ctx is cancelled, and returns the added node as a member of the group.
// It returns a dsserr.AlreadyExists error if the node is not waiting or does not run this group in this cluster, or
// if peerURL is already used by another member, and a dsserr.Unavailable error if the node cannot be reached or this
// node is not a member of the group.
func (c *Consensus) AddNode(ctx context.Context, peerURL string) (node.Info, error) {
	peer, err := c.transport.fetchPeer(ctx, peerURL)
	if err != nil {
		return node.Info{}, stacktrace.Propagate(err, "failed to fetch node at %s", peerURL)
	}
	if peer.URL != peerURL {
		return node.Info{}, stacktrace.NewErrorWithCode(dsserr.BadRequest, "the node at %s describes itself as being at %s", peerURL, peer.URL)
	}
	if peer.State != node.Waiting {
		return node.Info{}, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "the node at %s is %s, only a waiting node can be added", peerURL, peer.State)
	}

	err = c.proposeConfChange(ctx, &raftpb.ConfChangeSingle{Type: raftpb.ConfChangeAddNode.Enum(), NodeId: new(peer.ID)}, confChangeContext{URL: peerURL})
	if err != nil {
		return node.Info{}, err
	}
	return node.Info{ID: peer.ID, URL: peerURL, State: node.Member}, nil
}

// RemoveNode proposes removing nodeID from the Raft group. Its ID is retired and can never be reused.
// It blocks until the change is applied or ctx is cancelled.
func (c *Consensus) RemoveNode(ctx context.Context, nodeID uint64) error {
	return c.proposeConfChange(ctx, &raftpb.ConfChangeSingle{Type: raftpb.ConfChangeRemoveNode.Enum(), NodeId: new(nodeID)}, confChangeContext{})
}

// Members returns the membership of the Raft group as currently known by this node.
// It returns a dsserr.NotFound error while this node is waiting to bootstrap or join the group.
func (c *Consensus) Members() (Membership, error) {
	if c.node.Current() == node.Waiting {
		return Membership{}, stacktrace.NewErrorWithCode(dsserr.NotFound, "this node is waiting to bootstrap or join the Raft group")
	}

	// A node found removed at startup has no raft node, so it knows no leader.
	var leaderID uint64
	if raftNode := c.node.RaftNode(); raftNode != nil {
		leaderID = raftNode.Status().Lead
	}

	return Membership{
		ClusterID: c.transport.clusterID(),
		LeaderID:  leaderID,
		Nodes:     c.membership.nodes(),
	}, nil
}

func (c *Consensus) Node() (node.Info, error) {
	return node.Info{
		ID:    c.nodeID,
		URL:   c.publicURL,
		State: c.node.Current(),
	}, nil
}

// Bootstrap starts a new Raft group with this node as its only member, and blocks until this node is its leader (bootstrap is completed)
// or ctx is cancelled. Only a waiting node can bootstrap a group, otherwise a dsserr.AlreadyExists error is returned.
func (c *Consensus) Bootstrap(ctx context.Context) (Membership, error) {
	selfCtx, err := json.Marshal(member{URL: c.publicURL})
	if err != nil {
		return Membership{}, stacktrace.Propagate(err, "failed to marshal the context of this node")
	}

	// We pass member as the Context so that every node records the URL of this node when applying the config change.
	started, state := c.startNode(func() raft.Node {
		return raft.StartNode(c.config, []raft.Peer{{ID: c.nodeID, Context: selfCtx}})
	})
	if !started {
		return Membership{}, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "only a waiting node can bootstrap a Raft group, the state of this node is %s", state)
	}
	c.logger.Info("bootstrapping a new Raft group")

	ticker := time.NewTicker(c.tickInterval)
	defer ticker.Stop()
	for {
		membership, err := c.Members()
		if err != nil {
			return Membership{}, err
		}
		if membership.LeaderID == c.nodeID && len(membership.Nodes) > 0 {
			return membership, nil
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return Membership{}, ctx.Err()
		}
	}
}

// RunAsLeader runs fn in the background when this node is the leader of the group.
// The context of fn derives from ctx and is canceled when this node stops being the leader or the raft updates consumer exits.
// fn must return once its context is done, otherwise it keeps running after this node stops being the leader.
func (c *Consensus) RunAsLeader(ctx context.Context, fn func(ctx context.Context)) {
	c.leader.add(ctx, fn)
}

// Ready returns whether this node can serve requests, i.e. it is a member of the Raft group, currently knows its leader
// and the Raft group is not stopped.
func (c *Consensus) Ready() bool {
	select {
	case <-c.stopC:
		return false
	case <-c.consumer.doneC:
		return false
	default:
	}

	select {
	case <-c.consumer.readyC:
	default:
		return false
	}

	return c.node.RaftNode().Status().Lead != raft.None

}

// readyNode returns the raft node, waiting up to an election timeout for this node to be ready,
// or a dsserr.Unavailable error if this node cannot serve requests.
func (c *Consensus) readyNode(ctx context.Context) (raft.Node, error) {
	timeout := time.After(time.Duration(c.config.ElectionTick) * c.tickInterval)
	for !c.Ready() {
		select {
		case <-time.After(c.tickInterval):
		case <-timeout:
			return nil, stacktrace.NewErrorWithCode(dsserr.Unavailable, "this node is not ready to serve requests, its Raft group has no leader or is stopped")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return c.node.Member()
}

// proposeConfChange proposes a single membership change and blocks until it is applied or ctx is cancelled.
// Note that Raft only allows one config change in flight: a change proposed while another one is pending
// is silently dropped by the leader, in which case this blocks until ctx is cancelled.
// The changes proposed by this node are queued, so that they do not drop each other. A change proposed by
// another node, or right after a leader election, can still be dropped.
func (c *Consensus) proposeConfChange(ctx context.Context, change *raftpb.ConfChangeSingle, cctx confChangeContext) error {
	select {
	case c.confChangeC <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-c.confChangeC }()

	raftNode, err := c.node.Member()
	if err != nil {
		return err
	}

	err = c.membership.validateConfChange(change, cctx.member)
	if err != nil {
		return err
	}

	cctx.ID = uuid.New().String()

	ctxBuf, err := json.Marshal(cctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal config change context")
	}

	cc := &raftpb.ConfChangeV2{
		Changes: []*raftpb.ConfChangeSingle{change},
		Context: ctxBuf,
	}

	applied := c.tracker.track(cctx.ID)

	err = raftNode.ProposeConfChange(ctx, cc)
	if err != nil {
		c.tracker.untrack(cctx.ID, ProposalResult{Error: err})
		return stacktrace.Propagate(err, "failed to propose config change to Raft")
	}

	// ProposalResult.Result is always nil for config changes (see applyConfigChangeV2), so we only check the error.
	select {
	case res := <-applied:
		return res.Error

	case <-ctx.Done():
		c.tracker.untrack(cctx.ID, ProposalResult{Error: ctx.Err()})
		return ctx.Err()
	}
}

// Process implements the rafthttp.Raft interface.
// It is called by the transport when a message is received from another Raft node.
func (c *Consensus) Process(ctx context.Context, m *raftpb.Message) error {
	if m.GetTo() != c.nodeID {
		return stacktrace.NewError("received a message for node %d, this node is %d", m.GetTo(), c.nodeID)
	}
	if c.membership.isIDRemoved(m.GetFrom()) {
		c.logger.Warn("rejected a message from a removed node", zap.Uint64("from", m.GetFrom()))
		return removedNodeError{}
	}

	raftNode, state := c.node.Get()
	if state == node.Waiting {
		// A waiting node only receives messages from the leader of a group it was added to, so it means that it is joining the group.
		// It starts with no peers, which raft.RestartNode allows unlike raft.StartNode,
		// and learns the configuration and the application state of the group from the leader.
		if started, _ := c.startNode(func() raft.Node { return raft.RestartNode(c.config) }); started {
			c.logger.Info("joining the Raft group", zap.Uint64("from", m.GetFrom()))
		}
		// The raft node was started either above or concurrently, by another message or by Bootstrap.
		raftNode, state = c.node.Get()
	}
	if state == node.Removed {
		return stacktrace.NewError("this node was removed from the Raft group")
	}
	return raftNode.Step(ctx, m)
}

// removedNodeError is returned by Process for a message from a removed node. rafthttp answers it with a 403, on which
// the removed node stops.
type removedNodeError struct{}

func (removedNodeError) Error() string {
	return "the sender was removed from the Raft group"
}

// WriteTo implements the writerToResponse interface of rafthttp.
func (e removedNodeError) WriteTo(w http.ResponseWriter) {
	http.Error(w, e.Error(), http.StatusForbidden)
}

// IsIDRemoved implements the rafthttp.Raft interface.
func (c *Consensus) IsIDRemoved(id uint64) bool {
	return c.membership.isIDRemoved(id)
}

// ReportUnreachable implements the rafthttp.Raft interface.
func (c *Consensus) ReportUnreachable(id uint64) {
	raftNode, err := c.node.Member()
	if err != nil {
		c.logger.Warn("failed to report unreachable node", zap.Uint64("nodeID", id), zap.Error(err))
		return
	}
	raftNode.ReportUnreachable(id)
}

// ReportSnapshot implements the rafthttp.Raft interface.
func (c *Consensus) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	raftNode, err := c.node.Member()
	if err != nil {
		c.logger.Warn("failed to report snapshot status", zap.Uint64("nodeID", id), zap.Error(err))
		return
	}
	raftNode.ReportSnapshot(id, status)
}
