package repos

import "context"

// Repository contains all of the repo interfaces.
type Repository interface {
	ISA
	Subscription
}

type Transactor interface {
	Repository
	// InTxnRetrier supplies a new repo, that will perform all of the DB accesses
	// in a Txn, and will retry any Txn's that fail due to retry-able errors
	// (typically contention).
	InTxnRetrier(ctx context.Context, f func(ctx context.Context, repo Repository) error) error
	Close() error
}
