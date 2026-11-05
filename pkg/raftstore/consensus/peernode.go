package consensus

import (
	"context"
	"encoding/json"
	"net/http"
	"path"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.uber.org/zap"
)

// nodePath is the path at which a node describes itself in a group on the Raft listener.
// A member reads it to learn the ID of a node before adding it.
var nodePath = path.Join(rafthttp.RaftPrefix, "node")

// serveNode describes this node in the Raft group.
func (c *Consensus) serveNode(w http.ResponseWriter, r *http.Request) {
	node, err := c.Node()
	if err != nil {
		http.Error(w, "failed to describe this node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(node); err != nil {
		c.logger.Warn("failed to send the description of this node", zap.Error(err))
	}
}

// fetchNode reads the description in this Raft group of the node at peerURL, from its Raft listener.
func (c *Consensus) fetchNode(ctx context.Context, peerURL string) (Node, error) {
	if _, err := types.NewURLs([]string{peerURL}); err != nil {
		return Node{}, stacktrace.NewErrorWithCode(dsserr.BadRequest, "invalid URL %q: %v", peerURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, peerURL+nodePath, nil)
	if err != nil {
		return Node{}, stacktrace.Propagate(err, "failed to create the request to %s", peerURL)
	}
	req.Header.Set("X-Etcd-Cluster-ID", c.transport.ClusterID.String())

	resp, err := c.peerClient.Do(req)
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
