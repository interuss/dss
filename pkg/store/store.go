package store

import (
	"context"
	"io"

	"github.com/coreos/go-semver/semver"
)

type Store[R any] interface {
	io.Closer
	Interact(context.Context) (R, error)
	Transact(ctx context.Context, f func(context.Context, R) error) error
	GetVersion(ctx context.Context) (*semver.Version, error)
}
