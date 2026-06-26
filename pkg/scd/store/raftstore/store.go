package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/scd/repos"
	scdraftparams "github.com/interuss/dss/pkg/scd/store/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for Raft-based storage.
type repo struct{}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := scdraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get scd raft parameters")
	}
	return raftstore.Init(ctx, logger.With(zap.String("service", "scd")), params, func() repos.Repository { return &repo{} })
}
