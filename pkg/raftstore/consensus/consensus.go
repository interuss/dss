package consensus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	v2stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

const (
	defaultClusterID uint64 = 1
)

type Consensus struct {
	logger *zap.Logger

	node raft.Node

	transport *rafthttp.Transport
	server    *http.Server

	storage *storage
	errorC  chan error
}

func NewConsensus(ctx context.Context, logger *zap.Logger, nodeID uint64, peers map[uint64]*url.URL, dataDir string, snapshotCatchupEntries uint64) (*Consensus, error) {
	storage, _, err := newStorage(ctx, logger.With(zap.String("component", "storage")), dataDir, nodeID, snapshotCatchupEntries)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize storage")
	}

	consensus := &Consensus{
		logger: logging.WithValuesFromContext(ctx, logger),

		storage: storage,
		errorC:  make(chan error, 1),
	}

	err = consensus.initTransport(ctx, nodeID, defaultClusterID, peers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize transport")
	}

	return consensus, nil
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
		ErrorC:      c.errorC,
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
			c.errorC <- err
		}
	}()

	c.transport = transport
	return nil
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
func (c *Consensus) RegisterStore(name string, provider snapshotProvider) {
	c.storage.registerSnapshotProvider(name, provider)
}
