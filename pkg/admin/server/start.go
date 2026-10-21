package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/interuss/dss/pkg/admin"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	"go.uber.org/zap"
)

func Start(logger *zap.Logger, addr string, stores map[string]any) *http.Server {
	raftGroups := map[string]admin.RaftGroup{}
	for name, store := range stores {
		// Only the stores backed by a Raft group can have their membership managed.
		if group, ok := store.(admin.RaftGroup); ok {
			raftGroups[name] = group
		}
	}

	// TODO: add the authorizer when authentication is implemented for the admin API.
	router := restapi.MakeAPIRouter(&Server{Logger: logger, RaftGroups: raftGroups}, nil)
	handler := &api.MultiRouter{Routers: []api.PartialRouter{&router}}

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}

	go func() {
		logger.Info("starting admin server")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("admin server failed", zap.Error(err))
		}
	}()

	return server
}
