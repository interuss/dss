package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

type RequestType = string

type RaftRepo[R any] interface {
	GetRepo() R
	// Apply is called on every committed entry. The proposal must be applied atomically.
	Apply(ctx context.Context, proposal consensus.Proposal) (any, error)
	GetSnapshot() ([]byte, error)
	RestoreFromSnapshot(data []byte) error
}

type Store[R any] struct {
	logger *zap.Logger

	name     string
	raftRepo RaftRepo[R]
	cancel   context.CancelFunc

	Consensus *consensus.Consensus
}

func Init[R any](ctx context.Context, logger *zap.Logger, locality string, params raftparams.ConnectParameters, r RaftRepo[R]) (*Store[R], error) {
	ctx, cancel := context.WithCancel(ctx)

	store := &Store[R]{
		raftRepo: r,
		logger:   logging.WithValuesFromContext(ctx, logger),
		cancel:   cancel,
	}
	commitC := make(chan consensus.EntryCommit)
	go store.processCommits(ctx, commitC)

	consensusInstance, err := consensus.NewConsensus(ctx, logger, locality, params, r.GetSnapshot, commitC)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize consensus")
	}

	store.Consensus = consensusInstance

	return store, nil
}

// Transact passes the action to the consensus layer and blocks until it is committed and applied.
// The processCommits loop will call Apply on the proposal when it is committed.
func (s *Store[R]) Transact(ctx context.Context, action store.Action[R]) (any, error) {
	return s.Consensus.HandleClientRequest(ctx, action.RequestType(), action.Payload(), action.IsReadOnly())
}

func (s *Store[R]) Interact(_ context.Context) (R, error) {
	return s.raftRepo.GetRepo(), nil
}

// Close shuts down the consensus instance and processCommits loop.
func (s *Store[R]) Close() error {
	s.Consensus.Stop(context.Background())
	s.cancel()
	return nil
}

// processCommits reads committed entries from the consensus layer and applies them via Apply.
func (s *Store[R]) processCommits(ctx context.Context, commitCh <-chan consensus.EntryCommit) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("stopping commit processing loop")
			return
		case commit, ok := <-commitCh:
			if !ok {
				s.logger.Info("commit channel closed, stopping commit processing loop")
				return
			}

			if commit.SnapshotData != nil {
				if err := s.raftRepo.RestoreFromSnapshot(commit.SnapshotData); err != nil {
					s.logger.Error("failed to restore from snapshot", zap.Error(err))
				}
				continue
			}

			ctx = timestamp.WithRequestTimestamp(ctx, commit.Prop.Timestamp)
			result, err := s.raftRepo.Apply(ctx, commit.Prop)
			commit.Done <- consensus.ProposalResult{Result: result, Error: err}
		}
	}
}
