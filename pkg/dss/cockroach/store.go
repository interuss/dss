package cockroach

import (
	"context"
	"database/sql"
	"strconv"
	"time"
	// Pull in the postgres database driver
)

type Store struct {
	*sql.DB
}

type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Convert updatedAt to a string, why not make it smaller
// WARNING: Changing this will cause RMW errors
// 32 is the highest value allowed by strconv
var versionBase = 32

func versionStringToTimestamp(s string) (time.Time, error) {
	var t time.Time
	nanos, err := strconv.ParseUint(s, versionBase, 64)
	if err != nil {
		return t, err
	}
	return time.Unix(0, int64(nanos)), nil
}

func timestampToVersionString(t time.Time) string {
	return strconv.FormatUint(uint64(t.UnixNano()), versionBase)
}

// Close closes the underlying DB connection.
func (s *Store) Close() error {
	return s.DB.Close()
}

// Bootstrap bootstraps the underlying database with required tables.
//
// TODO: We should handle database migrations properly, but bootstrap both us
// *and* the database with this manual approach here.
func (s *Store) Bootstrap(ctx context.Context) error {
	const query = `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY,
		owner STRING NOT NULL,
		url STRING NOT NULL,
		notification_index INT4 DEFAULT 0,
		starts_at TIMESTAMPTZ,
		ends_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL,
		INDEX starts_at_idx (starts_at),
		INDEX ends_at_idx (ends_at),
		CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
	);
	CREATE TABLE IF NOT EXISTS cells_subscriptions (
		cell_id INT64 NOT NULL,
		cell_level INT CHECK (cell_level BETWEEN 0 and 30),
		subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE,
		updated_at TIMESTAMPTZ NOT NULL,
		PRIMARY KEY (cell_id, subscription_id),
		INDEX cell_id_idx (cell_id),
		INDEX subscription_id_idx (subscription_id)
	);
	CREATE TABLE IF NOT EXISTS identification_service_areas (
		id UUID PRIMARY KEY,
		owner STRING NOT NULL,
		url STRING NOT NULL,
		starts_at TIMESTAMPTZ,
		ends_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL,
		INDEX starts_at_idx (starts_at),
		INDEX ends_at_idx (ends_at),
		CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
	);
	CREATE TABLE IF NOT EXISTS cells_identification_service_areas (
		cell_id INT64 NOT NULL,
		cell_level INT CHECK (cell_level BETWEEN 0 and 30),
		identification_service_area_id UUID NOT NULL REFERENCES identification_service_areas (id) ON DELETE CASCADE,
		updated_at TIMESTAMPTZ NOT NULL,	
		PRIMARY KEY (cell_id, identification_service_area_id),
		INDEX cell_id_idx (cell_id),
		INDEX identification_service_area_id_idx (identification_service_area_id)
	);
	`
	_, err := s.ExecContext(ctx, query)
	return err
}

// cleanUp drops all required tables from the store, useful for testing.
func (s *Store) cleanUp(ctx context.Context) error {
	const query = `
	DROP TABLE IF EXISTS cells_subscriptions;
	DROP TABLE IF EXISTS subscriptions;
	DROP TABLE IF EXISTS cells_identification_service_areas;
	DROP TABLE IF EXISTS identification_service_areas;`

	_, err := s.ExecContext(ctx, query)
	return err
}
