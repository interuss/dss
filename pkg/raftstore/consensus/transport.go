package consensus

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	"github.com/interuss/stacktrace"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	v2stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

// peerTransport bundles the rafthttp transport, peer description endpoint, and listener registration of a Raft group.
type peerTransport struct {
	logger    *zap.Logger
	transport *rafthttp.Transport
	endpoint  *peerEndpoint
	listener  *listener
	stopOnce  sync.Once
}

func newPeerTransport(
	ctx context.Context,
	logger *zap.Logger,
	nodeID uint64,
	clusterID uint64,
	nodeURL *url.URL,
	tlsInfo etcdtransport.TLSInfo,
	raftReceiver rafthttp.Raft,
	describe func() (node.Info, error),
) (*peerTransport, error) {
	nodeIDStr := fmt.Sprintf("%d", nodeID)
	if tlsInfo.Empty() {
		logger.Warn("--raft_insecure is set: Raft nodes do not authenticate each other and connection is not encrypted. Don't use this node in production environments.")
	}

	transport := &rafthttp.Transport{
		Logger: logging.WithValuesFromContext(ctx, logger.With(zap.String("component", "transport"))),
		ID:     types.ID(nodeID),
		// URLs are advertised to the peers this node sends messages to, which lets a node joining the cluster
		// reply to the leader before it learns the members of the cluster.
		URLs:        types.URLs{*nodeURL},
		ClusterID:   types.ID(clusterID),
		Raft:        raftReceiver,
		ServerStats: v2stats.NewServerStats(nodeIDStr, nodeIDStr),
		LeaderStats: v2stats.NewLeaderStats(logger, nodeIDStr),
		// ErrorC is buffered to make sure rafthttp can send errors even if the raft updates consumer is not listening yet.
		// The buffer size of 1 is sufficient because the transport will only send one error before stopping.
		ErrorC:  make(chan error, 1),
		TLSInfo: tlsInfo,
	}

	if err := transport.Start(); err != nil {
		return nil, stacktrace.Propagate(err, "failed to start transport")
	}

	endpoint, err := newPeerEndpoint(logger, transport.ClusterID, tlsInfo, describe)
	if err != nil {
		transport.Stop()
		return nil, err
	}

	l, err := getListener()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get the listener")
	}
	if err := l.register(transport.ClusterID, endpoint.handler(transport.Handler())); err != nil {
		transport.Stop()
		endpoint.close()
		return nil, stacktrace.Propagate(err, "failed to register the transport to the listener")
	}

	return &peerTransport{
		logger:    logger,
		transport: transport,
		endpoint:  endpoint,
		listener:  l,
	}, nil
}

func (t *peerTransport) send(msgs []*raftpb.Message) {
	t.transport.Send(msgs)
}

func (t *peerTransport) errorC() <-chan error {
	return t.transport.ErrorC
}

func (t *peerTransport) fetchPeer(ctx context.Context, peerURL string) (node.Info, error) {
	return t.endpoint.fetch(ctx, peerURL)
}

func (t *peerTransport) clusterID() uint64 {
	return uint64(t.transport.ClusterID)
}

func (t *peerTransport) leaderStats() *v2stats.LeaderStats {
	return t.transport.LeaderStats
}

func (t *peerTransport) raw() *rafthttp.Transport {
	return t.transport
}

func (t *peerTransport) stopPeers() {
	t.stopOnce.Do(func() {
		t.transport.Stop()
		t.logger.Info("transport stopped")
	})
}

func (t *peerTransport) close() {
	t.listener.unregister(t.transport.ClusterID)
	t.logger.Info("unregistered from listener")
	t.stopPeers()
	t.endpoint.close()
}
