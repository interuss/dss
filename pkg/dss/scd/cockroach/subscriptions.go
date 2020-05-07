package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	dsserr "github.com/interuss/dss/pkg/errors"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	// Defined in requirement DSS0030.
	maxSubscriptionsPerArea = 10
)

var (
	subscriptionFieldsWithIndices   [11]string
	subscriptionFieldsWithPrefix    string
	subscriptionFieldsWithoutPrefix string
)

func init() {
	subscriptionFieldsWithIndices[0] = "id"
	subscriptionFieldsWithIndices[1] = "owner"
	subscriptionFieldsWithIndices[2] = "version"
	subscriptionFieldsWithIndices[3] = "url"
	subscriptionFieldsWithIndices[4] = "notification_index"
	subscriptionFieldsWithIndices[5] = "notify_for_operations"
	subscriptionFieldsWithIndices[6] = "notify_for_constraints"
	subscriptionFieldsWithIndices[7] = "implicit"
	subscriptionFieldsWithIndices[8] = "starts_at"
	subscriptionFieldsWithIndices[9] = "ends_at"
	subscriptionFieldsWithIndices[10] = "updated_at"

	subscriptionFieldsWithoutPrefix = strings.Join(
		subscriptionFieldsWithIndices[:], ",",
	)

	withPrefix := make([]string, 11)
	for idx, field := range subscriptionFieldsWithIndices {
		withPrefix[idx] = "scd_subscriptions." + field
	}

	subscriptionFieldsWithPrefix = strings.Join(
		withPrefix[:], ",",
	)
}

func (c *Store) fetchSubscriptions(ctx context.Context, q queryable, query string, args ...interface{}) ([]*scdmodels.Subscription, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*scdmodels.Subscription
	for rows.Next() {
		var (
			s         = new(scdmodels.Subscription)
			updatedAt time.Time
		)
		err := rows.Scan(
			&s.ID,
			&s.Owner,
			&s.Version,
			&s.BaseURL,
			&s.NotificationIndex,
			&s.NotifyForOperations,
			&s.NotifyForConstraints,
			&s.ImplicitSubscription,
			&s.StartTime,
			&s.EndTime,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}
		payload = append(payload, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Store) fetchSubscriptionsForNotification(
	ctx context.Context, q queryable, cells []int64) ([]*scdmodels.Subscription, error) {
	// TODO(dsansome): upgrade to cockroachdb 19.2.0 and convert this to a single
	// UPDATE FROM query.

	// First: get unique subscription IDs.
	var query = `
			SELECT DISTINCT
				subscription_id
			FROM
				scd_cells_subscriptions
			WHERE
				cell_id = ANY($1)`
	rows, err := q.QueryContext(ctx, query, pq.Array(cells))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptionIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		subscriptionIDs = append(subscriptionIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Next: update the notification_index of each one and return the rest of the
	// data.
	var updateQuery = fmt.Sprintf(`
			UPDATE
				scd_subscriptions
			SET
				notification_index = notification_index + 1
			WHERE
				id = ANY($1)
			AND
				ends_at >= $2
			RETURNING
				%s`, subscriptionFieldsWithoutPrefix)
	return c.fetchSubscriptions(
		ctx, q, updateQuery, pq.Array(subscriptionIDs), c.clock.Now())
}

func (c *Store) fetchSubscription(ctx context.Context, q queryable, query string, args ...interface{}) (*scdmodels.Subscription, error) {
	subs, err := c.fetchSubscriptions(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(subs) > 1 {
		return nil, multierr.Combine(err, fmt.Errorf("query returned %d subscriptions", len(subs)))
	}
	if len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return subs[0], nil
}

func (c *Store) fetchSubscriptionByID(ctx context.Context, q queryable, id scdmodels.ID) (*scdmodels.Subscription, error) {
	var query = fmt.Sprintf(`
		SELECT
			%s
		FROM
			scd_subscriptions
		WHERE
			id = $1
		AND
			ends_at >= $2`, subscriptionFieldsWithPrefix)
	return c.fetchSubscription(ctx, q, query, id, c.clock.Now())
}

func (c *Store) fetchSubscriptionByIDAndOwner(ctx context.Context, q queryable, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error) {
	var query = fmt.Sprintf(`
		SELECT
			%s
		FROM
			scd_subscriptions
		WHERE
			id = $1
		AND
			owner = $2
		AND
			ends_at >= $3`, subscriptionFieldsWithPrefix)
	return c.fetchSubscription(ctx, q, query, id, owner, c.clock.Now())
}

// fetchMaxSubscriptionCountByCellAndOwner counts how many subscriptions the
// owner has in each one of these cells, and returns the number of subscriptions
// in the cell with the highest number of subscriptions.
func (c *Store) fetchMaxSubscriptionCountByCellAndOwner(
	ctx context.Context, q queryable, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	var query = `
    SELECT
      IFNULL(MAX(subscriptions_per_cell_id), 0)
    FROM (
      SELECT
        COUNT(*) AS subscriptions_per_cell_id
      FROM
        scd_subscriptions AS s,
        scd_cells_subscriptions as c
      WHERE
        s.id = c.subscription_id AND
        s.owner = $1 AND
        c.cell_id = ANY($2) AND
        s.ends_at >= $3
      GROUP BY c.cell_id
    )`

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	row := q.QueryRowContext(ctx, query, owner, pq.Array(cids), c.clock.Now())
	var ret int
	err := row.Scan(&ret)
	return ret, err
}

func (c *Store) pushSubscription(ctx context.Context, q queryable, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		WITH v AS (
			SELECT
				version
			FROM
				scd_subscriptions
			WHERE
				id = $1
		)
		UPSERT INTO
		  scd_subscriptions
		  (%s)
		VALUES
			($1, $2, COALESCE((SELECT version from v), 0) + 1, $3, $4, $5, $6, false, $7, $8, transaction_timestamp())
		RETURNING
			%s`, subscriptionFieldsWithoutPrefix, subscriptionFieldsWithPrefix)
		subscriptionCellQuery = `
		UPSERT INTO
			scd_cells_subscriptions
			(cell_id, cell_level, subscription_id)
		VALUES
			($1, $2, $3)
		`
		deleteLeftOverCellsForSubscriptionQuery = `
			DELETE FROM
				scd_cells_subscriptions
			WHERE
				cell_id != ALL($1)
			AND
				subscription_id = $2`
	)

	cids := make([]int64, len(s.Cells))
	clevels := make([]int, len(s.Cells))

	for i, cell := range s.Cells {
		cids[i] = int64(cell)
		clevels[i] = cell.Level()
	}

	cells := s.Cells
	s, err := c.fetchSubscription(ctx, q, upsertQuery,
		s.ID,
		s.Owner,
		s.BaseURL,
		s.NotificationIndex,
		s.NotifyForOperations,
		s.NotifyForConstraints,
		s.StartTime,
		s.EndTime)
	if err != nil {
		return nil, err
	}
	s.Cells = cells

	for i := range cids {
		if _, err := q.ExecContext(ctx, subscriptionCellQuery, cids[i], clevels[i], s.ID); err != nil {
			return nil, err
		}
	}

	if _, err := q.ExecContext(ctx, deleteLeftOverCellsForSubscriptionQuery, pq.Array(cids), s.ID); err != nil {
		return nil, err
	}

	return s, nil
}

// GetSubscription returns the subscription identified by "id".
func (c *Store) GetSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error) {
	return c.fetchSubscriptionByIDAndOwner(ctx, c.DB, id, owner)
}

// InsertSubscription inserts subscription into the store and returns
// the resulting subscription including its ID.
func (c *Store) InsertSubscription(ctx context.Context, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {

	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}
	old, err := c.fetchSubscriptionByID(ctx, tx, s.ID)
	switch {
	case err == sql.ErrNoRows:
		break
	case err != nil:
		return nil, multierr.Combine(err, tx.Rollback())
	}

	switch {
	case old == nil && !s.Version.Empty():
		// The user wants to update an existing subscription, but one wasn't found.
		return nil, multierr.Combine(dsserr.NotFound(s.ID.String()), tx.Rollback())
	case old != nil && s.Version.Empty():
		// The user wants to create a new subscription but it already exists.
		return nil, multierr.Combine(dsserr.AlreadyExists(s.ID.String()), tx.Rollback())
	case old != nil && !s.Version.Matches(old.Version):
		// The user wants to update a subscription but the version doesn't match.
		return nil, multierr.Combine(dsserr.VersionMismatch("old version"), tx.Rollback())
	case old != nil && old.Owner != s.Owner:
		return nil, multierr.Combine(dsserr.PermissionDenied(fmt.Sprintf("Subscription is owned by %s", old.Owner)), tx.Rollback())
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := s.AdjustTimeRange(c.clock.Now(), old); err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	// Check the user hasn't created too many subscriptions in this area.
	count, err := c.fetchMaxSubscriptionCountByCellAndOwner(ctx, tx, s.Cells, s.Owner)
	if err != nil {
		c.logger.Warn("Error fetching max subscription count", zap.Error(err))
		return nil, multierr.Combine(dsserr.Internal(
			"failed to fetch subscription count, rejecting request"), tx.Rollback())
	} else if count >= maxSubscriptionsPerArea {
		errMsg := "too many existing subscriptions in this area already"
		if old != nil {
			errMsg = errMsg + ", rejecting update request"
		}
		return nil, multierr.Combine(dsserr.Exhausted(errMsg), tx.Rollback())
	}

	newSubscription, err := c.pushSubscription(ctx, tx, s)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return newSubscription, nil
}

// DeleteSubscription deletes the subscription identified by "id" and
// returns the deleted subscription.
func (c *Store) DeleteSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner, version scdmodels.Version) (*scdmodels.Subscription, error) {
	const (
		query = `
		DELETE FROM
			scd_subscriptions
		WHERE
			id = $1
		AND
			owner = $2`
	)

	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}

	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := c.fetchSubscriptionByID(ctx, tx, id)
	switch {
	case err == sql.ErrNoRows: // Return a 404 here.
		return nil, multierr.Combine(dsserr.NotFound(id.String()), tx.Rollback())
	case err != nil:
		return nil, multierr.Combine(err, tx.Rollback())
	case !version.Empty() && !version.Matches(old.Version):
		return nil, multierr.Combine(dsserr.VersionMismatch("old version"), tx.Rollback())
	case old != nil && old.Owner != owner:
		return nil, multierr.Combine(dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner)), tx.Rollback())
	}

	if _, err := tx.ExecContext(ctx, query, id, owner); err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return old, nil
}

// SearchSubscriptions returns all subscriptions in "cells".
func (c *Store) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*scdmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_subscriptions
			LEFT JOIN
				(SELECT DISTINCT 
					scd_cells_subscriptions.subscription_id
				FROM
					scd_cells_subscriptions
				WHERE
					scd_cells_subscriptions.cell_id = ANY($1)
				)
			AS
				unique_subscription_ids
			ON
				scd_subscriptions.id = unique_subscription_ids.subscription_id
			WHERE
				scd_subscriptions.owner = $2
			AND
				ends_at >= $3`, subscriptionFieldsWithPrefix)
	)

	if len(cells) == 0 {
		return nil, dsserr.BadRequest("no location provided")
	}

	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	subscriptions, err := c.fetchSubscriptions(
		ctx, tx, query, pq.Array(cids), owner, c.clock.Now())
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}
