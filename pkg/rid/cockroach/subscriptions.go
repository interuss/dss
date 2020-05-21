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
	maxSubscriptionsPerArea         = 10
	subscriptionFields              = "subscriptions.id, subscriptions.owner, subscriptions.url, subscriptions.notification_index, subscriptions.cells, subscriptions.starts_at, subscriptions.ends_at, subscriptions.updated_at"
	subscriptionFieldsWithoutPrefix = "id, owner, url, notification_index, cells, starts_at, ends_at, updated_at"
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
		return nil, multierr.Combine(err, fmt.Errorf("query returned %d subscriptions", len(subs)))
	}
	if len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return subs[0], nil
}

// TODO: ƒactor this out
func (c *SubscriptionStore) getOneByID(ctx context.Context, q dsssql.Queryable, id dssmodels.ID) (*ridmodels.Subscription, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM subscriptions
		WHERE id = $1
		AND ends_at >= $2`, subscriptionFields)
	return c.processOne(ctx, q, query, id, c.clock.Now())
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
      FROM
        subscriptions AS s,
        cells_subscriptions as c
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

func (c *SubscriptionStore) pushSubscription(ctx context.Context, q dsssql.Queryable, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		udpateQuery = fmt.Sprintf(`
		UPDATE
		  subscriptions
		SET (%s) = ($1, $2, $3, $4, $5, $6, transaction_timestamp())
		WHERE id = $1 AND updated_at = $7
		RETURNING
			%s`, subscriptionFieldsWithoutPrefix, subscriptionFields)
		insertQuery = fmt.Sprintf(`
		INSERT INTO
		  subscriptions
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, transaction_timestamp())
		RETURNING
			%s`, subscriptionFieldsWithoutPrefix, subscriptionFields)
	)

	cids := make([]int64, len(s.Cells))

	for i, cell := range s.Cells {
		cids[i] = int64(cell)
	}

	var err error
	if s.Version.Empty() {
		s, err = c.processOne(ctx, q, insertQuery,
			s.ID,
			s.Owner,
			s.URL,
			s.NotificationIndex,
			s.Cells,
			s.StartTime,
			s.EndTime)
		if err != nil {
			return nil, err
		}
	} else {
		s, err = c.processOne(ctx, q, udpateQuery,
			s.ID,
			s.Owner,
			s.URL,
			s.NotificationIndex,
			s.Cells,
			s.StartTime,
			s.EndTime,
			s.Version.ToTimestamp())
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Get returns the subscription identified by "id".
func (c *SubscriptionStore) Get(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	return c.getOneByID(ctx, c.DB, id)
}

// Update updates the Subscription.. not yet implemented.
func (c *SubscriptionStore) Update(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	return nil, dsserr.Internal("not yet implemented")
}

// Insert inserts subscription into the store and returns
// the resulting subscription including its ID.
func (c *SubscriptionStore) Insert(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
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
	} else if count >= maxSubscriptionsPerArea {
		return nil, multierr.Combine(dsserr.Exhausted(
			"too many existing subscriptions in this area already"), tx.Rollback())
	}

	newSubscription, err := c.pushSubscription(ctx, tx, s)
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return newSubscription, nil
}

// Delete deletes the subscription identified by "id" and
// returns the deleted subscription.
func (c *SubscriptionStore) Delete(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	const (
		query = `
		DELETE FROM
			subscriptions
		WHERE
			id = $1
			AND owner = $2
			AND updated_at = $3
		RETURNING *`
	)
	return c.processOne(ctx, c.DB, query, s.ID, s.Owner, s.Version.ToTimestamp())
}

// Search returns all subscriptions in "cells".
func (c *SubscriptionStore) Search(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
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

// SearchByOwner returns all subscriptions in "cells".
func (c *SubscriptionStore) SearchByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
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

	return c.process(ctx, c.DB, query, pq.Array(cids), owner, c.clock.Now())
}
