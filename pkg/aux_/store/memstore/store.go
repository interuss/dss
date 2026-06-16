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
	state state
}

// state is the serializable in-memory state.
type state struct {
	// Participants holds pool participants metadata, keyed by locality.
	Participants map[string]*participant
	// Heartbeats holds the latest heartbeat per (locality, source).
	Heartbeats map[heartbeatKey]auxmodels.Heartbeat
	// participants holds pool participants metadata, keyed by locality.
	participants map[string]*participant
	// heartbeats holds the latest heartbeat per (locality, source).
	heartbeats map[heartbeatKey]auxmodels.Heartbeat
}

type participant struct {
	PublicEndpoint string
	UpdatedAt      time.Time
}

type heartbeatKey struct {
	Locality string
	Source   string
}

func newRepo() *repo {
	return &repo{
		state: state{
			Participants: map[string]*participant{},
			Heartbeats:   map[heartbeatKey]auxmodels.Heartbeat{},
		}}
}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "aux_", newRepo())
}

func (r *repo) GetRepo() repos.Repository { return r }
