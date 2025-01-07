package cockroach

import (
	"context"
	"fmt"
	"strings"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"

	"github.com/golang/geo/s2"
)

var (
	subscriptionFieldsWithIndices   [12]string
	subscriptionFieldsWithPrefix    string
	subscriptionFieldsWithoutPrefix string
)

// TODO Update database schema and fields below.
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
	subscriptionFieldsWithIndices[10] = "cells"
	subscriptionFieldsWithIndices[11] = "updated_at"

	subscriptionFieldsWithoutPrefix = strings.Join(
		subscriptionFieldsWithIndices[:], ",",
	)

	withPrefix := make([]string, 12)
	for idx, field := range subscriptionFieldsWithIndices {
		withPrefix[idx] = "scd_subscriptions." + field
	}

	subscriptionFieldsWithPrefix = strings.Join(
		withPrefix[:], ",",
	)
}

func (c *repo) fetchCellsForSubscription(ctx context.Context, q dsssql.Queryable, id dssmodels.ID) (s2.CellUnion, error) {
	var (
		cellsQuery = `
			SELECT
				unnest(cells) as cell_id
			FROM
				scd_subscriptions
			WHERE
				id = $1
		`
	)

	uid, err := id.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	rows, err := q.Query(ctx, cellsQuery, uid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", cellsQuery)
	}
	defer rows.Close()

	var (
		cu   s2.CellUnion
		cidi int64
	)
	for rows.Next() {
		err := rows.Scan(&cidi)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Subscription cell row")
		}
		cu = append(cu, s2.CellID(cidi))
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}
	return cu, nil
}

func (c *repo) fetchSubscriptions(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*scdmodels.Subscription, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", query)
	}
	defer rows.Close()

	var payload []*scdmodels.Subscription
	var cids []int64
	for rows.Next() {
		var (
			s         = new(scdmodels.Subscription)
			updatedAt time.Time
			version   int
		)
		err = rows.Scan(
			&s.ID,
			&s.Manager,
			&version,
			&s.USSBaseURL,
			&s.NotificationIndex,
			&s.NotifyForOperationalIntents,
			&s.NotifyForConstraints,
			&s.ImplicitSubscription,
			&s.StartTime,
			&s.EndTime,
			&cids,
			&updatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Subscription row")
		}
		s.Version = scdmodels.NewOVNFromTime(updatedAt, s.ID.String())
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error generating Subscription version")
		}
		s.SetCells(cids)
		payload = append(payload, s)
	}
	if err = rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}

	return payload, nil
}

func (c *repo) fetchSubscription(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*scdmodels.Subscription, error) {
	subs, err := c.fetchSubscriptions(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(subs) > 1 {
		return nil, stacktrace.NewError("Query returned %d subscriptions when only 0 or 1 was expected", len(subs))
	}
	if len(subs) == 0 {
		return nil, nil
	}
	return subs[0], nil
}

func (c *repo) fetchSubscriptionByID(ctx context.Context, q dsssql.Queryable, id dssmodels.ID) (*scdmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_subscriptions
			WHERE
				id = $1::uuid`, subscriptionFieldsWithPrefix)
	)
	uid, err := id.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	result, err := c.fetchSubscription(ctx, q, query, uid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Subscription")
	}
	if result == nil {
		return nil, nil
	}
	result.Cells, err = c.fetchCellsForSubscription(ctx, q, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching cells for Subscription")
	}
	return result, nil
}

func (c *repo) pushSubscription(ctx context.Context, q dsssql.Queryable, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		WITH v AS (
			SELECT
				version
			FROM
				scd_subscriptions
			WHERE
				id = $1::uuid
		)
		INSERT INTO
		  scd_subscriptions
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, transaction_timestamp())
		ON CONFLICT (%s) DO UPDATE
			SET %s = $2,
				%s = $3::int,
				%s = $4,
				%s = $5::int,
				%s = $6,
				%s = $7,
				%s = $8::bool,
				%s = $9,
				%s = $10,
				%s = $11,
				%s = transaction_timestamp()
		RETURNING
			%s`,
			subscriptionFieldsWithoutPrefix,
			subscriptionFieldsWithIndices[0],
			subscriptionFieldsWithIndices[1],
			subscriptionFieldsWithIndices[2],
			subscriptionFieldsWithIndices[3],
			subscriptionFieldsWithIndices[4],
			subscriptionFieldsWithIndices[5],
			subscriptionFieldsWithIndices[6],
			subscriptionFieldsWithIndices[7],
			subscriptionFieldsWithIndices[8],
			subscriptionFieldsWithIndices[9],
			subscriptionFieldsWithIndices[10],
			subscriptionFieldsWithIndices[11],
			subscriptionFieldsWithPrefix,
		)
	)

	cids := make([]int64, len(s.Cells))
	// TODO get rid of clevels?
	clevels := make([]int, len(s.Cells))

	for i, cell := range s.Cells {
		cids[i] = int64(cell)
		clevels[i] = cell.Level()
	}

	id, err := s.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	s, err = c.fetchSubscription(ctx, q, upsertQuery,
		id,
		s.Manager,
		0,
		s.USSBaseURL,
		s.NotificationIndex,
		s.NotifyForOperationalIntents,
		s.NotifyForConstraints,
		s.ImplicitSubscription,
		s.StartTime,
		s.EndTime,
		cids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Subscription from upsert query")
	}
	if s == nil {
		return nil, stacktrace.NewError("Upsert query did not return a Subscription")
	}

	return s, nil
}

// GetSubscription returns the subscription identified by "id".
func (c *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	sub, err := c.fetchSubscriptionByID(ctx, c.q, id)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	} else if sub == nil {
		return nil, nil
	}
	return sub, nil
}

// Implements repos.Subscription.UpsertSubscription
func (c *repo) UpsertSubscription(ctx context.Context, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	newSubscription, err := c.pushSubscription(ctx, c.q, s)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}
	newSubscription.Cells = s.Cells

	return newSubscription, nil
}

// DeleteSubscription deletes the subscription identified by "id" and
// returns the deleted subscription.
func (c *repo) DeleteSubscription(ctx context.Context, id dssmodels.ID) error {
	const (
		query = `
		DELETE FROM
			scd_subscriptions
		WHERE
			id = $1`
	)

	uid, err := id.PgUUID()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	res, err := c.q.Exec(ctx, query, uid)
	if err != nil {
		return stacktrace.Propagate(err, "Error in query: %s", query)
	}

	if res.RowsAffected() == 0 {
		return stacktrace.NewError("Attempted to delete non-existent Subscription")
	}

	return nil
}

// Implements SubscriptionStore.SearchSubscriptions
func (c *repo) SearchSubscriptions(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_subscriptions
				WHERE
					cells && $1
				AND
					COALESCE(starts_at <= $3, true)
				AND
					COALESCE(ends_at >= $2, true)
				LIMIT $4`, subscriptionFieldsWithPrefix)
	)

	// TODO: Lazily calculate & cache spatial covering so that it is only ever
	// computed once on a particular Volume4D
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not calculate spatial covering")
	}

	if len(cells) == 0 {
		return nil, nil
	}

	subscriptions, err := c.fetchSubscriptions(
		ctx, c.q, query, dsssql.CellUnionToCellIds(cells), v4d.StartTime, v4d.EndTime, dssmodels.MaxResultLimit)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to fetch Subscriptions")
	}

	return subscriptions, nil
}

// Implements scd.repos.Subscription.IncrementNotificationIndices
func (c *repo) IncrementNotificationIndices(ctx context.Context, subscriptionIds []dssmodels.ID) ([]int, error) {
	var updateQuery = `
			UPDATE scd_subscriptions
			SET notification_index = notification_index + 1
			WHERE id = ANY($1)
			RETURNING notification_index`

	ids := make([]string, len(subscriptionIds))
	for i, id := range subscriptionIds {
		ids[i] = id.String()
	}

	rows, err := c.q.Query(ctx, updateQuery, ids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", updateQuery)
	}
	defer rows.Close()

	var indices []int
	for rows.Next() {
		var notificationIndex int
		err := rows.Scan(&notificationIndex)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning notification index row")
		}
		indices = append(indices, notificationIndex)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}

	if len(indices) != len(subscriptionIds) {
		return nil, stacktrace.NewError(
			"Expected %d notification_index results when incrementing but got %d instead",
			len(subscriptionIds), len(indices))
	}

	return indices, nil
}

func (c *repo) LockSubscriptionsOnCells(ctx context.Context, cells s2.CellUnion) error {

	const query = `
		SELECT
			id
		FROM
			scd_subscriptions
		WHERE
			cells && $1
		FOR UPDATE
	`

	_, err := c.q.Exec(ctx, query, dsssql.CellUnionToCellIds(cells))
	if err != nil {
		return stacktrace.Propagate(err, "Error in query: %s", query)
	}

	return nil
}

// ListExpiredSubscriptions lists all subscriptions older than the threshold.
// Their age is determined by their end time, or by their update time if they do not have an end time.
func (c *repo) ListExpiredSubscriptions(ctx context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	expiredSubsQuery := fmt.Sprintf(`
        SELECT
            %s
        FROM
            scd_subscriptions
        WHERE
            scd_subscriptions.ends_at IS NOT NULL AND scd_subscriptions.ends_at <= $1
            OR
            scd_subscriptions.ends_at IS NULL AND scd_subscriptions.updated_at <= $1 -- use last update time as reference if there is no end time
        LIMIT $2`, subscriptionFieldsWithPrefix)

	subscriptions, err := c.fetchSubscriptions(
		ctx, c.q, expiredSubsQuery,
		threshold,
		dssmodels.MaxResultLimit,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to fetch Subscriptions")
	}

	return subscriptions, nil

}
