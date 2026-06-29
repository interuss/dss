package memstore

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/memstore"
	"go.uber.org/zap"
)

// repo is a full implementation of aux_.repos.Repository for memory-based storage.
type repo struct{}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "aux_", &repo{})
}

func (r *repo) GetRepo() repos.Repository { return r }
