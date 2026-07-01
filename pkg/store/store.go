package store

import (
	"context"
	"io"

	"github.com/interuss/stacktrace"
)

// Action represents a set of operations to be performed on a Repo type R.
type Action[R any] interface {
	ActionMetadata
	Execute(ctx context.Context, r R) (any, error)
	Payload() any
}

type ActionMetadata interface {
	RequestType() string
	IsReadOnly() bool
}

// store.Store is the generic means to access and interact with any type of data backing the DSS
// may ever use, by obtaining a means to perform R-specific (repo type) operations.
type Store[R any] interface {
	io.Closer
	// Obtain a Repo (repo type R) that doesn't need transactional guarantees (for instance,
	// read-only).
	Interact(context.Context) (R, error)
	// Attempt to apply the operations in action to the R Repo it is supplied.  All operations performed
	// on the R Repo by action will be applied or rejected atomically.
	Transact(ctx context.Context, action Action[R]) (any, error)
}

// TransactWithResult wraps Store.Transact and returns the result of the action as a specific type, or an error if the action failed or returned a result of a different type.
// We use this to avoid having to cast the result of Transact in every handler.
func TransactWithResult[R any, Res any](ctx context.Context, s Store[R], action Action[R]) (Res, error) {
	var emptyResult Res
	result, err := s.Transact(ctx, action)
	if err != nil {
		return emptyResult, err
	}
	res, ok := result.(Res)
	if !ok {
		return emptyResult, stacktrace.NewError("unexpected result type %T, want %T", result, emptyResult)
	}
	return res, nil
}

type ActionAdapter[R any, T ActionMetadata] struct {
	Data T
	Run  func(ctx context.Context, r R, data T) (any, error)
}

func (b *ActionAdapter[R, T]) RequestType() string { return b.Data.RequestType() }

func (b *ActionAdapter[R, T]) IsReadOnly() bool { return b.Data.IsReadOnly() }

func (b *ActionAdapter[R, T]) Payload() any { return b.Data }

func (b *ActionAdapter[R, T]) Execute(ctx context.Context, r R) (any, error) {
	return b.Run(ctx, r, b.Data)
}

// TODO: This is a placeholder struct that needs to be removed once all handlers are converted into Actions
type ActionFunction[R any] struct {
	f func(context.Context, R) error
}

func NewActionFunction[R any](f func(context.Context, R) error) *ActionFunction[R] {
	return &ActionFunction[R]{f: f}
}

func (a *ActionFunction[R]) RequestType() string { return "" }

func (a *ActionFunction[R]) IsReadOnly() bool { return false }

func (a *ActionFunction[R]) Payload() any { return nil }

func (a *ActionFunction[R]) Execute(ctx context.Context, r R) (any, error) {
	return nil, a.f(ctx, r)
}

const (
	CodeRetryable = stacktrace.ErrorCode(1)
)
