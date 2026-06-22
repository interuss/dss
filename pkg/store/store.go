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
	// Attempt to apply the operations in f to the R Repo it is supplied.  All operations performed
	// on the R Repo by f will be applied or rejected atomically.
	// requestType and payload are used by the Raftstore to build the proposal.
	// The returned any is the proposal result (also Raftstore only).
	Transact(ctx context.Context, requestType string, payload any, f func(context.Context, R) error) (any, error)
}

const (
	CodeRetryable = stacktrace.ErrorCode(1)
)
