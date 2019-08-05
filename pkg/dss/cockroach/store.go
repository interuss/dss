package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/lib/pq" // Pull in the postgres database driver
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
	"go.uber.org/multierr"
)

type scanner interface {
	Scan(fields ...interface{}) error
}

type subscriptionsRow struct {
	id                string
	owner             string
	url               string
	typesFilter       sql.NullString
	notificationIndex int32
	lastUsedAt        pq.NullTime
	beginsAt          pq.NullTime
	expiresAt         pq.NullTime
	updatedAt         time.Time
}

func (sr *subscriptionsRow) scan(scanner scanner) error {
	return scanner.Scan(&sr.id,
		&sr.owner,
		&sr.url,
		&sr.typesFilter,
		&sr.notificationIndex,
		&sr.lastUsedAt,
		&sr.beginsAt,
		&sr.expiresAt,
		&sr.updatedAt,
	)
}

func (sr *subscriptionsRow) toProtobuf() (*dspb.Subscription, error) {
	result := &dspb.Subscription{
		Id:    sr.id,
		Owner: sr.owner,
		Callbacks: &dspb.SubscriptionCallbacks{
			IdentificationServiceAreaUrl: sr.url,
		},
		NotificationIndex: int32(sr.notificationIndex),
	}

	if sr.beginsAt.Valid {
		ts, err := ptypes.TimestampProto(sr.beginsAt.Time)
		if err != nil {
			return nil, err
		}
		result.Begins = ts
	}

	if sr.expiresAt.Valid {
		ts, err := ptypes.TimestampProto(sr.expiresAt.Time)
		if err != nil {
			return nil, err
		}
		result.Expires = ts
	}

	return result, nil
}

// Store is an implementation of dss.Store using
// Cockroach DB as its backend store.
type Store struct {
	*sql.DB
}

// Close closes the underlying DB connection.
func (s *Store) Close() error {
	return s.DB.Close()
}

// insertSubscriptionUnchecked inserts subscription into the store and returns
// the resulting subscription including its ID.
//
// Please note that this function is only meant to be used in tests/benchmarks
// to bootstrap a store with values. In particular, the function is not
// validating timestamps of last updates.
func (s *Store) insertSubscriptionUnchecked(ctx context.Context, subscription *dspb.Subscription, cells s2.CellUnion) (*dspb.Subscription, error) {
	const (
		subscriptionQuery = `
		INSERT INTO
			subscriptions
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, transaction_timestamp())
		RETURNING
			*`
		subscriptionCellQuery = `
		INSERT INTO
			cells_subscriptions
		VALUES
			($1, $2, $3, transaction_timestamp())
		`
	)

	sr := subscriptionsRow{
		id:                subscription.GetId(),
		owner:             subscription.GetOwner(),
		url:               subscription.Callbacks.GetIdentificationServiceAreaUrl(),
		notificationIndex: subscription.GetNotificationIndex(),
	}

	if ts := subscription.GetBegins(); ts != nil {
		begins, err := ptypes.Timestamp(ts)
		if err != nil {
			return nil, err
		}
		sr.beginsAt.Time = begins
		sr.beginsAt.Valid = true
	}

	if ts := subscription.GetExpires(); ts != nil {
		expires, err := ptypes.Timestamp(ts)
		if err != nil {
			return nil, err
		}
		sr.expiresAt.Time = expires
		sr.expiresAt.Valid = true
	}

	tx, err := s.Begin()
	if err != nil {
		return nil, err
	}

	if err := sr.scan(tx.QueryRowContext(
		ctx,
		subscriptionQuery,
		sr.id,
		sr.owner,
		sr.url,
		sr.typesFilter,
		sr.notificationIndex,
		sr.lastUsedAt,
		sr.beginsAt,
		sr.expiresAt,
	)); err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	for _, cell := range cells {
		if _, err := tx.ExecContext(ctx, subscriptionCellQuery, cell, cell.Level(), sr.id); err != nil {
			return nil, multierr.Combine(err, tx.Rollback())
		}
	}

	result := &dspb.Subscription{
		Id:    sr.id,
		Owner: sr.owner,
		Callbacks: &dspb.SubscriptionCallbacks{
			IdentificationServiceAreaUrl: sr.url,
		},
		NotificationIndex: sr.notificationIndex,
	}

	if sr.beginsAt.Valid {
		ts, err := ptypes.TimestampProto(sr.beginsAt.Time)
		if err != nil {
			return nil, multierr.Combine(err, tx.Rollback())
		}
		result.Begins = ts
	}

	if sr.expiresAt.Valid {
		ts, err := ptypes.TimestampProto(sr.expiresAt.Time)
		if err != nil {
			return nil, multierr.Combine(err, tx.Rollback())
		}
		result.Expires = ts
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, err
}

// GetSubscription returns the subscription identified by "id".
func (s *Store) GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error) {
	const (
		subscriptionQuery = `
		SELECT * FROM
			subscriptions
		WHERE
			id = $1`
	)

	tx, err := s.Begin()
	if err != nil {
		return nil, err
	}

	sr := subscriptionsRow{}

	if err := sr.scan(tx.QueryRowContext(ctx, subscriptionQuery, id)); err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	result, err := sr.toProtobuf()
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteSubscription deletes the subscription identified by "id" and
// returns the deleted subscription.
func (s *Store) DeleteSubscription(ctx context.Context, id string) (*dspb.Subscription, error) {
	const (
		subscriptionQuery = `
		DELETE FROM
			subscriptions
		WHERE
			id = $1
		RETURNING
			*`
	)

	tx, err := s.Begin()
	if err != nil {
		return nil, err
	}

	sr := subscriptionsRow{}

	if err := sr.scan(tx.QueryRowContext(ctx, subscriptionQuery, id)); err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	result, err := sr.toProtobuf()
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

// SearchSubscriptions returns all subscriptions in "cells".
func (s *Store) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*dspb.Subscription, error) {
	const (
		subscriptionsInCellsQuery = `
			SELECT
				subscriptions.*
			FROM
				subscriptions
			LEFT JOIN 
				(SELECT DISTINCT cells_subscriptions.subscription_id FROM cells_subscriptions WHERE cells_subscriptions.cell_id = ANY($1))
			AS
				unique_subscription_ids
			ON
				subscriptions.id = unique_subscription_ids.subscription_id
			WHERE
				subscriptions.owner = $2`
	)

	if len(cells) == 0 {
		return nil, errors.New("missing cell IDs for query")
	}

	tx, err := s.Begin()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, subscriptionsInCellsQuery, pq.Array(cells), owner)
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}
	defer rows.Close()

	var (
		row    = &subscriptionsRow{}
		result = []*dspb.Subscription{}
	)

	for rows.Next() {
		if err := row.scan(rows); err != nil {
			return nil, multierr.Combine(err, tx.Rollback())
		}
		pb, err := row.toProtobuf()
		if err != nil {
			return nil, multierr.Combine(err, tx.Rollback())
		}

		result = append(result, pb)
	}

	if err := rows.Err(); err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
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
		types_filter STRING,
		notification_index INT4 DEFAULT 0,
		last_used_at TIMESTAMPTZ,
		begins_at TIMESTAMPTZ,
		expires_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL,
		INDEX begins_at_idx (begins_at),
		INDEX expires_at_idx (expires_at),
		CHECK (begins_at IS NULL OR expires_at IS NULL OR begins_at < expires_at)
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
		starts_at TIMESTAMPTZ NOT NULL,
		ends_at TIMESTAMPTZ NOT NULL,
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
