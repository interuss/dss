package store

import (
	"context"
	"io"

	"github.com/interuss/stacktrace"
)

// Request carries the information needed by the Raftstore to handle a request.
type Request struct {
	RequestType string
	Payload     any
}

// store.Store is the generic means to access and interact with any type of data backing the DSS
// may ever use, by obtaining a means to perform R-specific (repo type) operations.
type Store[R any] interface {
	io.Closer
	// Obtain a Repo (repo type R) that doesn't need transactional guarantees (for instance,
	// read-only).
	Interact(context.Context) (R, error)
	// Transact atomically applies or rejects a request.
	// f specifies the operations to be performed on the repo.
	// request specifies the type and payload of request.
	// any is the request result.
	Transact(ctx context.Context, request Request, f func(context.Context, R) error) (any, error)
}

const (
	CodeRetryable = stacktrace.ErrorCode(1)
)
