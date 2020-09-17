package cockroach

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/dpjacques/clockwork"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"

	"github.com/golang/geo/s2"
	repos "github.com/interuss/dss/pkg/rid/repos"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	subscriptionFields       = "id, owner, url, notification_index, cells, starts_at, ends_at, writer, updated_at"
	updateSubscriptionFields = "id, url, notification_index, cells, starts_at, ends_at, writer, updated_at"
)

func NewISASubscriptionRepo(ctx context.Context, db dssql.Queryable, dbVersion semver.Version, logger *zap.Logger, clock clockwork.Clock) repos.Subscription {
	if dbVersion.Compare(v310) >= 0 {
		return &subscriptionRepo{
			Queryable: db,
			logger:    logger,
			clock:     clock,
		}
	}
	return &subscriptionRepoV3{
		Queryable: db,
		logger:    logger,
		clock:     clock,
	}
}

// subscriptions is an implementation of the SubscriptionRepo for CRDB.
type subscriptionRepo struct {
	dssql.Queryable

	clock  clockwork.Clock
	logger *zap.Logger
}

// process a query that should return one or many subscriptions.
func (c *subscriptionRepo) process(ctx context.Context, query string, args ...interface{}) ([]*ridmodels.Subscription, error) {
	rows, err := c.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, fmt.Sprintf("Error in query: %s", query))
	}
	defer rows.Close()

	var payload []*ridmodels.Subscription
	cids := pq.Int64Array{}

	var writer sql.NullString
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
			&writer,
			&s.Version,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Subscription row")
		}
		s.Writer = writer.String
		s.SetCells(cids)
		payload = append(payload, s)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}
	return payload, nil
}

// processOne processes a query that should return exactly a single subscription.
func (c *subscriptionRepo) processOne(ctx context.Context, query string, args ...interface{}) (*ridmodels.Subscription, error) {
	subs, err := c.process(ctx, query, args...)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}
	if len(subs) > 1 {
		return nil, stacktrace.NewError("Query returned %d subscriptions when only 0 or 1 was expected", len(subs))
	}
	if len(subs) == 0 {
		return nil, nil
	}
	return subs[0], nil
}

// MaxSubscriptionCountInCellsByOwner counts how many subscriptions the
// owner has in each one of these cells, and returns the number of subscriptions
// in the cell with the highest number of subscriptions.
func (c *subscriptionRepo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	// TODO:steeling this query is expensive. The standard defines the max sub
	// per "area", but area is loosely defined. Since we may not have to be so
	// strict we could keep this count in memory, (or in some other storage).
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

	row := c.QueryRowContext(ctx, query, owner, c.clock.Now(), pq.Int64Array(cids))
	var ret int
	err := row.Scan(&ret)
	return ret, stacktrace.Propagate(err, "Error scanning subscription count row")
}

// GetSubscription returns the subscription identified by "id".
// Returns nil, nil if not found
func (c *subscriptionRepo) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	// TODO(steeling) we should enforce startTime and endTime to not be null at the DB level.
	var query = fmt.Sprintf(`
		SELECT %s FROM subscriptions
		WHERE id = $1`, subscriptionFields)
	return c.processOne(ctx, query, id)
}

// UpdateSubscription updates the Subscription.. not yet implemented.
// Returns nil, nil if ID, version not found
func (c *subscriptionRepo) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		updateQuery = fmt.Sprintf(`
		UPDATE
		  subscriptions
		SET (%s) = ($1, $2, $3, $4, $5, $6, $7, transaction_timestamp())
		WHERE id = $1 AND updated_at = $8
		RETURNING
			%s`, updateSubscriptionFields, subscriptionFields)
	)

	cids := make([]int64, len(s.Cells))

	for i, cell := range s.Cells {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, stacktrace.Propagate(err, "Error validating cell")
		}
		cids[i] = int64(cell)
	}

	return c.processOne(ctx, updateQuery,
		s.ID,
		s.URL,
		s.NotificationIndex,
		pq.Int64Array(cids),
		s.StartTime,
		s.EndTime,
		s.Writer,
		s.Version.ToTimestamp())
}

// InsertSubscription inserts subscription into the store and returns
// the resulting subscription including its ID.
func (c *subscriptionRepo) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		insertQuery = fmt.Sprintf(`
		INSERT INTO
		  subscriptions
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, transaction_timestamp())
		RETURNING
			%s`, subscriptionFields, subscriptionFields)
	)

	cids := make([]int64, len(s.Cells))

	for i, cell := range s.Cells {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, stacktrace.Propagate(err, "Error validating cell")
		}
		cids[i] = int64(cell)
	}

	return c.processOne(ctx, insertQuery,
		s.ID,
		s.Owner,
		s.URL,
		s.NotificationIndex,
		pq.Int64Array(cids),
		s.StartTime,
		s.EndTime,
		s.Writer)
}

// DeleteSubscription deletes the subscription identified by ID.
// It must be done in a txn and the version verified.
// Returns nil, nil if ID, version not found
func (c *subscriptionRepo) DeleteSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
		DELETE FROM
			subscriptions
		WHERE
			id = $1
			AND updated_at = $2
		RETURNING %s`, subscriptionFields)
	)
	return c.processOne(ctx, query, s.ID, s.Version.ToTimestamp())
}

// UpdateNotificationIdxsInCells incremement the notification for each sub in the given cells.
func (c *subscriptionRepo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var updateQuery = fmt.Sprintf(`
			UPDATE subscriptions
			SET notification_index = notification_index + 1
			WHERE
				cells && $1
				AND ends_at >= $2
			RETURNING %s`, subscriptionFields)

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}
	return c.process(
		ctx, updateQuery, pq.Int64Array(cids), c.clock.Now())
}

// SearchSubscriptions returns all subscriptions in "cells".
func (c *subscriptionRepo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				subscriptions
			WHERE
				cells && $1
			AND
				ends_at >= $2`, subscriptionFields)
	)

	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "no location provided")
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	return c.process(ctx, query, pq.Int64Array(cids), c.clock.Now())
}

// SearchSubscriptionsByOwner returns all subscriptions in "cells".
func (c *subscriptionRepo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
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
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "no location provided")
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	return c.process(ctx, query, pq.Int64Array(cids), owner, c.clock.Now())
}

// ListExpiredSubscriptions lists all expired Subscriptions based on expiredTime and writer
// expiredTime is compared with Subscription.endTime
func (c *subscriptionRepo) ListExpiredSubscriptions(ctx context.Context, writer string) ([]*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
	SELECT
		%s
	FROM
		subscriptions
	WHERE
		ends_at + INTERVAL '30' MINUTE <= CURRENT_TIMESTAMP
	AND
		writer = $1`, subscriptionFields)
	)

	if len(writer) == 0 {
		return nil, stacktrace.NewError("Writer is required")
	}

	return c.process(ctx, query, writer)
}
