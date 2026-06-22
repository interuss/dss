package memstore

import (
	"context"
	"time"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/memstore"
	"github.com/interuss/stacktrace"
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

// clone returns a copy of s with independent maps and participant records.
func (s state) clone() state {
	ps := make(map[string]*participant, len(s.Participants))
	for k, v := range s.Participants {
		cp := *v
		ps[k] = &cp
	}
	hb := make(map[heartbeatKey]auxmodels.Heartbeat, len(s.Heartbeats))
	for k, v := range s.Heartbeats {
		hb[k] = v
	}
	return state{Participants: ps, Heartbeats: hb}
}

// Checkpoint returns a fast, restorable in-memory copy of the current state.
func (r *repo) Checkpoint() any {
	return r.state.clone()
}

// Restore replaces the current state with a checkpoint previously returned by
// Checkpoint. The checkpoint is copied, so it stays reusable.
func (r *repo) Restore(cp any) error {
	s, ok := cp.(state)
	if !ok {
		return stacktrace.NewError("Invalid checkpoint type %T", cp)
	}
	r.state = s.clone()
	return nil
}
