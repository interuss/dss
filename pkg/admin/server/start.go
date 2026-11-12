package server

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// Start serves the admin API on the Unix socket at socketPath, with requests failing after timeout and the context of
// the handlers expiring after handlerTimeout.
// The directory of the socket is created if needed. A leftover socket at socketPath is replaced.
func Start(ctx context.Context, logger *zap.Logger, socketPath string, timeout time.Duration, handlerTimeout time.Duration, raftGroups map[string]*consensus.Consensus) (*http.Server, error) {
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0o700); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create the directory of the admin socket at %s", socketDir)
	}

	if err := os.Remove(socketPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, stacktrace.Propagate(err, "failed to remove the previous admin socket at %s", socketPath)
	}
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "unix", socketPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to listen on the admin socket at %s", socketPath)
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		_ = listener.Close()
		return nil, stacktrace.Propagate(err, "failed to restrict the permissions of the admin socket at %s", socketPath)
	}

	multiRouter := &api.MultiRouter{}
	multiRouter.SetRouter("admin", restapi.MakeAPIRouter(&Server{Logger: logger, raftGroups: raftGroups}, nil))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), handlerTimeout)
		defer cancel()
		multiRouter.ServeHTTP(w, r.WithContext(ctx))
	})

	server := &http.Server{
		Handler:           http.TimeoutHandler(handler, timeout, `{"message":"request timeout"}`),
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}

	go func() {
		logger.Info("starting admin server")
		// The socket file is removed when the server is shut down, which closes the listener.
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("admin server failed", zap.Error(err))
		}
	}()

	return server, nil
}
