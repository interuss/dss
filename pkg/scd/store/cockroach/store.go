package cockroach

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/coreos/go-semver/semver"
	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/scd/repos"
	dsssql "github.com/interuss/dss/pkg/sql"
	"go.uber.org/zap"
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
func NewStore(db *cockroach.DB, logger *zap.Logger) *Store {
	return &Store{
		db:     db,
		logger: logger,
		clock:  DefaultClock,
	}
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
	return crdb.ExecuteTx(ctx, s.db.DB, nil /* nil txopts */, func(tx *sql.Tx) error {
		return f(ctx, &repo{
			q:      tx,
			logger: s.logger,
			clock:  s.clock,
		})
	})
}

// Close closes the underlying DB connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// GetVersion returns the semver.Version for the Database.
// If the DB was is not bootstrapped using the schema manager we throw and error
func (s *Store) GetVersion(ctx context.Context) (*semver.Version, error) {
	return cockroach.GetVersion(ctx, s.db, DatabaseName)
}
