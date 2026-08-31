package memstore

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/memstore"
	"go.uber.org/zap"
)

type locality string

// repo is a full implementation of aux_.repos.Repository for memory-based storage.
type repo struct {
	state      state
	checkpoint state
}

// state is the serializable in-memory state.
type state struct {
	// Participants holds pool participants metadata, keyed by locality.
	Participants map[locality]*participant
	// Heartbeats holds the latest heartbeat per (locality, source).
	Heartbeats map[heartbeatKey]*heartbeat
}

type participant struct {
	PublicEndpoint string
	UpdatedAt      time.Time
}

type heartbeatKey struct {
	Locality locality
	Source   string
}

type heartbeat struct {
	Timestamp                   *time.Time
	NextHeartbeatExpectedBefore *time.Time
	Reporter                    string
}

func newRepo() *repo {

	state := state{
		Participants: map[locality]*participant{},
		Heartbeats:   map[heartbeatKey]*heartbeat{},
	}

	return &repo{
		state:      state,
		checkpoint: state.clone(),
	}
}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "aux_", newRepo())
}

func (r *repo) GetRepo() repos.Repository { return r }

func clonePtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	return new(*v)
}

func (h *heartbeat) clone() *heartbeat {
	cp := *h
	cp.Timestamp = clonePtr(h.Timestamp)
	cp.NextHeartbeatExpectedBefore = clonePtr(h.NextHeartbeatExpectedBefore)
	return &cp
}

func (p *participant) clone() *participant {
	cp := *p
	return &cp
}

// clone returns a copy of s with independent maps and participant records.
func (s state) clone() state {
	ps := make(map[locality]*participant, len(s.Participants))
	for k, v := range s.Participants {
		ps[k] = v.clone()
	}
	hb := make(map[heartbeatKey]*heartbeat, len(s.Heartbeats))
	for k, v := range s.Heartbeats {
		hb[k] = v.clone()
	}
	return state{Participants: ps, Heartbeats: hb}
}

// Checkpoint ask the repo to store a quick, internal checkpoint with its current state.
// There is at most one check point, any existing checkpoint is overwritten
func (r *repo) Checkpoint() {
	r.checkpoint = r.state.clone()
}

// Restore replaces the current state with the latest checkpoint. May be called multiple time
// to restore the same checkpoint.
func (r *repo) Restore() {
	r.state = r.checkpoint.clone()
}
