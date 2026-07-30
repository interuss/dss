package memstore

import (
	"context"
	"sync"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

type MemRepo[R any] interface {
	GetRepo() R
	GetSnapshot() ([]byte, error)
	RestoreFromSnapshot([]byte) error

	// Checkpoint ask the repo to store a quick, internal checkpoint with its current state.
	// There is at most one check point, any existing checkpoint is overwritten
	Checkpoint()

	// Restore replaces the current state with the latest checkpoint. May be called multiple time
	// to restore the same checkpoint.
	Restore()
}

// Memstore is a special kind of store:
// Store instances store data in memory. There is no persistent storage.
// Store instances are a singleton.
// Repository usage is not thread-safe.
// It's used by raftstore for projected storage.
type Store[R any] struct {
	logger *zap.Logger

	name    string
	memRepo MemRepo[R]
}

var (
	stores   = map[string]any{}
	storesMu sync.Mutex
)

func Init[R any](ctx context.Context, logger *zap.Logger, name string, r MemRepo[R]) (*Store[R], error) {

	storesMu.Lock()
	defer storesMu.Unlock()
	if s, ok := stores[name]; ok {
		return s.(*Store[R]), nil
	}

	store := &Store[R]{
		name:    name,
		logger:  logging.WithValuesFromContext(ctx, logger),
		memRepo: r,
	}

	stores[name] = store
	return store, nil
}

func (s *Store[R]) Transact(ctx context.Context, _ store.OperationRequest) (any, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "Transact not implemented for memstore")
}

func (s *Store[R]) Interact(_ context.Context) (R, error) {
	return s.memRepo.GetRepo(), nil
}

func (s *Store[R]) Checkpoint() {
	s.memRepo.Checkpoint()
}

func (s *Store[R]) Restore() {
	s.memRepo.Restore()
}

func (s *Store[R]) Close() error {
	return nil
}
