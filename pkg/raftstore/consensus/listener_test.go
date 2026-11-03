package consensus

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.uber.org/zap"
)

func namedHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, name)
	})
}

func TestListenerRoutesByClusterID(t *testing.T) {
	l := &listener{handlers: map[types.ID]http.Handler{
		types.ID(1): namedHandler("rid"),
		types.ID(2): namedHandler("scd"),
	}}

	tests := []struct {
		name       string
		path       string
		clusterID  string
		wantStatus int
		wantBody   string
	}{
		{name: "routed to rid", path: rafthttp.RaftPrefix, clusterID: types.ID(1).String(), wantStatus: http.StatusOK, wantBody: "rid"},
		{name: "routed to scd", path: rafthttp.RaftStreamPrefix + "/message/1", clusterID: types.ID(2).String(), wantStatus: http.StatusOK, wantBody: "scd"},
		{name: "unknown cluster ID", path: rafthttp.RaftPrefix, clusterID: types.ID(3).String(), wantStatus: http.StatusPreconditionFailed},
		{name: "missing cluster ID", path: rafthttp.RaftPrefix, wantStatus: http.StatusPreconditionFailed},
		{name: "probing without cluster ID", path: rafthttp.ProbingPrefix, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, tt.path, nil)
			if tt.clusterID != "" {
				req.Header.Set("X-Etcd-Cluster-ID", tt.clusterID)
			}
			rec := httptest.NewRecorder()

			l.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestListenerLifecycle(t *testing.T) {
	_, err := getListener()
	require.Error(t, err, "not started")

	require.NoError(t, StartListener(t.Context(), zap.NewNop(), "127.0.0.1:0", etcdtransport.TLSInfo{}, func(error) {}))
	require.Error(t, StartListener(t.Context(), zap.NewNop(), "127.0.0.1:0", etcdtransport.TLSInfo{}, func(error) {}), "already started")

	l, err := getListener()
	require.NoError(t, err)

	require.NoError(t, l.register(types.ID(1), namedHandler("rid")))
	require.NoError(t, l.register(types.ID(2), namedHandler("scd")))
	require.Error(t, l.register(types.ID(1), namedHandler("rid")), "duplicate cluster ID")

	l.unregister(types.ID(1))
	l.unregister(types.ID(2))
	same, err := getListener()
	require.NoError(t, err)
	assert.Same(t, l, same, "the listener is kept when no group is registered")

	require.NoError(t, StopListener(t.Context()))
	_, err = getListener()
	require.Error(t, err, "stopped")
}
