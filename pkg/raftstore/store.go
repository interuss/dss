package raftstore

import (
	"context"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

type Store[R any] struct{}

func Init[R any]() *Store[R] {
	return &Store[R]{}
}

// Transact proposes the entry to Raft and blocks until it is committed and applied.
func (s *Store[R]) Transact(ctx context.Context, f func(context.Context, R) error) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "Transact not yet implemented for Raft store")
}

// Interact returns a repository that can be used to query the store without proposing a Raft entry.
func (s *Store[R]) Interact(_ context.Context) (R, error) {
	var empty R
	return empty, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "Interact not yet implemented for Raft store")
}

// Close shuts down the consensus instance.
func (s *Store[R]) Close() error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "Close not yet implemented for Raft store")
}
