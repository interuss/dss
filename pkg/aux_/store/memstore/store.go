package memstore

import (
	"context"
	"time"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/memstore"
	"go.uber.org/zap"
)

// repo is a full implementation of aux_.repos.Repository for memory-based storage.
type repo struct {
	// participants holds pool participants metadata, keyed by locality.
	participants map[string]*participant
	// heartbeats holds the latest heartbeat per (locality, source).
	heartbeats map[heartbeatKey]auxmodels.Heartbeat
}

type participant struct {
	publicEndpoint string
	updatedAt      time.Time
}

type heartbeatKey struct {
	locality string
	source   string
}

func newRepo() *repo {
	return &repo{
		participants: map[string]*participant{},
		heartbeats:   map[heartbeatKey]auxmodels.Heartbeat{},
	}
}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "aux_", newRepo())
}

func (r *repo) GetRepo() repos.Repository { return r }
