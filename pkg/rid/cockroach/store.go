package cockroach

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/repos"
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
)

// Store is an implementation of dss.Store using
// Cockroach DB as its backend store.
// TODO: Add the SCD interfaces here, and collapse this store with the
// outer pkg/cockroach
type Store struct {
	*ISAStore
	*SubscriptionStore

	db *cockroach.DB
}

// InTxnRetrier supplies a new repo, that will perform all of the DB accesses
// in a Txn, and will retry any Txn's that fail due to retry-able errors
// (typically contention).
// Note: Currently the Newly supplied Repo *does not* support nested calls
// to InTxnRetrier.
func (s *Store) InTxnRetrier(ctx context.Context, f func(ctx context.Context, repo repos.Repository) error) error {
	// TODO: consider what tx opts we want to support.
	// TODO: we really need to remove the upper cockroach package, and have one
	// "store" for everything
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()
	return crdb.ExecuteTx(ctx, s.db.DB, nil /* nil txopts */, func(tx *sql.Tx) error {
		// Is this recover still necessary?
		defer recoverRollbackRepanic(ctx, tx)
		err := f(ctx.WithValue(ctx, dssql.DBKey, tx), &storeCopy)
		return err
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
		ISAStore:          &ISAStore{dssql.Queryable{db}, logger},
		SubscriptionStore: &SubscriptionStore{dssql.Queryable{db}, DefaultClock, logger},
		db:                db,
	}
}

// Bootstrap bootstraps the underlying database with required tables.
//
// TODO: We should handle database migrations properly, but bootstrap both us
// *and* the database with this manual approach here.
func (s *Store) Bootstrap(ctx context.Context) error {
	/*
			The following tables correspond to the ASTM Remote ID standard A2.5.2.3:
			a) Cell ID:
					cells_identification_service_areas.cell_id
			 		cells_subscriptions.cell_id
			b) Subscription
				 	i. subscriptions.id
				 ii. subscriptions.owner
				iii. subscriptions.url
				 iv. subscriptions.starts_at and subscriptions.ends_at
				  v. the mapping from cells_subscriptions.subscription_id and cell_id
						 to subscriptions.id
				 vi. subscriptions.notification_index
			c) ISA
		 		 	i. identification_service_areas.id
				 ii. identification_service_areas.owner
				iii. identification_service_areas.url
				 iv. identification_service_areas.starts_at and
						 identification_service_areas.ends_at
				  v. the mapping from
						 cells_identification_service_areas.subscription_id and cell_id
						 to cells_identification_service_areas.id
	*/
	const query = `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY,
		owner STRING NOT NULL,
		url STRING NOT NULL,
		notification_index INT4 DEFAULT 0,
		starts_at TIMESTAMPTZ,
		ends_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL,
		cells INT64[] NOT NULL CHECK (array_length(cells, 1) > 0 AND array_length(cells, 1) IS NOT NULL),
		INDEX owner_idx (owner),
		INVERTED INDEX cells_idx (cells),
		INDEX starts_at_idx (starts_at),
		INDEX ends_at_idx (ends_at),
		CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
	);
	CREATE TABLE IF NOT EXISTS identification_service_areas (
		id UUID PRIMARY KEY,
		owner STRING NOT NULL,
		url STRING NOT NULL,
		starts_at TIMESTAMPTZ,
		ends_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL,
		cells INT64[] NOT NULL CHECK (array_length(cells, 1) IS NOT NULL),
		INDEX owner_idx (owner),
		INVERTED INDEX cells_idx (cells),
		INDEX starts_at_idx (starts_at),
		INDEX ends_at_idx (ends_at),
		INDEX updated_at_idx (updated_at),
		CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
	);
	`
	// TODO: example schema_versions. The onerow_enforcer ensures that 2 versions
	// can't exist at the same time.
	// CREATE TABLE schema_versions (
	// 	onerow_enforcer bool PRIMARY KEY DEFAULT TRUE CHECK(onerow_enforcer)
	// 	schema_version STRING NOT NULL,
	// );

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// GetVersion returns the current semver of the schema
func (s *Store) GetVersion(ctx context.Context) (string, error) {
	// We treat the existence of cells_subscriptions as running on the initial
	// version, 1.0.0
	const query = `
		SELECT EXISTS (
  		SELECT *
		  FROM information_schema.tables 
   		WHERE table_name = 'cells_subscriptions'
   )`
	row := s.db.QueryRowContext(ctx, query)
	var ret bool
	err := row.Scan(&ret)
	if err != nil {
		return "", err
	}
	if ret {
		// Base version
		return "v1.0.0", nil
	}
	// Version without cells joins table.
	// TODO: leverage proper migrations and use something like the query below.
	return "v2.0.0", nil
}

// 	// TODO steeling: we should leverage this function. Instead, we don't have
// 	// a great way migrate/seed the DB, and we can't combine that with the code
// 	// here.
// func (s *Store) GetVersion() float {

// 	const query = `
//     SELECT
//       IFNULL(schema_version, 'v1.0.0')
//     FROM
//     	schema_versions
//   	LIMIT 1`
// 	row := s.QueryRowContext(ctx, query)
// 	var ret string
// 	err := row.Scan(&ret)
// 	return ret, err
// }
