package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"

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

func (c *Store) fetchCellsForSubscription(ctx context.Context, q dsssql.Queryable, id scdmodels.ID) (s2.CellUnion, error) {
	var (
		cellsQuery = `
			SELECT
				cell_id
			FROM
				scd_cells_subscriptions
			WHERE
				subscription_id = $1
		`
	)

	rows, err := q.QueryContext(ctx, cellsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("fetchCellsForSubscription Query error: %s", err)
	}
	defer rows.Close()

	var (
		cu   s2.CellUnion
		cidi int64
	)
	for rows.Next() {
		if err := rows.Scan(&cidi); err != nil {
			return nil, fmt.Errorf("fetchCellsForSubscription row scan error: %s", err)
		}
		cu = append(cu, s2.CellID(cidi))
	}

	return cu, rows.Err()
}

func (c *Store) fetchSubscriptions(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*scdmodels.Subscription, error) {
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
	ctx context.Context, q dsssql.Queryable, cells []int64) ([]*scdmodels.Subscription, error) {
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
	// TODO: Relevant Subscriptions are ones that overlap the old or new
	//   Operation/Constraint Volume4D, not just any that end after the current
	//   time.  The query logic below needs to be fixed.
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

func (c *Store) fetchSubscription(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*scdmodels.Subscription, error) {
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

func (c *Store) fetchSubscriptionByID(ctx context.Context, q dsssql.Queryable, id scdmodels.ID) (*scdmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_subscriptions
			WHERE
				id = $1`, subscriptionFieldsWithPrefix)
	)
	result, err := c.fetchSubscription(ctx, q, query, id)
	if err != nil {
		return nil, err
	}
	result.Cells, err = c.fetchCellsForSubscription(ctx, q, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Store) fetchSubscriptionByIDAndOwner(ctx context.Context, q dsssql.Queryable, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error) {
	var query = fmt.Sprintf(`
		SELECT
			%s
		FROM
			scd_subscriptions
		WHERE
			id = $1
		AND
			owner = $2`, subscriptionFieldsWithPrefix)
	result, err := c.fetchSubscription(ctx, q, query, id, owner)
	if err != nil {
		return nil, err
	}
	result.Cells, err = c.fetchCellsForSubscription(ctx, q, id)
	if err != nil {
		return nil, err
	}
	ops, err := c.searchOperations(ctx, q, &dssmodels.Volume4D{
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeLo: result.AltitudeLo,
			AltitudeHi: result.AltitudeHi,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return result.Cells, nil
			}),
		},
		StartTime: result.StartTime,
		EndTime:   result.EndTime,
	}, owner)
	if err != nil {
		return nil, err
	}
	for _, op := range ops {
		result.DependentOperations = append(result.DependentOperations, op.ID)
	}
	return result, nil
}

// fetchMaxSubscriptionCountByCellAndOwner counts how many subscriptions the
// owner has in each one of these cells, and returns the number of subscriptions
// in the cell with the highest number of subscriptions.
func (c *Store) fetchMaxSubscriptionCountByCellAndOwner(
	ctx context.Context, q dsssql.Queryable, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
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

func (c *Store) pushSubscription(ctx context.Context, q dsssql.Queryable, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {
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
			($1, $2, COALESCE((SELECT version from v), 0) + 1, $3, $4, $5, $6, $7, $8, $9, transaction_timestamp())
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
		s.ImplicitSubscription,
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
	sub, err := c.fetchSubscriptionByIDAndOwner(ctx, c.tx, id, owner)
	switch err {
	case nil:
		return sub, nil
	case sql.ErrNoRows:
		return nil, dsserr.NotFound(id.String())
	default:
		return nil, err
	}
}

// UpsertSubscription upserts subscription into the store and returns
// the resulting subscription including its ID.
func (c *Store) UpsertSubscription(ctx context.Context, s *scdmodels.Subscription) (*scdmodels.Subscription, []*scdmodels.Operation, error) {
	old, err := c.fetchSubscriptionByID(ctx, c.tx, s.ID)
	switch {
	case err == sql.ErrNoRows:
		break
	case err != nil:
		return nil, nil, err
	}

	switch {
	case old == nil && !s.Version.Empty():
		// The user wants to update an existing subscription, but one wasn't found.
		return nil, nil, dsserr.NotFound(s.ID.String())
	case old != nil && s.Version.Empty():
		// The user wants to create a new subscription but it already exists.
		return nil, nil, dsserr.AlreadyExists(s.ID.String())
	case old != nil && !s.Version.Matches(old.Version):
		// The user wants to update a subscription but the version doesn't match.
		return nil, nil, dsserr.VersionMismatch("old version")
	case old != nil && old.Owner != s.Owner:
		return nil, nil, dsserr.PermissionDenied(fmt.Sprintf("Subscription is owned by %s", old.Owner))
	}

	// Check the user hasn't created too many subscriptions in this area.
	// TODO: Revisit this logic since the SCD standard does not specify a maximum
	// number of Subscriptions
	count, err := c.fetchMaxSubscriptionCountByCellAndOwner(ctx, c.tx, s.Cells, s.Owner)
	if err != nil {
		c.logger.Warn("Error fetching max subscription count", zap.Error(err))
		return nil, nil, dsserr.Internal(
			"failed to fetch subscription count, rejecting request")
	} else if count >= maxSubscriptionsPerArea {
		errMsg := "too many existing subscriptions in this area already"
		if old != nil {
			errMsg = errMsg + ", rejecting update request"
		}
		return nil, nil, dsserr.Exhausted(errMsg)
	}

	newSubscription, err := c.pushSubscription(ctx, c.tx, s)
	if err != nil {
		return nil, nil, err
	}
	newSubscription.Cells = s.Cells

	var affectedOperations []*scdmodels.Operation
	if len(s.Cells) > 0 {
		ops, err := c.searchOperations(ctx, c.tx, &dssmodels.Volume4D{
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			SpatialVolume: &dssmodels.Volume3D{
				AltitudeLo: s.AltitudeLo,
				AltitudeHi: s.AltitudeHi,
				Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
					return s.Cells, nil
				}),
			},
		}, s.Owner)
		if err != nil {
			return nil, nil, err
		}
		affectedOperations = ops
	}

	return newSubscription, affectedOperations, nil
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
			owner = $2
		AND
			0 = ALL (
				SELECT
					COALESCE(COUNT(scd_operations.id), 0)
				AS
					counter
				FROM
					scd_operations
				JOIN
					scd_subscriptions
				ON
					scd_operations.subscription_id = scd_subscriptions.id
				WHERE
					scd_operations.subscription_id = $1
			)`
	)

	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := c.fetchSubscriptionByID(ctx, c.tx, id)
	switch {
	case err == sql.ErrNoRows: // Return a 404 here.
		return nil, dsserr.NotFound(id.String())
	case err != nil:
		return nil, err
	case !version.Empty() && !version.Matches(old.Version):
		return nil, dsserr.VersionMismatch("old version")
	case old != nil && old.Owner != owner:
		return nil, dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner))
	}

	res, err := c.tx.ExecContext(ctx, query, id, owner)
	if err != nil {
		return nil, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, dsserr.BadRequest("failed to delete implicit subscription with active operation")
	}
	return old, nil
}

// Implements SubscriptionStore.SearchSubscriptions
func (c *Store) SearchSubscriptions(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_subscriptions
			JOIN
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
				COALESCE(starts_at <= $3, true)
			AND
				COALESCE(ends_at >= $2, true)`, subscriptionFieldsWithPrefix)
	)

	// TODO: Lazily calculate & cache spatial covering so that it is only ever
	// computed once on a particular Volume4D
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, err
	}

	if len(cells) == 0 {
		return nil, nil
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	subscriptions, err := c.fetchSubscriptions(
		ctx, c.tx, query, pq.Array(cids), v4d.StartTime, v4d.EndTime)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}
