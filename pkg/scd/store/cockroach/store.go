package cockroach

import (
	"context"

	// TODO: issue with below library: https://githubhelp.com/cockroachdb/cockroach-go/issues/65?ref=https://githubhelp.com
	// "github.com/cockroachdb/cockroach-go/crdb/crdbpgx"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/coreos/go-semver/semver"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/scd/repos"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v4"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

const (
	// currentMajorSchemaVersion is the current major schema version.
	currentMajorSchemaVersion = 3
)

var (
	// DefaultClock is what is used as the Store's clock, returned from Dial.
	DefaultClock = clockwork.NewRealClock()

	// DatabaseName is the name of database storing strategic conflict detection data.
	DatabaseName = "scd"
)

// repo is an implementation of repos.Repo using
// a CockroachDB transaction.
type repo struct {
	q      dsssql.Queryable
	logger *zap.Logger
	clock  clockwork.Clock
}

// Store is an implementation of an scd.Store using
// a CockroachDB database.
type Store struct {
	db     *cockroach.DB
	logger *zap.Logger
	clock  clockwork.Clock
}

// NewStore returns a Store instance connected to a cockroach instance via db.
func NewStore(ctx context.Context, db *cockroach.DB, logger *zap.Logger) (*Store, error) {
	store := &Store{
		db:     db,
		logger: logger,
		clock:  DefaultClock,
	}

	if err := store.CheckCurrentMajorSchemaVersion(ctx); err != nil {
		return nil, stacktrace.Propagate(err, "Strategic conflict detection schema version check failed")
	}

	return store, nil
}

// CheckCurrentMajorSchemaVersion returns nil if s supports the current major schema version.
func (s *Store) CheckCurrentMajorSchemaVersion(ctx context.Context) error {
	vs, err := s.GetVersion(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get database schema version for strategic conflict detection")
	}
	if vs == cockroach.UnknownVersion {
		return stacktrace.NewError("Strategic conflict detection database has not been bootstrapped with Schema Manager, Please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas")
	}

	if currentMajorSchemaVersion != vs.Major {
		return stacktrace.NewError("Unsupported schema version for strategic conflict detection! Got %s, requires major version of %d. Please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas", vs, currentMajorSchemaVersion)
	}

	return nil
}

// Interact implements store.Interactor interface.
func (s *Store) Interact(_ context.Context) (repos.Repository, error) {
	return &repo{
		q:      s.db,
		logger: s.logger,
		clock:  s.clock,
	}, nil
}

// Transact implements store.Transactor interface.
func (s *Store) Transact(ctx context.Context, f func(context.Context, repos.Repository) error) error {
	return crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		return f(ctx, &repo{
			q:      tx,
			logger: s.logger,
			clock:  s.clock,
		})
	})
}

// Close closes the underlying DB connection.
func (s *Store) Close() error {
	s.db.Close()
	return nil
}

// GetVersion returns the Version string for the Database.
// If the DB was is not bootstrapped using the schema manager we throw and error
func (s *Store) GetVersion(ctx context.Context) (*semver.Version, error) {
	return s.db.GetVersion(ctx, DatabaseName)
}
