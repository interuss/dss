package memstore

// Memstore are a special kind of store:
// Store instances store data in memory. There is no persistent storage.
// Store instances are a singleton.
// Repository usage is not thread-safe.
//
// As of now, they are made to be used by raftstorage.
// Adaptations could be done to use them directly in the future.

import (
	"context"
	"sync"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

type MemRepo[R any] interface {
	GetRepo() R
	GetSnapshot() ([]byte, error)
	RestoreFromSnapshot([]byte) error
	Checkpoint() any
	Restore(any) error
}

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

func (s *Store[R]) Transact(ctx context.Context, requestType string, payload any, _ func(context.Context, R) error) (any, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "Transact not implemented for memstore")
}

func (s *Store[R]) Interact(_ context.Context) (R, error) {
	return s.memRepo.GetRepo(), nil
}

// Checkpoint returns a fast, restorable in-memory copy of the current state.
func (s *Store[R]) Checkpoint() any {
	return s.memRepo.Checkpoint()
}

// Restore replaces the current state with a checkpoint returned by Checkpoint.
func (s *Store[R]) Restore(cp any) error {
	return s.memRepo.Restore(cp)
}

func (s *Store[R]) Close() error {
	return nil
}
