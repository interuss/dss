package consensus

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/interuss/dss/pkg/logging"
	params "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	v2stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"

	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const serverShutdownTimeout = 5 * time.Second

type Consensus struct {
	logger *zap.Logger

	nodeID uint64
	node   raft.Node

	transport *rafthttp.Transport
	server    *http.Server

	storage   *storage
	commitChs map[string]chan EntryCommit

	tracker *proposalsTracker

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64
}

func NewConsensus(ctx context.Context, logger *zap.Logger, peers map[uint64]*url.URL, connectParams params.ConnectParameters) (*Consensus, error) {
	storage, old, err := newStorage(ctx, logger.With(zap.String("component", "storage")), connectParams.DataDir, connectParams.NodeID, connectParams.SnapshotCatchupEntries)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize storage")
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
		node = raft.StartNode(config, peersList(peers))
	}

	consensus := &Consensus{
		logger: logging.WithValuesFromContext(ctx, logger),

		nodeID: connectParams.NodeID,
		node:   node,

		storage:   storage,
		commitChs: make(map[string]chan EntryCommit),
		tracker:   newProposalsTracker(),
	}

	err = consensus.initTransport(ctx, connectParams.NodeID, connectParams.ClusterID, peers)
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

	go func() {
		err := consensus.handleReady(connectParams.TickInterval, connectParams.SnapshotIntervalEntries)
		if err != nil {
			consensus.logger.Error("handleReady exited with error, shutting down consensus", zap.Error(err))
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()
		if shutdownErr := consensus.server.Shutdown(shutdownCtx); shutdownErr != nil {
			consensus.logger.Error("failed to shutdown http server", zap.Error(shutdownErr))
		} else {
			consensus.logger.Info("http server shutdown complete")
		}

		consensus.transport.Stop()
		consensus.logger.Info("transport stopped")
		consensus.node.Stop()
		consensus.logger.Info("raft node stopped")
	}()

	return consensus, nil
}

// ProposeValue blocks until the proposal is committed and applied / dropped or until ctx is cancelled.
func (c *Consensus) ProposeValue(ctx context.Context, proposal Proposal) (any, error) {
	buf, err := json.Marshal(proposal)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal proposal")
	}

	applied := c.tracker.track(proposal.ID)

	err = c.node.Propose(ctx, buf)
	if err != nil {
		c.tracker.untrack(proposal.ID, ProposalResult{Error: err})
		return nil, stacktrace.Propagate(err, "failed to propose value to Raft")
	}

	select {
	case res := <-applied:
		return res.Result, res.Error

	case <-ctx.Done():
		c.tracker.untrack(proposal.ID, ProposalResult{Error: ctx.Err()})
		return nil, ctx.Err()
	}
}

func peersList(peers map[uint64]*url.URL) []raft.Peer {
	result := make([]raft.Peer, 0, len(peers))
	for id := range peers {
		result = append(result, raft.Peer{ID: id})
	}
	return result
}

func (c *Consensus) initTransport(ctx context.Context, nodeID uint64, clusterID uint64, peers map[uint64]*url.URL) error {
	connectParams := params.GetConnectParameters()

	if connectParams.Insecure {
		c.logger.Warn("Running in insecure mode: HTTPS is disabled or certificates are not verified. Use only for testing or trusted environments.")
	}

	tlsInfo := transport.TLSInfo{
		TrustedCAFile:      connectParams.CAFile,
		CertFile:           connectParams.CertFile,
		KeyFile:            connectParams.KeyFile,
		InsecureSkipVerify: connectParams.Insecure,
	}

	nodeIDStr := fmt.Sprintf("%d", nodeID)

	transport := &rafthttp.Transport{
		Logger:      logging.WithValuesFromContext(ctx, c.logger.With(zap.String("component", "transport"))),
		ID:          types.ID(nodeID),
		ClusterID:   types.ID(clusterID),
		Raft:        c,
		ServerStats: v2stats.NewServerStats(nodeIDStr, nodeIDStr),
		LeaderStats: v2stats.NewLeaderStats(c.logger, nodeIDStr),
		ErrorC:      make(chan error),
	}

	if !(connectParams.Insecure && connectParams.CAFile == "" && connectParams.CertFile == "" && connectParams.KeyFile == "") {
		transport.TLSInfo = tlsInfo
	}

	if err := transport.Start(); err != nil {
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

	var err error
	c.server, err = c.createServer(listeningAddr, transport.Handler(), tlsInfo, connectParams.Insecure)
	if err != nil {
		return err
	}

	c.transport = transport
	return nil
}

func (c *Consensus) createServer(addr string, handler http.Handler, tlsInfo transport.TLSInfo, insecure bool) (*http.Server, error) {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	if insecure {
		if tlsInfo.CertFile == "" || tlsInfo.KeyFile == "" {
			// Certificates provided: no HTTPS
			go c.startServer(server, false, "", "")
		} else {
			// Certificates provided: HTTPS with no verification
			cfg, err := tlsInfo.ServerConfig()
			if err != nil {
				return nil, stacktrace.NewError("failed to create TLS config")
			}
			cfg.ClientAuth = tls.NoClientCert
			server.TLSConfig = cfg
			go c.startServer(server, true, tlsInfo.CertFile, tlsInfo.KeyFile)
		}
	} else {
		// Secure mode: HTTPS with full verification
		cfg, err := tlsInfo.ServerConfig()
		if err != nil {
			return nil, stacktrace.NewError("failed to create TLS config")
		}
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		server.TLSConfig = cfg
		go c.startServer(server, true, tlsInfo.CertFile, tlsInfo.KeyFile)
	}

	return server, nil
}

func (c *Consensus) startServer(server *http.Server, useTLS bool, certFile, keyFile string) {
	var err error
	if useTLS {
		err = server.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		c.logger.Error("http server error", zap.Error(err))
		c.transport.ErrorC <- err
	}
}

// handleReady processes the Ready channel of the Raft node and applies committed entries to the state machine
func (c *Consensus) handleReady(tickInterval time.Duration, snapshotInterval uint64) error {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.node.Tick()
		case err := <-c.transport.ErrorC:
			return stacktrace.Propagate(err, "transport error")
		case rd, ok := <-c.node.Ready():
			if !ok {
				return stacktrace.NewError("could not read from Ready(), shutting down handler")
			}

			err := c.storage.handleReceivedState(rd.Snapshot, rd.HardState, rd.Entries)
			if err != nil {
				return stacktrace.Propagate(err, "failed to handle received snapshot")
			}

			if !raft.IsEmptySnap(rd.Snapshot) {
				if rd.Snapshot.Metadata.Index <= c.appliedIndex {
					return stacktrace.NewError("snapshot index %d shall be greater than current applied index %d", rd.Snapshot.Metadata.Index, c.appliedIndex)
				}

				err = c.dispatchSnapshot(rd.Snapshot.Data)
				if err != nil {
					return stacktrace.Propagate(err, "failed to dispatch snapshot")
				}

				c.confState = rd.Snapshot.Metadata.ConfState
				c.snapshotIndex = rd.Snapshot.Metadata.Index
				c.appliedIndex = rd.Snapshot.Metadata.Index
			}

			c.updateSnapshotConfState(rd.Messages)
			c.transport.Send(rd.Messages)

			entries, err := c.entriesToApply(rd.CommittedEntries)
			if err != nil {
				return stacktrace.Propagate(err, "failed to get entries to apply")
			}

			err = c.publishEntries(entries, snapshotInterval)
			if err != nil {
				return stacktrace.Propagate(err, "failed to publish entries")
			}

			c.node.Advance()
		}
	}
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
			err := c.processNormalEntry(entry.Data, &wg)
			if err != nil {
				return stacktrace.Propagate(err, "failed to process normal entry")
			}
		case raftpb.EntryConfChange:
			err := c.processConfigChangeEntry(entry.Data)
			if err != nil {
				return stacktrace.Propagate(err, "failed to process config change entry")
			}
		case raftpb.EntryConfChangeV2:
			triggerSnapshot, err = c.processConfigChangeV2Entry(entry.Data)
			if err != nil {
				return stacktrace.Propagate(err, "failed to process config change v2 entry")
			}
		}
	}

	// wait for all entries to be applied before updating the applied index and potentially triggering a snapshot
	wg.Wait()
	c.appliedIndex = entries[len(entries)-1].Index

	if triggerSnapshot || c.appliedIndex-c.snapshotIndex >= snapshotInterval {
		err := c.storage.triggerSnapshot(c.appliedIndex, &c.confState)
		if err != nil {
			return stacktrace.Propagate(err, "failed to trigger snapshot")
		}

		c.snapshotIndex = c.appliedIndex
	}

	return nil
}

// processNormalEntry passes the proposal to the store and waits for the result to be returned before untracking it.
func (c *Consensus) processNormalEntry(data []byte, wg *sync.WaitGroup) error {
	if len(data) <= 0 {
		return nil
	}

	prop := Proposal{}
	err := json.Unmarshal(data, &prop)
	if err != nil {
		return stacktrace.Propagate(err, "failed to unmarshal committed proposal")
	}

	//if readOnly proposal and we did not initiate it, skip it (noop)
	if prop.ReadOnly && !c.tracker.isPending(prop.ID) {
		return nil
	}

	applyDoneC := make(chan ProposalResult, 1)
	wg.Go(func() {
		res := <-applyDoneC
		if c.tracker.isPending(prop.ID) {
			c.tracker.untrack(prop.ID, res)
		}
	})

	ch, ok := c.commitChs[prop.DBName]
	if !ok {
		return stacktrace.NewError("no commit channel found for %s", prop.DBName)
	}

	ch <- EntryCommit{Prop: prop, Done: applyDoneC}
	return nil
}

func (c *Consensus) dispatchSnapshot(snapshotData []byte) error {
	var snapshot map[string][]byte
	err := json.Unmarshal(snapshotData, &snapshot)
	if err != nil {
		return stacktrace.Propagate(err, "failed to unmarshal snapshot data")
	}

	var eg errgroup.Group
	for name, data := range snapshot {
		eg.Go(func() error {
			ch, ok := c.commitChs[name]
			if !ok {
				return stacktrace.NewError("no commit channel found for %s", name)
			}

			ch <- EntryCommit{SnapshotData: data}
			return nil
		})
	}

	return eg.Wait()
}

// raftpb.ConfChange is still used internally by Raft, we just need to apply the change to the node.
// Changes requested by clients are processed by processConfigChangeV2Entry.
func (c *Consensus) processConfigChangeEntry(data []byte) error {
	var cc raftpb.ConfChange
	err := cc.Unmarshal(data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to unmarshal config change data")
	}

	c.confState = *c.node.ApplyConfChange(cc)
	return nil
}

func (c *Consensus) processConfigChangeV2Entry(data []byte) (bool, error) {
	var cc raftpb.ConfChangeV2
	err := cc.Unmarshal(data)
	if err != nil {
		return false, stacktrace.Propagate(err, "failed to unmarshal config change data")
	}

	c.confState = *c.node.ApplyConfChange(cc)

	// TODO - implement config changes when triggered by a proposal
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

// Process implements the rafthttp.Raft interface.
func (c *Consensus) Process(ctx context.Context, m raftpb.Message) error {
	return c.node.Step(ctx, m)
}

// IsIDRemoved implements the rafthttp.Raft interface.
func (c *Consensus) IsIDRemoved(id uint64) bool {
	return false
}

// ReportUnreachable implements the rafthttp.Raft interface.
func (c *Consensus) ReportUnreachable(id uint64) {
	c.node.ReportUnreachable(id)
}

// ReportSnapshot implements the rafthttp.Raft interface.
func (c *Consensus) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	c.node.ReportSnapshot(id, status)
}

// RegisterStore allows registering a snapshot provider function for a specific store
// and returns the channel on which committed entries for that store will be sent.
func (c *Consensus) RegisterStore(name string, provider snapshotProvider) chan EntryCommit {
	c.storage.registerSnapshotProvider(name, provider)
	ch := make(chan EntryCommit)
	c.commitChs[name] = ch
	return ch
}
