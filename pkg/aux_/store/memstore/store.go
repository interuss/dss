package memstore

import (
	"context"
	"time"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/memstore"
	"go.uber.org/zap"
)

type locality string

// repo is a full implementation of aux_.repos.Repository for memory-based storage.
type repo struct {
	state state
}

// state is the serializable in-memory state.
type state struct {
	// Participants holds pool participants metadata, keyed by locality.
	Participants map[locality]*participant
	// Heartbeats holds the latest heartbeat per (locality, source).
	Heartbeats map[heartbeatKey]auxmodels.Heartbeat
}

type participant struct {
	PublicEndpoint string
	UpdatedAt      time.Time
}

type heartbeatKey struct {
	Locality locality
	Source   string
}

func newRepo() *repo {
	return &repo{
		state: state{
			Participants: map[locality]*participant{},
			Heartbeats:   map[heartbeatKey]auxmodels.Heartbeat{},
		}}
}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "aux_", newRepo())
}

func (r *repo) GetRepo() repos.Repository { return r }
