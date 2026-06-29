package store

import (
	"context"
	"io"

	"github.com/interuss/stacktrace"
)

// Action represents a set of operations to be performed on a Repo type R.
type Action[R any] interface {
	Execute(ctx context.Context, r R) (any, error)
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

// TODO: This is a placeholder struct that needs to be removed once all handlers are converted into Actions
type ActionFunction[R any] struct {
	f func(context.Context, R) error
}

func NewActionFunction[R any](f func(context.Context, R) error) *ActionFunction[R] {
	return &ActionFunction[R]{f: f}
}

func (a *ActionFunction[R]) RequestType() string { return "" }

func (a *ActionFunction[R]) IsReadOnly() bool { return false }

func (a *ActionFunction[R]) Execute(ctx context.Context, r R) (any, error) {
	return nil, a.f(ctx, r)
}

const (
	CodeRetryable = stacktrace.ErrorCode(1)
)
