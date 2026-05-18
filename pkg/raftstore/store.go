package raftstore

import (
	"context"
	"sync"

	"github.com/interuss/dss/pkg/raftstore/consensus"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

var (
	sharedConsensus     *consensus.Consensus
	sharedConsensusOnce sync.Once
	sharedConsensusErr  error
)

type Store[R any] struct {
	newRepo   func() R
	consensus *consensus.Consensus
}

func Init[R any](ctx context.Context, logger *zap.Logger, newRepo func() R) (*Store[R], error) {
	// scd, rid and aux will share the same consensus instance, so we initialize it once.
	sharedConsensusOnce.Do(func() {
		params := raftparams.GetConnectParameters()
		peers, err := params.PeerMap()
		if err != nil {
			sharedConsensusErr = stacktrace.Propagate(err, "failed to parse peer map")
			return
		}

		sharedConsensus, sharedConsensusErr = consensus.NewConsensus(ctx, logger, params.ID, peers, params.DataDir, params.SnapshotCatchupEntries)
		if sharedConsensusErr != nil {
			sharedConsensusErr = stacktrace.Propagate(sharedConsensusErr, "failed to initialize consensus")
		}
	})
	if sharedConsensusErr != nil {
		return nil, sharedConsensusErr
	}

	// TODO: implement
	sharedConsensus.RegisterStore("provider", func() ([]byte, error) {
		return nil, nil
	})

	return &Store[R]{
		newRepo:   newRepo,
		consensus: sharedConsensus,
	}, nil
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
