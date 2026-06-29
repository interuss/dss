package memstore

import (
	"context"

	"github.com/interuss/dss/pkg/memstore"
	"github.com/interuss/dss/pkg/scd/repos"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for memory-based storage.
type repo struct{}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "scd", &repo{})
}

func (r *repo) GetRepo() repos.Repository { return r }
