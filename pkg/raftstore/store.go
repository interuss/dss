package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/raftstore/consensus"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

type Store[R any] struct {
	newRepo   func() R
	consensus *consensus.Consensus
}

func Init[R any](ctx context.Context, logger *zap.Logger, params raftparams.ConnectParameters, newRepo func() R) (*Store[R], error) {
	commitC := make(chan consensus.EntryCommit)
	consensusInstance, err := consensus.NewConsensus(ctx, logger, params, func() ([]byte, error) { return nil, nil }, commitC)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize consensus")
	}
	// TODO: start consumer goroutine reading from commitC

	return &Store[R]{
		newRepo:   newRepo,
		consensus: consensusInstance,
	}, nil
}

// Transact proposes the entry to Raft and blocks until it is committed and applied.
func (s *Store[R]) Transact(ctx context.Context, action store.Action[R]) (any, error) {
	// TODO: implement
	return nil, nil
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
