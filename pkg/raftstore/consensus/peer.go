package consensus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.uber.org/zap"
)

// nodePath is the path at which a node describes itself in a group on the Raft listener.
// A member reads it to learn the ID of a node before adding it.
var nodePath = path.Join(rafthttp.RaftPrefix, "node")

// peerEndpoint is the HTTP interface through which the nodes of a group describe themselves to each other.
type peerEndpoint struct {
	logger    *zap.Logger
	clusterID types.ID
	// describe returns the description of this node in the group.
	describe func() (Node, error)
	client   *http.Client
}

func newPeerEndpoint(logger *zap.Logger, clusterID types.ID, tlsInfo etcdtransport.TLSInfo, describe func() (Node, error)) (*peerEndpoint, error) {
	roundTripper, err := rafthttp.NewRoundTripper(tlsInfo, 0)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create the client to other nodes")
	}
	return &peerEndpoint{
		logger:    logger,
		clusterID: clusterID,
		describe:  describe,
		client:    &http.Client{Transport: roundTripper},
	}, nil
}

// handler serves the description of this node, and the Raft transport behind raftHandler on every other path.
func (e *peerEndpoint) handler(raftHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+nodePath, e.serve)
	mux.Handle("/", raftHandler)
	return mux
}

// serve describes this node in the Raft group.
func (e *peerEndpoint) serve(w http.ResponseWriter, r *http.Request) {
	node, err := e.describe()
	if err != nil {
		http.Error(w, "failed to describe this node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(node); err != nil {
		e.logger.Warn("failed to send the description of this node", zap.Error(err))
	}
}

// fetch reads the description in this Raft group of the node at peerURL, from its Raft listener.
func (e *peerEndpoint) fetch(ctx context.Context, peerURL string) (Node, error) {
	if _, err := types.NewURLs([]string{peerURL}); err != nil {
		return Node{}, stacktrace.NewErrorWithCode(dsserr.BadRequest, "invalid URL %q: %v", peerURL, err)
	}

	nodeURL, err := url.JoinPath(peerURL, nodePath)
	if err != nil {
		return Node{}, stacktrace.NewErrorWithCode(dsserr.BadRequest, "invalid URL %q: %v", peerURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nodeURL, nil)
	if err != nil {
		return Node{}, stacktrace.Propagate(err, "failed to create the request to %s", peerURL)
	}
	req.Header.Set("X-Etcd-Cluster-ID", e.clusterID.String())

	resp, err := e.client.Do(req)
	if err != nil {
		return Node{}, stacktrace.PropagateWithCode(err, dsserr.Unavailable, "failed to reach the node at %s", peerURL)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusPreconditionFailed:
		return Node{}, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "the node at %s does not run this Raft group in this cluster", peerURL)
	default:
		return Node{}, stacktrace.NewErrorWithCode(dsserr.Unavailable, "unexpected status %d from the node at %s", resp.StatusCode, peerURL)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return Node{}, stacktrace.PropagateWithCode(err, dsserr.Unavailable, "invalid response from the node at %s", peerURL)
	}
	return node, nil
}

// close releases the connections of the client to the other nodes.
func (e *peerEndpoint) close() {
	e.client.CloseIdleConnections()
}
