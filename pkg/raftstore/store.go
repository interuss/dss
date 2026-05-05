package raftstore

import (
	"context"
)

type Store[R any] struct{}

func Init[R any]() *Store[R] {
	return &Store[R]{}
}

// Transact proposes the entry to Raft and blocks until it is committed and applied.
func (s *Store[R]) Transact(ctx context.Context, f func(context.Context, R) error) error {
	panic("not implemented")
}

// Interact returns a repository that can be used to query the store without proposing a Raft entry.
func (s *Store[R]) Interact(_ context.Context) (R, error) {
	panic("not implemented")
}

// Close shuts down the consensus instance.
func (s *Store[R]) Close() error {
	panic("not implemented")
}
