package cockroach

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	dsssql "github.com/interuss/dss/pkg/sql"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	// Defined in requirement DSS0030.
	maxSubscriptionsPerArea = 10
	subscriptionFields      = "id, owner, url, notification_index, cells, starts_at, ends_at, updated_at"
)

// SubscriptionStore is an implementation of the SubscriptionRepo for CRDB.
type SubscriptionStore struct {
	*cockroach.DB

	clock  clockwork.Clock
	logger *zap.Logger
}

// process a query that should return one or many subscriptions.
func (c *SubscriptionStore) process(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*ridmodels.Subscription, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*ridmodels.Subscription
	cids := pq.Int64Array{}

	for rows.Next() {
		s := new(ridmodels.Subscription)

		err := rows.Scan(
			&s.ID,
			&s.Owner,
			&s.URL,
			&s.NotificationIndex,
			&cids,
			&s.StartTime,
			&s.EndTime,
			&s.Version,
		)
		if err != nil {
			return nil, err
		}
		s.SetCells(cids)
		payload = append(payload, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payload, nil
}

// processOne processes a query that should return exactly a single subscription.
func (c *SubscriptionStore) processOne(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*ridmodels.Subscription, error) {
	subs, err := c.process(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(subs) > 1 {
		return nil, fmt.Errorf("query returned %d subscriptions", len(subs))
	}
	if len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return subs[0], nil
}

// fetchMaxSubscriptionCountByCellAndOwner counts how many subscriptions the
// owner has in each one of these cells, and returns the number of subscriptions
// in the cell with the highest number of subscriptions.
func (c *SubscriptionStore) fetchMaxSubscriptionCountByCellAndOwner(
	ctx context.Context, q dsssql.Queryable, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	var query = `
    SELECT
      IFNULL(MAX(subscriptions_per_cell_id), 0)
    FROM (
      SELECT
        COUNT(*) AS subscriptions_per_cell_id
      FROM (
      	SELECT unnest(cells) as cell_id
      	FROM subscriptions
      	WHERE owner = $1
      		AND ends_at >= $2
      )
      WHERE
        cell_id = ANY($3)
      GROUP BY cell_id
    )`

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	row := q.QueryRowContext(ctx, query, owner, c.clock.Now(), pq.Int64Array(cids))
	var ret int
	err := row.Scan(&ret)
	return ret, err
}

func (c *SubscriptionStore) push(ctx context.Context, q dsssql.Queryable, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		udpateQuery = fmt.Sprintf(`
		UPDATE
		  subscriptions
		SET (%s) = ($1, $2, $3, $4, $5, $6, $7, transaction_timestamp())
		WHERE id = $1 AND updated_at = $8
		RETURNING
			%s`, subscriptionFields, subscriptionFields)
		insertQuery = fmt.Sprintf(`
		INSERT INTO
		  subscriptions
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, transaction_timestamp())
		RETURNING
			%s`, subscriptionFields, subscriptionFields)
	)

	cids := make([]int64, len(s.Cells))

	for i, cell := range s.Cells {
		cids[i] = int64(cell)
	}

	var err error
	var ret *ridmodels.Subscription
	if s.Version.Empty() {
		ret, err = c.processOne(ctx, q, insertQuery,
			s.ID,
			s.Owner,
			s.URL,
			s.NotificationIndex,
			pq.Int64Array(cids),
			s.StartTime,
			s.EndTime)
		if err != nil {
			return nil, err
		}
	} else {
		ret, err = c.processOne(ctx, q, udpateQuery,
			s.ID,
			s.Owner,
			s.URL,
			s.NotificationIndex,
			pq.Int64Array(cids),
			s.StartTime,
			s.EndTime,
			s.Version.ToTimestamp())
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

// GetSubscription returns the subscription identified by "id".
func (c *SubscriptionStore) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	// TODO(steeling) we should enforce startTime and endTime to not be null at the DB level.
	var query = fmt.Sprintf(`
		SELECT %s FROM subscriptions
		WHERE id = $1
		AND
			ends_at > $2`, subscriptionFields)
	return c.processOne(ctx, c.DB, query, id, c.clock.Now())
}

// UpdateSubscription updates the Subscription.. not yet implemented.
func (c *SubscriptionStore) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	return nil, dsserr.Internal("not yet implemented")
}

// InsertSubscription inserts subscription into the store and returns
// the resulting subscription including its ID.
func (c *SubscriptionStore) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}
	defer recoverRollbackRepanic(ctx, tx)

	// Check the user hasn't created too many subscriptions in this area.
	// TODO: bring this logic into the application
	count, err := c.fetchMaxSubscriptionCountByCellAndOwner(ctx, tx, s.Cells, s.Owner)
	if err != nil {
		c.logger.Warn("Error fetching max subscription count", zap.Error(err))
		return nil, multierr.Combine(dsserr.Internal(
			"failed to fetch subscription count, rejecting request"), tx.Rollback())
	}
	if count >= maxSubscriptionsPerArea {
		return nil, multierr.Combine(dsserr.Exhausted(
			"too many existing subscriptions in this area already"), tx.Rollback())
	}

	newSubscription, err := c.push(ctx, tx, s)
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return newSubscription, nil
}

// DeleteSubscription deletes the subscription identified by "id" and
// returns the deleted subscription.
func (c *SubscriptionStore) DeleteSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
		DELETE FROM
			subscriptions
		WHERE
			id = $1
			AND owner = $2
			AND updated_at = $3
		RETURNING %s`, subscriptionFields)
	)
	return c.processOne(ctx, c.DB, query, s.ID, s.Owner, s.Version.ToTimestamp())
}

// SearchSubscriptions returns all subscriptions in "cells".
func (c *SubscriptionStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				subscriptions
			WHERE
				cells && $1
			AND
				ends_at >= $3`, subscriptionFields)
	)

	if len(cells) == 0 {
		return nil, dsserr.BadRequest("no location provided")
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	return c.process(ctx, c.DB, query, pq.Array(cids), c.clock.Now())
}

// SearchSubscriptionsByOwner returns all subscriptions in "cells".
func (c *SubscriptionStore) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				subscriptions
			WHERE
				cells && $1
			AND
				subscriptions.owner = $2
			AND
				ends_at >= $3`, subscriptionFields)
	)

	if len(cells) == 0 {
		return nil, dsserr.BadRequest("no location provided")
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	return c.process(ctx, c.DB, query, pq.Int64Array(cids), owner, c.clock.Now())
}
