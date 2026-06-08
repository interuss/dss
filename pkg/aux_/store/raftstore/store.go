package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/raftstore"
	"go.uber.org/zap"
)

// repo is a full implementation of aux_.repos.Repository for Raft-based storage.
type repo struct{}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	return raftstore.Init[repos.Repository](ctx, logger, func() repos.Repository { return &repo{} })
}
