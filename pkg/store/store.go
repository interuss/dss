package store

import (
	"context"
	"io"
)

type Store[R any] interface {
	io.Closer
	Interact(context.Context) (R, error)
	Transact(ctx context.Context, f func(context.Context, R) error) error
}
