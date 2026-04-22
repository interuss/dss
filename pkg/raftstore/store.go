package raftstore

import (
	"context"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/interuss/dss/pkg/raftstore/params"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
	"github.com/jonboulle/clockwork"
)

var UnknownVersion = &semver.Version{}

// Store is a partial implementation of store.Store when the data backing is a raft store.
type Store[R any] struct {
}

// Config describes everything a raft-backed store needs to be initialized for a
// given specific package (rid, scd, aux, ...).
type Config[R any] struct {
	NewRepo func(q dsssql.Queryable, clock clockwork.Clock) R
}

func Dial[R any](ctx context.Context, connParams params.ConnectParameters) (*Store[R], error) {

	// Connect via connParams.Peers
	return &Store[R]{}, nil

}

// Init dials the database described by the global connect parameters (plus
// cfg.DBName), checks its schema version, and returns a ready-to-use Store[R].
// If withCheckCron is true, a periodic health-check cron is started.
func Init[R any](ctx context.Context, cfg Config[R], withCheckCron bool) (*Store[R], error) {

	db, err := Dial[R](ctx, params.GetConnectParameters())
	if err != nil {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return nil, stacktrace.PropagateWithCode(err, store.CodeRetryable, "Failed to connect to raft cluster")
		}
		return nil, stacktrace.Propagate(err, "Failed to connect to the raft cluster")
	}

	return db, nil
}

func (s *Store[R]) Interact(_ context.Context) (R, error) {
	var zero R
	return zero, stacktrace.NewError("Not implemented")
}

func (s *Store[R]) Transact(ctx context.Context, f func(context.Context, R) error) error {
	return stacktrace.NewError("Not implemented")
}

func (s *Store[R]) Close() error {
	return stacktrace.NewError("Not implemented")
}
