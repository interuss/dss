package cockroach

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	// Defined in requirement DSS0030.
	maxSubscriptionsPerArea = 10
)

var subscriptionFields = "subscriptions.id, subscriptions.owner, subscriptions.url, subscriptions.notification_index, subscriptions.starts_at, subscriptions.ends_at, subscriptions.updated_at"
var subscriptionFieldsWithoutPrefix = "id, owner, url, notification_index, starts_at, ends_at, updated_at"

func (c *Store) fetchSubscriptions(ctx context.Context, q queryable, query string, args ...interface{}) ([]*models.Subscription, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*models.Subscription
	for rows.Next() {
		s := new(models.Subscription)

		err := rows.Scan(
			&s.ID,
			&s.Owner,
			&s.Url,
			&s.NotificationIndex,
			&s.StartTime,
			&s.EndTime,
			&s.Version,
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

func (c *Store) fetchSubscriptionsByCellsWithoutOwner(ctx context.Context, q queryable, cells []int64, owner models.Owner) ([]*models.Subscription, error) {
	var subscriptionsQuery = fmt.Sprintf(`
		 SELECT
				%s
			FROM
				subscriptions
			LEFT JOIN
				(SELECT DISTINCT subscription_id FROM cells_subscriptions WHERE cell_id = ANY($1))
			AS
				unique_subscription_ids
			ON
				subscriptions.id = unique_subscription_ids.subscription_id
			WHERE
				subscriptions.owner != $2`, subscriptionFields)

	return c.fetchSubscriptions(ctx, q, subscriptionsQuery, pq.Array(cells), owner)
}

func (c *Store) fetchSubscription(ctx context.Context, q queryable, query string, args ...interface{}) (*models.Subscription, error) {
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

func (c *Store) fetchSubscriptionByID(ctx context.Context, q queryable, id models.ID) (*models.Subscription, error) {
	var query = fmt.Sprintf(`SELECT %s FROM subscriptions WHERE id = $1`, subscriptionFields)
	return c.fetchSubscription(ctx, q, query, id)
}

func (c *Store) fetchSubscriptionByIDAndOwner(ctx context.Context, q queryable, id models.ID, owner models.Owner) (*models.Subscription, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			subscriptions
		WHERE
			id = $1
			AND owner = $2`, subscriptionFields)
	return c.fetchSubscription(ctx, q, query, id, owner)
}

// fetchMaxSubscriptionCountByCellAndOwner counts how many subscriptions the
// owner has in each one of these cells, and returns the number of subscriptions
// in the cell with the highest number of subscriptions.
func (c *Store) fetchMaxSubscriptionCountByCellAndOwner(
	ctx context.Context, q queryable, cells s2.CellUnion, owner models.Owner) (int, error) {
	var query = `
    SELECT
      MAX(subscriptions_per_cell_id)
    FROM (
      SELECT
        COUNT(*) AS subscriptions_per_cell_id
      FROM
        subscriptions AS s,
        cells_subscriptions as c
      WHERE
        s.id = c.subscription_id AND
        s.owner = $1 AND
        c.cell_id = ANY($2)
      GROUP BY c.cell_id
    )`

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	row := q.QueryRowContext(ctx, query, owner, pq.Array(cids))
	var ret int
	err := row.Scan(&ret)
	return ret, err
}

func (c *Store) pushSubscription(ctx context.Context, q queryable, s *models.Subscription) (*models.Subscription, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		UPSERT INTO
		  subscriptions
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, transaction_timestamp())
		RETURNING
			%s`, subscriptionFieldsWithoutPrefix, subscriptionFields)
		subscriptionCellQuery = `
		UPSERT INTO
			cells_subscriptions
			(cell_id, cell_level, subscription_id)
		VALUES
			($1, $2, $3)
		`
		deleteLeftOverCellsForSubscriptionQuery = `
			DELETE FROM
				cells_subscriptions
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
		s.Url,
		s.NotificationIndex,
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

// Get returns the subscription identified by "id".
func (c *Store) GetSubscription(ctx context.Context, id models.ID) (*models.Subscription, error) {
	return c.fetchSubscriptionByID(ctx, c.DB, id)
}

// Insert inserts subscription into the store and returns
// the resulting subscription including its ID.
func (c *Store) InsertSubscription(ctx context.Context, s *models.Subscription) (*models.Subscription, error) {

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
	}

	// Check the user hasn't created too many subscriptions in this area.
	if old == nil {
		count, err := c.fetchMaxSubscriptionCountByCellAndOwner(ctx, tx, s.Cells, s.Owner)
		if err != nil {
			c.logger.Warn("Error fetching max subscription count", zap.Error(err))
		} else if count >= maxSubscriptionsPerArea {
			return nil, multierr.Combine(dsserr.Exhausted(
				"too many existing subscriptions in this area"), tx.Rollback())
		}
	}

	s, err = c.pushSubscription(ctx, tx, s)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s, nil
}

// DeleteSubscription deletes the subscription identified by "id" and
// returns the deleted subscription.
func (c *Store) DeleteSubscription(ctx context.Context, id models.ID, owner models.Owner, version *models.Version) (*models.Subscription, error) {
	const (
		query = `
		DELETE FROM
			subscriptions
		WHERE
			id = $1
			AND owner = $2`
	)

	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}

	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := c.fetchSubscriptionByIDAndOwner(ctx, tx, id, owner)
	switch {
	case err == sql.ErrNoRows: // Return a 404 here.
		return nil, multierr.Combine(dsserr.NotFound(id.String()), tx.Rollback())
	case err != nil:
		return nil, multierr.Combine(err, tx.Rollback())
	case !version.Empty() && !version.Matches(old.Version):
		return nil, multierr.Combine(dsserr.VersionMismatch("old version"), tx.Rollback())
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
func (c *Store) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner models.Owner) ([]*models.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				subscriptions
			LEFT JOIN
				(SELECT DISTINCT cells_subscriptions.subscription_id FROM cells_subscriptions WHERE cells_subscriptions.cell_id = ANY($1))
			AS
				unique_subscription_ids
			ON
				subscriptions.id = unique_subscription_ids.subscription_id
			WHERE
				subscriptions.owner = $2`, subscriptionFields)
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

	subscriptions, err := c.fetchSubscriptions(ctx, tx, query, pq.Array(cids), owner)
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}
