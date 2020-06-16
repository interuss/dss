package repos

import "context"

// Repository contains all of the repo interfaces.
type Repository interface {
	ISA
	Subscription
	// InTxnRetrier supplies a new repo, that will perform all of the DB accesses
	// in a Txn, and will retry any Txn's that fail due to retry-able errors
	// (typically contention).
	InTxnRetrier(ctx context.Context, f func(ctx context.Context) error) error
	Close() error
}
