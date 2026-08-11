package store

import (
	"context"
	"encoding/json"
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
	Encode     func(req OperationRequest) ([]byte, error)
	Decode     func(buf []byte) (OperationRequest, error)
	Execute    func(ctx context.Context, repo R, request OperationRequest) (any, error)
	IsReadOnly bool
}

// EncodeJSON is a general-purpose OperationHandler.Encode that marshals the request as JSON.
// TODO: remove once all operations use custom encoders / decoders.
func EncodeJSON(request OperationRequest) ([]byte, error) {
	buf, err := json.Marshal(request)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal %q request", request.OperationID())
	}
	return buf, nil
}

// DecodeJSON is a general-purpose OperationHandler.Decode that unmarshals JSON into a new T.
// T must be a pointer to a struct type that implements OperationRequest.
// TODO: remove once all operations use custom encoders / decoders.
func DecodeJSON[T OperationRequest](buf []byte) (OperationRequest, error) {
	req := new(T)
	if err := json.Unmarshal(buf, req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal request")
	}
	return *req, nil
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
