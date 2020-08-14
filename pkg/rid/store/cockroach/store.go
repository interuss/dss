package cockroach

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/coreos/go-semver/semver"
	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/palantir/stacktrace"
	"go.uber.org/zap"
)

var (
	// DefaultClock is what is used as the Store's clock, returned from Dial.
	DefaultClock = clockwork.NewRealClock()
	// DefaultTimeout is the timeout applied to the txn retrier.
	// Note that this is not applied everywhere, but only
	// on the txn retrier.
	// If a given deadline is already supplied on the context, the earlier
	// deadline is used
	// TODO: use this in other function calls
	DefaultTimeout = 10 * time.Second

	// DatabaseName is the name of database storing remote ID data.
	DatabaseName = "defaultdb"
)

type repo struct {
	repos.ISA
	*subscriptionRepo
}

// Store is an implementation of store.Store using Cockroach DB as its backend
// store.
//
// TODO: Add the SCD interfaces here, and collapse this store with the
// outer pkg/cockroach
type Store struct {
	db     *cockroach.DB
	logger *zap.Logger
	clock  clockwork.Clock
}

// Interact implements store.Interactor interface.
func (s *Store) Interact(ctx context.Context) (repos.Repository, error) {
	logger := logging.WithValuesFromContext(ctx, s.logger)
	storeVersion, err := s.GetVersion(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error determining database RID schema version")
	}

	return &repo{
		ISA: NewISARepo(ctx, s.db, storeVersion, logger),
		subscriptionRepo: &subscriptionRepo{
			Queryable: s.db,
			logger:    logger,
			clock:     s.clock,
		},
	}, nil
}

// Transact supplies a new repo, that will perform all of the DB accesses
// in a Txn, and will retry any Txn's that fail due to retry-able errors
// (typically contention).
func (s *Store) Transact(ctx context.Context, f func(repo repos.Repository) error) error {
	logger := logging.WithValuesFromContext(ctx, s.logger)
	// TODO: consider what tx opts we want to support.
	// TODO: we really need to remove the upper cockroach package, and have one
	// "store" for everything
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	storeVersion, err := s.GetVersion(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Error determining database RID schema version")
	}
	return crdb.ExecuteTx(ctx, s.db.DB, nil /* nil txopts */, func(tx *sql.Tx) error {
		// Is this recover still necessary?
		defer recoverRollbackRepanic(ctx, tx)
		return f(&repo{
			ISA: NewISARepo(ctx, tx, storeVersion, logger),
			subscriptionRepo: &subscriptionRepo{
				Queryable: tx,
				logger:    logger,
				clock:     s.clock,
			},
		})
	})
}

// Close closes the underlying DB connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func recoverRollbackRepanic(ctx context.Context, tx *sql.Tx) {
	if p := recover(); p != nil {
		if err := tx.Rollback(); err != nil {
			logging.WithValuesFromContext(ctx, logging.Logger).Error(
				"failed to rollback transaction", zap.Error(err),
			)
		}
	}
}

// NewStore returns a Store instance connected to a cockroach instance via db.
func NewStore(db *cockroach.DB, logger *zap.Logger) *Store {
	return &Store{
		db:     db,
		logger: logger,
		clock:  DefaultClock,
	}
}

// CleanUp removes all database tables managed by s.
func (s *Store) CleanUp(ctx context.Context) error {
	const query = `
	DELETE FROM subscriptions WHERE id IS NOT NULL;
	DELETE FROM identification_service_areas WHERE id IS NOT NULL;`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// GetVersion returns the semver.Version for the Database.
// If the DB was is not bootstrapped using the schema manager we throw and error
func (s *Store) GetVersion(ctx context.Context) (*semver.Version, error) {
	versionStr, err := cockroach.GetVersion(ctx, s.db, DatabaseName)
	if err != nil {
		return nil, err
	}
	return semver.New(strings.Replace(versionStr, "v", "", -1)), nil
}
