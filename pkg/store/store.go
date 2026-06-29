package store

import (
	"context"
	"io"

	"github.com/interuss/stacktrace"
)

// store.Store is the generic means to access and interact with any type of data backing the DSS
// may ever use, by obtaining a means to perform R-specific (repo type) operations.
type Store[R any] interface {
	io.Closer
	// Obtain a Repo (repo type R) that doesn't need transactional guarantees (for instance,
	// read-only).
	Interact(context.Context) (R, error)
	// Apply the operation request to the R Repo atomically.
	Transact(ctx context.Context, request OperationRequest) (any, error)
}

// OperationRequest identifies an operation to be performed on a Store.
type OperationRequest interface {
	OperationID() string
}

// OperationHandler holds the logic for the encoding and execution of a registered operation
type OperationHandler[R any] struct {
	Encode  func(req OperationRequest) ([]byte, error)
	Decode  func(buf []byte) (OperationRequest, error)
	Execute func(ctx context.Context, repo R, request OperationRequest) (any, error)
}

// TransactWithResult wraps Store.Transact and casts the result to ResultType, avoiding a cast at every call site.
func TransactWithResult[R any, ResultType any](ctx context.Context, store Store[R], request OperationRequest) (ResultType, error) {
	var empty ResultType
	transactionResult, err := store.Transact(ctx, request)
	if err != nil {
		return empty, err
	}
	resultType, ok := transactionResult.(ResultType)
	if !ok {
		return empty, stacktrace.NewError("unexpected result type %T, want %T", transactionResult, empty)
	}
	return resultType, nil
}

// FuncOperation wraps a closure as an OperationRequest for gradual migration.
// TODO: remove once all handlers use registered operations.
type FuncOperation[R any] struct {
	f func(context.Context, R) error
}

func NewFuncOperation[R any](f func(context.Context, R) error) *FuncOperation[R] {
	return &FuncOperation[R]{f: f}
}

func (a *FuncOperation[R]) OperationID() string                    { return "" }
func (a *FuncOperation[R]) Execute(ctx context.Context, r R) error { return a.f(ctx, r) }

const (
	CodeRetryable = stacktrace.ErrorCode(1)
)
