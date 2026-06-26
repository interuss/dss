package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	auxraftparams "github.com/interuss/dss/pkg/aux_/store/raftstore/params"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of aux_.repos.Repository for Raft-based storage.
type repo struct{}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := auxraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get aux raft parameters")
	}
	return raftstore.Init(ctx, logger.With(zap.String("service", "aux_")), params, func() repos.Repository { return &repo{} })
}
