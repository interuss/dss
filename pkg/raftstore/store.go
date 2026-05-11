package raftstore

import (
	"context"
)

type Store[R any] struct {
	newRepo func() R
}

func Init[R any](newRepo func() R) *Store[R] {
	return &Store[R]{newRepo: newRepo}
}

// Transact proposes the entry to Raft and blocks until it is committed and applied.
func (s *Store[R]) Transact(ctx context.Context, f func(context.Context, R) error) error {
	// TODO: implement
	return nil
}

// Interact returns a repository that can be used to query the store without proposing a Raft entry.
func (s *Store[R]) Interact(_ context.Context) (R, error) {
	return s.newRepo(), nil
}

// Close shuts down the consensus instance.
func (s *Store[R]) Close() error {
	// TODO: implement
	return nil
}
