package cockroach

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/scd/repos"
	dsssql "github.com/interuss/dss/pkg/sql"
	"go.uber.org/zap"
)

var (
	// DefaultClock is what is used as the Store's clock, returned from Dial.
	DefaultClock = clockwork.NewRealClock()
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

// Bootstrap bootstraps the underlying database with required tables.
//
// TODO: We should handle database migrations properly, but bootstrap both us
// *and* the database with this manual approach here.
func (s *Store) Bootstrap(ctx context.Context) error {
	const query = `
	CREATE TABLE IF NOT EXISTS scd_subscriptions (
		id UUID PRIMARY KEY,
		owner STRING NOT NULL,
		version INT4 NOT NULL DEFAULT 0,
		url STRING NOT NULL,
		notification_index INT4 DEFAULT 0,
		notify_for_operations BOOL DEFAULT false,
		notify_for_constraints BOOL DEFAULT false,
		implicit BOOL DEFAULT false,
		starts_at TIMESTAMPTZ,
		ends_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL,
		INDEX owner_idx (owner),
		INDEX starts_at_idx (starts_at),
		INDEX ends_at_idx (ends_at),
		CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
		CHECK (notify_for_operations OR notify_for_constraints)
	);
	CREATE TABLE IF NOT EXISTS scd_cells_subscriptions (
		cell_id INT64 NOT NULL,
		cell_level INT CHECK (cell_level BETWEEN 0 and 30),
		subscription_id UUID NOT NULL REFERENCES scd_subscriptions (id) ON DELETE CASCADE,
		PRIMARY KEY (cell_id, subscription_id),
		INDEX cell_id_idx (cell_id),
		INDEX subscription_id_idx (subscription_id)
	);
	CREATE TABLE IF NOT EXISTS scd_operations (
		id UUID PRIMARY KEY,
		owner STRING NOT NULL,
		version INT4 NOT NULL DEFAULT 0,
		url STRING NOT NULL,
		altitude_lower REAL,
		altitude_upper REAL,
		starts_at TIMESTAMPTZ,
		ends_at TIMESTAMPTZ,
		subscription_id UUID NOT NULL REFERENCES scd_subscriptions(id) ON DELETE CASCADE,
		updated_at TIMESTAMPTZ NOT NULL,
		INDEX owner_idx (owner),
		INDEX altitude_lower_idx (altitude_lower),
		INDEX altitude_upper_idx (altitude_upper),
		INDEX starts_at_idx (starts_at),
		INDEX ends_at_idx (ends_at),
		INDEX updated_at_idx (updated_at),
		INDEX subscription_id_idx (subscription_id),
		CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
	);
	CREATE TABLE IF NOT EXISTS scd_cells_operations (
		cell_id INT64 NOT NULL,
		cell_level INT CHECK (cell_level BETWEEN 0 and 30),
		operation_id UUID NOT NULL REFERENCES scd_operations (id) ON DELETE CASCADE,
		PRIMARY KEY (cell_id, operation_id),
		INDEX cell_id_idx (cell_id),
		INDEX operation_id_idx (operation_id)
	);
	`
	_, err := s.db.ExecContext(ctx, query)
	return err
}
