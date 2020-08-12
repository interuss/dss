package store

import (
	"context"

	"github.com/coreos/go-semver/semver"
	"github.com/interuss/dss/pkg/rid/repos"
)

// Store provides the means by which to obtain Repos with which to interact with
// the remote ID backing store.
type Store interface {
	Interactor
	Transactor

	// Close closes the store and releases all of its resources.
	Close() error

	// Get store version
	GetVersion(ctx context.Context) (*semver.Version, error)
}

// Interactor provides means to get hold of a repos.Repository instance *without* any
// isolation/atomicity guarantees.
type Interactor interface {
	// Interact returns a repos.Repository instance or an error in case of issues.
	Interact(context.Context) (repos.Repository, error)
}

// Transactor provides means to get hold of a repos.Repository instance in the context
// of a transaction, thus guaranteeing isolation/atomicity.
type Transactor interface {
	// Transact executes f and provides a repos.Repository instance that guarantees
	// isolation/atomicity.
	Transact(ctx context.Context, f func(repos.Repository) error) error
}
