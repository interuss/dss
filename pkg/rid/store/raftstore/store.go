package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/rid/repos"
	ridraftparams "github.com/interuss/dss/pkg/rid/store/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of rid.repos.Repository for Raft-based storage.
type repo struct{}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := ridraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get rid raft parameters")
	}
	return raftstore.Init(ctx, logger, params, func() repos.Repository { return &repo{} })
}
