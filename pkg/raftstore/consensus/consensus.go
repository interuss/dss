package consensus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

type Consensus struct {
	logger *zap.Logger

	nodeID uint64
	node   raft.Node

	transport *rafthttp.Transport
	server    *http.Server

	storage *storage
	commitC chan<- EntryCommit

	once            sync.Once
	shutdownTimeout time.Duration

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64
}

func NewConsensus(ctx context.Context, logger *zap.Logger, connectParams params.ConnectParameters, provider snapshotProvider, commitC chan<- EntryCommit) (*Consensus, error) {
	storage, old, err := newStorage(ctx, logger.With(zap.String("component", "storage")), connectParams.DataDir, connectParams.NodeID, provider, connectParams.SnapshotCatchupEntries)
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
		node = raft.StartNode(config, peersList(peers))
	}

	consensus := &Consensus{
		logger: logging.WithValuesFromContext(ctx, logger),

		nodeID: connectParams.NodeID,
		node:   node,

		storage: storage,
		commitC: commitC,

		shutdownTimeout: 2 * connectParams.ElectionInterval(),
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
		err := consensus.handleReady(connectParams.TickInterval)
		if err != nil {
			consensus.logger.Error("handleReady exited with error, shutting down consensus", zap.Error(err))
		}

		consensus.Stop(context.Background())
	}()

	return consensus, nil
}

func (c *Consensus) Stop(ctx context.Context) {
	c.once.Do(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), c.shutdownTimeout)
		defer cancel()
		if shutdownErr := c.server.Shutdown(shutdownCtx); shutdownErr != nil {
			c.logger.Error("failed to shutdown http server", zap.Error(shutdownErr))
		} else {
			c.logger.Info("http server shutdown complete")
		}

		c.transport.Stop()
		c.logger.Info("transport stopped")
		c.node.Stop()
		c.logger.Info("raft node stopped")
	})
}

func peersList(peers map[uint64]*url.URL) []raft.Peer {
	result := make([]raft.Peer, 0, len(peers))
	for id := range peers {
		result = append(result, raft.Peer{ID: id})
	}
	return result
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
		ErrorC:      make(chan error),
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

	c.server = &http.Server{
		Addr:    listeningAddr,
		Handler: transport.Handler(),
	}

	go func() {
		err := c.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Error("http server error", zap.Error(err))
			c.transport.ErrorC <- err
		}
	}()

	c.transport = transport
	return nil
}

// handleReady processes the Ready channel of the Raft node and applies committed entries to the state machine
func (c *Consensus) handleReady(tickInterval time.Duration) error {
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

				c.commitC <- EntryCommit{SnapshotData: rd.Snapshot.Data}

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

			err = c.publishEntries(entries)
			if err != nil {
				return stacktrace.Propagate(err, "failed to publish entries")
			}

			c.node.Advance()
		}
	}
}

// TODO implement
func (c *Consensus) publishEntries(_ []raftpb.Entry) error {
	return nil
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
