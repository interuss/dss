package cockroach

import (
	"context"
	"fmt"
	"strings"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/jackc/pgtype"

	"github.com/golang/geo/s2"
	"github.com/interuss/stacktrace"
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

	rows, err := q.Query(ctx, cellsQuery, id.PgUUID())
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
	pgCids := pgtype.Int8Array{}
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
			&pgCids,
			&updatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Subscription row")
		}
		s.Version = scdmodels.NewOVNFromTime(updatedAt, s.ID.String())
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error generating Subscription version")
		}
		var cids []int64
		if err := pgCids.AssignTo(&cids); err != nil {
			return nil, stacktrace.Propagate(err, "Error Converting jackc/pgtype to array")
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
				id = $1`, subscriptionFieldsWithPrefix)
	)
	result, err := c.fetchSubscription(ctx, q, query, id.PgUUID())
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
				id = $1
		)
		UPSERT INTO
		  scd_subscriptions
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, transaction_timestamp())
		RETURNING
			%s`, subscriptionFieldsWithoutPrefix, subscriptionFieldsWithPrefix)
	)

	cids := make([]int64, len(s.Cells))
	clevels := make([]int, len(s.Cells))

	for i, cell := range s.Cells {
		cids[i] = int64(cell)
		clevels[i] = cell.Level()
	}

	var pgCids pgtype.Int8Array
	if err := pgCids.Set(cids); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	s, err := c.fetchSubscription(ctx, q, upsertQuery,
		s.ID.PgUUID(),
		s.Manager,
		0,
		s.USSBaseURL,
		s.NotificationIndex,
		s.NotifyForOperationalIntents,
		s.NotifyForConstraints,
		s.ImplicitSubscription,
		s.StartTime,
		s.EndTime,
		pgCids)
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

	res, err := c.q.Exec(ctx, query, id.PgUUID())
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
					COALESCE(ends_at >= $2, true)`, subscriptionFieldsWithPrefix)
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

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	var pgCids pgtype.Int8Array

	if err := pgCids.Set(cids); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	subscriptions, err := c.fetchSubscriptions(
		ctx, c.q, query, pgCids, v4d.StartTime, v4d.EndTime)
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

	var pgIds pgtype.UUIDArray
	err := pgIds.Set(ids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}
	rows, err := c.q.Query(ctx, updateQuery, pgIds)
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
