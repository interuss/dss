package consensus

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.uber.org/zap"
)

// listener serves the transports of all the Raft groups (aux, rid, scd) on a single port.
// Each request is routed to the transport of its group by its X-Etcd-Cluster-ID header, which carries the GroupClusterID.
type listener struct {
	logger *zap.Logger
	addr   string
	server *http.Server

	// groups maps the cluster ID of each Raft group to its registration.
	groups map[types.ID]groupRegistration
}

type groupRegistration struct {
	handler http.Handler
	errC    chan<- error
}

var (
	listenerMu sync.RWMutex
	// raftListener is started by the first Raft group to register and shut down once no group is registered.
	raftListener *listener
)

// registerGroup serves handler for the requests of the Raft group clusterID on addr.
// Server errors are sent to errC.
func registerGroup(ctx context.Context, logger *zap.Logger, addr string, clusterID types.ID, handler http.Handler, errC chan<- error) error {
	listenerMu.Lock()
	defer listenerMu.Unlock()

	if raftListener == nil {
		ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
		if err != nil {
			return stacktrace.Propagate(err, "failed to listen on %s", addr)
		}

		l := &listener{
			logger: logger.With(zap.String("address", addr)),
			addr:   addr,
			groups: map[types.ID]groupRegistration{},
		}
		l.server = &http.Server{Handler: l}
		raftListener = l

		go func() {
			err := l.server.Serve(ln)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				l.logger.Error("http server error", zap.Error(err))
				l.notifyError(err)
			}
		}()
	}

	if raftListener.addr != addr {
		return stacktrace.NewError("all Raft groups must listen on the same address: got %s, already listening on %s", addr, raftListener.addr)
	}

	if _, ok := raftListener.groups[clusterID]; ok {
		return stacktrace.NewError("a Raft group with cluster ID %s is already registered on the listener", clusterID)
	}
	raftListener.groups[clusterID] = groupRegistration{handler: handler, errC: errC}
	return nil
}

// unregisterGroup stops serving the Raft group clusterID. The listener is shut down once it serves no group.
func unregisterGroup(ctx context.Context, clusterID types.ID) error {
	server := removeGroup(clusterID)
	if server == nil {
		return nil
	}

	return server.Shutdown(ctx)
}

// removeGroup removes the Raft group clusterID from the listener.
// It returns the server of the listener to be shut down if no group is left, nil otherwise.
func removeGroup(clusterID types.ID) *http.Server {
	listenerMu.Lock()
	defer listenerMu.Unlock()

	if raftListener == nil {
		return nil
	}

	delete(raftListener.groups, clusterID)
	if len(raftListener.groups) > 0 {
		return nil
	}

	server := raftListener.server
	raftListener = nil
	return server
}

func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	listenerMu.RLock()
	handler, err := l.route(r)
	listenerMu.RUnlock()

	if err != nil {
		// %#s formats the error on a single line, without the stack.
		http.Error(w, fmt.Sprintf("%#s", err), http.StatusPreconditionFailed)
		return
	}
	handler.ServeHTTP(w, r)
}

// route returns the handler of the group the request is addressed to.
func (l *listener) route(r *http.Request) (http.Handler, error) {
	// Probing requests carry no cluster ID, any group can answer them.
	if r.URL.Path == rafthttp.ProbingPrefix {
		for _, g := range l.groups {
			return g.handler, nil
		}
		return nil, stacktrace.NewError("no Raft group is registered")
	}

	header := r.Header.Get("X-Etcd-Cluster-ID")
	clusterID, err := types.IDFromString(header)
	if err != nil {
		return nil, stacktrace.Propagate(rafthttp.ErrClusterIDMismatch, "invalid cluster ID %q", header)
	}
	g, ok := l.groups[clusterID]
	if !ok {
		return nil, stacktrace.Propagate(rafthttp.ErrClusterIDMismatch, "no Raft group with cluster ID %s", clusterID)
	}
	return g.handler, nil
}

func (l *listener) notifyError(err error) {
	listenerMu.RLock()
	defer listenerMu.RUnlock()
	for _, g := range l.groups {
		select {
		case g.errC <- err:
		default:
		}
	}
}
