package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/locality"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

type RaftRepo[R any] interface {
	GetRepo() R
	// Apply is called on every committed entry. The proposal must be applied atomically.
	// The any return mirrors store.OperationHandler.Execute: different requests yield different
	// concrete result types. Callers recover the type via store.TransactWithResult.
	Apply(ctx context.Context, proposal consensus.Proposal) (any, error)

	// GetSnapshot returns a serialized view of current state, suitable
	// for restoring via RestoreFromSnapshot.
	GetSnapshot() ([]byte, error)

	// RestoreFromSnapshot replaces all state with the snapshot in data.
	// data is always the output of a prior GetSnapshot.
	RestoreFromSnapshot(data []byte) error

	// Checkpoint saves the current state, called before every proposal is applied.
	Checkpoint()

	// Restore reverts the state to the last Checkpoint, called when Apply returns an error so a
	// failed proposal cannot leave a partial mutation in place.
	Restore()
}

type Store[R any] struct {
	logger *zap.Logger

	raftRepo RaftRepo[R]
	cancel   context.CancelFunc
	registry map[string]store.OperationHandler[R]

	Consensus *consensus.Consensus

	done chan struct{}
}

func Init[R any](ctx context.Context, logger *zap.Logger, locality string, params raftparams.ConnectParameters, r RaftRepo[R], registry map[string]store.OperationHandler[R]) (*Store[R], error) {
	ctx, cancel := context.WithCancel(ctx)

	store := &Store[R]{
		raftRepo: r,
		logger:   logging.WithValuesFromContext(ctx, logger),
		cancel:   cancel,
		registry: registry,
		done:     make(chan struct{}),
	}
	commitC := make(chan consensus.EntryCommit)
	go func() {
		defer close(store.done)
		store.processCommits(ctx, commitC)
	}()

	consensusInstance, err := consensus.NewConsensus(ctx, logger, locality, params, r.GetSnapshot, commitC)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize consensus")
	}

	store.Consensus = consensusInstance

	return store, nil
}

// Transact proposes the entry to Raft and blocks until it is committed and applied.
func (s *Store[R]) Transact(ctx context.Context, request store.OperationRequest) (any, error) {
	handler, ok := s.registry[request.OperationID()]
	if !ok {
		return nil, stacktrace.NewError("no handler registered for operation %q", request.OperationID())
	}
	payload, err := handler.Encode(request)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode op %q", request.OperationID())
	}
	return s.Consensus.HandleClientRequest(ctx, consensus.RequestType(request.OperationID()), payload, handler.IsReadOnly)
}

// Interact returns the underlying Raft repo which, for every operation, will propose it to Raft and return the results.
func (s *Store[R]) Interact(_ context.Context) (R, error) {
	return s.raftRepo.GetRepo(), nil
}

// Close shuts down the consensus instance and processCommits loop.
// TODO: pass a context to Stop then to consensus.Stop (see issue: https://github.com/interuss/dss/issues/1610).
func (s *Store[R]) Close() error {
	s.Consensus.Stop(context.Background())
	s.cancel()
	s.logger.Info("waiting for commit processing goroutine to exit")
	<-s.done
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
					s.logger.Fatal("failed to restore from snapshot", zap.Error(err))
				}
				continue
			}

			proposalCtx := timestamp.NewContext(ctx, commit.Prop.Timestamp)
			proposalCtx = locality.NewContext(proposalCtx, commit.Prop.Locality)
			s.raftRepo.Checkpoint()
			result, err := s.raftRepo.Apply(proposalCtx, commit.Prop)
			if err != nil {
				s.raftRepo.Restore()
			}
			commit.Done <- consensus.ProposalResult{Result: result, Error: err}
		}
	}
}
