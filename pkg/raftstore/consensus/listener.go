package consensus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/interuss/stacktrace"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.uber.org/zap"
)

// listener serves the transports of all the Raft groups (aux, rid, scd) on a single port.
// Each request is routed to the transport of its group by its X-Etcd-Cluster-ID header, which carries the GroupClusterID.
// Its lifecycle is the one of the process (see StartListener and StopListener), independently of the groups registered on it.
type listener struct {
	logger *zap.Logger
	server *http.Server

	mu sync.RWMutex
	// handlers maps the cluster ID of each Raft group to the handler of its transport.
	handlers map[types.ID]http.Handler
}

var raftListener *listener

// StartListener starts serving the transports of the Raft groups on addr, with mutual TLS unless tlsInfo is empty.
// It must be called before any Raft group is created.
// A failure of the server after it started is reported to onError.
func StartListener(ctx context.Context, logger *zap.Logger, addr string, tlsInfo etcdtransport.TLSInfo, onError func(error)) error {
	if raftListener != nil {
		return stacktrace.NewError("the Raft listener is already started")
	}

	ln, err := newListener(ctx, addr, tlsInfo, logger)
	if err != nil {
		return stacktrace.Propagate(err, "failed to listen on %s", addr)
	}

	l := &listener{
		logger:   logger.With(zap.String("address", addr)),
		handlers: map[types.ID]http.Handler{},
	}
	l.server = &http.Server{Handler: l}
	raftListener = l

	go func() {
		err := l.server.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.logger.Error("http server error", zap.Error(err))
			onError(err)
		}
	}()

	l.logger.Info("raft listener started")
	return nil
}

// StopListener shuts down the listener started by StartListener.
func StopListener(ctx context.Context) error {
	l := raftListener
	raftListener = nil

	if l == nil {
		return nil
	}
	return l.server.Shutdown(ctx)
}

func getListener() (*listener, error) {
	if raftListener == nil {
		return nil, stacktrace.NewError("the Raft listener is not started")
	}
	return raftListener, nil
}

// register serves handler for the requests of the Raft group clusterID.
func (l *listener) register(clusterID types.ID, handler http.Handler) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.handlers[clusterID]; ok {
		return stacktrace.NewError("a Raft group with cluster ID %s is already registered on the listener", clusterID)
	}
	l.handlers[clusterID] = handler
	return nil
}

// unregister stops serving the Raft group clusterID.
func (l *listener) unregister(clusterID types.ID) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.handlers, clusterID)
}

func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, err := l.route(r)
	if err != nil {
		// %#s formats the error on a single line, without the stack.
		http.Error(w, fmt.Sprintf("%#s", err), http.StatusPreconditionFailed)
		return
	}
	handler.ServeHTTP(w, r)
}

// route returns the handler of the group the request is addressed to.
func (l *listener) route(r *http.Request) (http.Handler, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Probing requests carry no cluster ID, any group can answer them.
	if r.URL.Path == rafthttp.ProbingPrefix {
		for _, h := range l.handlers {
			return h, nil
		}
		return nil, stacktrace.NewError("no Raft group is registered")
	}

	header := r.Header.Get("X-Etcd-Cluster-ID")
	clusterID, err := types.IDFromString(header)
	if err != nil {
		return nil, stacktrace.Propagate(rafthttp.ErrClusterIDMismatch, "invalid cluster ID %q", header)
	}
	h, ok := l.handlers[clusterID]
	if !ok {
		return nil, stacktrace.Propagate(rafthttp.ErrClusterIDMismatch, "no Raft group with cluster ID %s", clusterID)
	}
	return h, nil
}
