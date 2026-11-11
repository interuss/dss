package store

import (
	"context"

	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/stacktrace"
)

// store.Store is the generic means to access and interact with any type of data backing the DSS
// may ever use, by obtaining a means to perform R-specific (repo type) operations.
type Store[R any] interface {
	Close(ctx context.Context) error
	// Obtain a Repo (repo type R) that doesn't need transactional guarantees (for instance,
	// read-only).
	Interact(context.Context) (R, error)
	// Apply the operation request to the R Repo atomically.
	Transact(ctx context.Context, request OperationRequest) (any, error)
	// Ready reports whether the store can serve requests.
	Ready() bool
}

// OperationRequest identifies an operation to be performed on a Store.
type OperationRequest interface {
	OperationID() string
}

// OperationHandler holds the logic for the encoding and execution of a registered operation
type OperationHandler[R any] struct {
	EncodePayload func(req OperationRequest) (*raftpb.Payload, error)
	DecodePayload func(payload *raftpb.Payload) (OperationRequest, error)
	Execute       func(ctx context.Context, repo R, request OperationRequest) (any, error)
	IsReadOnly    bool
}

// ExecutePayload decodes the payload with the handler registered for operationID and executes the resulting request.
func ExecutePayload[R any](ctx context.Context, registry map[string]OperationHandler[R], operationID string, repo R, payload *raftpb.Payload) (any, error) {
	handler, ok := registry[operationID]
	if !ok {
		return nil, stacktrace.NewError("no handler registered for operation %q", operationID)
	}

	request, err := handler.DecodePayload(payload)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode op %q", operationID)
	}

	return handler.Execute(ctx, repo, request)
}

// TransactWithResult wraps Store.Transact and casts the result to ResultType, avoiding a cast at every call site.
func TransactWithResult[R any, ResultType any](ctx context.Context, store Store[R], request OperationRequest) (ResultType, error) {
	var empty ResultType
	transactionResult, err := store.Transact(ctx, request)
	if err != nil {
		return empty, err
	}
	if resultType, ok := transactionResult.(ResultType); ok {
		return resultType, nil
	}
	return empty, stacktrace.NewError("unexpected result type %T, want %T", transactionResult, empty)
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
