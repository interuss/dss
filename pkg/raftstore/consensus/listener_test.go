package consensus

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	l := &listener{groups: map[types.ID]groupRegistration{
		types.ID(1): {handler: namedHandler("rid")},
		types.ID(2): {handler: namedHandler("scd")},
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

func TestRegisterGroup(t *testing.T) {
	logger := zap.NewNop()
	errC := make(chan error, 1)
	addr := "127.0.0.1:0"

	require.NoError(t, registerGroup(t.Context(), logger, addr, types.ID(1), namedHandler("rid"), errC))
	require.NoError(t, registerGroup(t.Context(), logger, addr, types.ID(2), namedHandler("scd"), errC))
	listener := raftListener

	require.Error(t, registerGroup(t.Context(), logger, addr, types.ID(1), namedHandler("rid"), errC), "duplicate cluster ID")
	require.Error(t, registerGroup(t.Context(), logger, "127.0.0.1:1", types.ID(3), namedHandler("aux"), errC), "different address")

	require.NoError(t, unregisterGroup(t.Context(), types.ID(1)))
	assert.Same(t, listener, raftListener, "the listener is kept while a group is registered")

	require.NoError(t, unregisterGroup(t.Context(), types.ID(2)))
	assert.Nil(t, raftListener, "the listener is shut down once no group is registered")
}
