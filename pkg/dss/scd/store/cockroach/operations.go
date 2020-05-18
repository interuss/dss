package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/lib/pq"
	"go.uber.org/multierr"
)

var (
	operationFieldsWithIndices   [10]string
	operationFieldsWithPrefix    string
	operationFieldsWithoutPrefix string
)

func init() {
	operationFieldsWithIndices[0] = "id"
	operationFieldsWithIndices[1] = "owner"
	operationFieldsWithIndices[2] = "version"
	operationFieldsWithIndices[3] = "url"
	operationFieldsWithIndices[4] = "altitude_lower"
	operationFieldsWithIndices[5] = "altitude_upper"
	operationFieldsWithIndices[6] = "starts_at"
	operationFieldsWithIndices[7] = "ends_at"
	operationFieldsWithIndices[8] = "subscription_id"
	operationFieldsWithIndices[9] = "updated_at"

	operationFieldsWithoutPrefix = strings.Join(
		operationFieldsWithIndices[:], ",",
	)

	withPrefix := make([]string, len(operationFieldsWithIndices))
	for idx, field := range operationFieldsWithIndices {
		withPrefix[idx] = "scd_operations." + field
	}

	operationFieldsWithPrefix = strings.Join(
		withPrefix[:], ",",
	)
}

func (s *Store) fetchOperations(ctx context.Context, q queryable, query string, args ...interface{}) ([]*scdmodels.Operation, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*scdmodels.Operation
	for rows.Next() {
		var (
			o         = &scdmodels.Operation{}
			updatedAt time.Time
		)
		if err := rows.Scan(
			&o.ID,
			&o.Owner,
			&o.Version,
			&o.USSBaseURL,
			&o.AltitudeLower,
			&o.AltitudeUpper,
			&o.StartTime,
			&o.EndTime,
			&o.SubscriptionID,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		o.OVN = scdmodels.NewOVNFromTime(updatedAt)
		payload = append(payload, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payload, nil
}

func (s *Store) fetchOperation(ctx context.Context, q queryable, query string, args ...interface{}) (*scdmodels.Operation, error) {
	operations, err := s.fetchOperations(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(operations) > 1 {
		return nil, multierr.Combine(err, fmt.Errorf("query returned %d operations", len(operations)))
	}
	if len(operations) == 0 {
		return nil, sql.ErrNoRows
	}
	return operations[0], nil
}

func (s *Store) fetchOperationByID(ctx context.Context, q queryable, id scdmodels.ID) (*scdmodels.Operation, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM
			scd_operations
		WHERE
			id = $1
		AND
			ends_at >= $2`, operationFieldsWithoutPrefix)
	return s.fetchOperation(ctx, q, query, id, s.clock.Now())
}

// pushOperation creates/updates the Operation identified by "id" and owned by
// "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created/updated Operation and all Subscriptions
// affected by the operation.
func (s *Store) pushOperation(ctx context.Context, q queryable, operation *scdmodels.Operation) (
	*scdmodels.Operation, []*scdmodels.Subscription, error) {
	var (
		upsertOperationsQuery = fmt.Sprintf(`
			WITH v AS (
				SELECT
					version
				FROM
					scd_operations
				WHERE
					id = $1
			)
			UPSERT INTO
				scd_operations
				(%s)
			VALUES
				($1, $2, COALESCE((SELECT version from v), 0) + 1, $3, $4, $5, $6, $7, $8, transaction_timestamp())
			RETURNING
				%s`, operationFieldsWithoutPrefix, operationFieldsWithPrefix)
		upsertCellsForOperationQuery = `
			UPSERT INTO
				scd_cells_operations
				(cell_id, cell_level, operation_id)
			VALUES
				($1, $2, $3)`
		deleteLeftOverCellsForOperationQuery = `
			DELETE FROM
				scd_cells_operations
			WHERE
				cell_id != ALL($1)
			AND
				operation_id = $2`
	)

	cids := make([]int64, len(operation.Cells))
	clevels := make([]int, len(operation.Cells))

	for i, cell := range operation.Cells {
		cids[i] = int64(cell)
		clevels[i] = cell.Level()
	}

	cells := operation.Cells
	operation, err := s.fetchOperation(ctx, q, upsertOperationsQuery,
		operation.ID,
		operation.Owner,
		operation.USSBaseURL,
		operation.AltitudeLower,
		operation.AltitudeUpper,
		operation.StartTime,
		operation.EndTime,
		operation.SubscriptionID,
	)
	if err != nil {
		return nil, nil, err
	}
	operation.Cells = cells

	for i := range cids {
		if _, err := q.ExecContext(ctx, upsertCellsForOperationQuery, cids[i], clevels[i], operation.ID); err != nil {
			return nil, nil, err
		}
	}

	if _, err := q.ExecContext(ctx, deleteLeftOverCellsForOperationQuery, pq.Array(cids), operation.ID); err != nil {
		return nil, nil, err
	}

	subscriptions, err := s.fetchSubscriptionsForNotification(ctx, q, cids)
	if err != nil {
		return nil, nil, err
	}

	return operation, subscriptions, nil
}

func (s *Store) populateOperationCells(ctx context.Context, q queryable, o *scdmodels.Operation) error {
	const query = `
	SELECT
		cell_id
	FROM
		scd_cells_operations
	WHERE operation_id = $1`

	rows, err := q.QueryContext(ctx, query, o.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cell int64
	o.Cells = s2.CellUnion{}

	for rows.Next() {
		if err := rows.Scan(&cell); err != nil {
			return err
		}
		o.Cells = append(o.Cells, s2.CellID(uint64(cell)))
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetOperation(ctx context.Context, id scdmodels.ID) (*scdmodels.Operation, error) {
	sub, err := s.fetchOperationByID(ctx, s.DB, id)
	switch err {
	case nil:
		return sub, nil
	case sql.ErrNoRows:
		return nil, dsserr.NotFound(id.String())
	default:
		return nil, err
	}
}

func (s *Store) DeleteOperation(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	var (
		deleteQuery = `
			DELETE FROM
				scd_operations
			WHERE
				id = $1
			AND
				owner = $2
		`
		deleteImplicitSubscriptionQuery = `
			DELETE FROM
				scd_subscriptions
			WHERE
				id = $1
			AND
				owner = $2
			AND
				implicit = true
			AND
				0 = ALL (
					SELECT
						COALESCE(COUNT(id),0)
					FROM
						scd_operations
					WHERE
						subscription_id = $1
				)
		`
	)

	tx, err := s.Begin()
	if err != nil {
		return nil, nil, err
	}

	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := s.fetchOperationByID(ctx, tx, id)
	switch {
	case err == sql.ErrNoRows: // Return a 404 here.
		return nil, nil, multierr.Combine(dsserr.NotFound(id.String()), tx.Rollback())
	case err != nil:
		return nil, nil, multierr.Combine(err, tx.Rollback())
	case old != nil && old.Owner != owner:
		return nil, nil, multierr.Combine(dsserr.PermissionDenied(fmt.Sprintf("Operation is owned by %s", old.Owner)), tx.Rollback())
	}
	if err := s.populateOperationCells(ctx, tx, old); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	cids := make([]int64, len(old.Cells))
	for i, cell := range old.Cells {
		cids[i] = int64(cell)
	}
	subscriptions, err := s.fetchSubscriptionsForNotification(ctx, tx, cids)
	if err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if _, err := tx.ExecContext(ctx, deleteQuery, id, owner); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}
	if _, err := tx.ExecContext(ctx, deleteImplicitSubscriptionQuery, old.SubscriptionID, owner); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return old, subscriptions, nil
}

func (s *Store) UpsertOperation(ctx context.Context, operation *scdmodels.Operation, key []scdmodels.OVN) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	tx, err := s.Begin()
	if err != nil {
		return nil, nil, err
	}

	old, err := s.fetchOperationByID(ctx, tx, operation.ID)
	switch {
	case err == sql.ErrNoRows:
		break
	case err != nil:
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	switch {
	case old == nil && !operation.Version.Empty():
		// The user wants to update an existing Operation, but one wasn't found.
		return nil, nil, multierr.Combine(dsserr.NotFound(operation.ID.String()), tx.Rollback())
	case old != nil && operation.Version.Empty():
		// The user wants to create a new Operation but it already exists.
		return nil, nil, multierr.Combine(dsserr.AlreadyExists(operation.ID.String()), tx.Rollback())
	case old != nil && !operation.Version.Matches(old.Version):
		// The user wants to update an Operation but the version doesn't match.
		return nil, nil, multierr.Combine(dsserr.VersionMismatch("old version"), tx.Rollback())
	case old != nil && old.Owner != operation.Owner:
		return nil, nil, multierr.Combine(dsserr.PermissionDenied(fmt.Sprintf("Operation is owned by %s", old.Owner)), tx.Rollback())
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := operation.ValidateTimeRange(); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	// TODO(tvoss): Investigate whether we can fold the check for OVNs into the
	// the upsert query by means of a CTE and a coalescing condition testing
	// whether all affected OVNs are matched.
	switch operation.State {
	case scdmodels.OperationStateAccepted, scdmodels.OperationStateActivated:
		operations, err := s.searchOperations(ctx, tx, &dssmodels.Volume4D{
			StartTime: operation.StartTime,
			EndTime:   operation.EndTime,
			SpatialVolume: &dssmodels.Volume3D{
				AltitudeHi: operation.AltitudeUpper,
				AltitudeLo: operation.AltitudeLower,
			},
		}, operation.Owner)
		if err != nil {
			return nil, nil, multierr.Combine(err, tx.Rollback())
		}

		keyIdx := map[scdmodels.OVN]struct{}{}
		for _, ovn := range key {
			keyIdx[ovn] = struct{}{}
		}

		for _, op := range operations {
			if _, match := keyIdx[op.OVN]; !match {
				return nil, nil, multierr.Combine(dsserr.VersionMismatch("ovn for affected operation differs"))
			}
		}
	default:
		// We default to not checking the OVNs for now for all other operation states.
	}

	area, subscribers, err := s.pushOperation(ctx, tx, operation)
	if err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return area, subscribers, nil
}

func (s *Store) searchOperations(ctx context.Context, q queryable, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error) {
	var (
		operationsIntersectingVolumeQuery = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_operations
			JOIN
				(SELECT DISTINCT
					scd_cells_operations.operation_id
				FROM
					scd_cells_operations
				WHERE
					scd_cells_operations.cell_id = ANY($1)
				)
			AS
				unique_operations
			ON
				scd_operations.id = unique_operations.operation_id
			WHERE
				COALESCE(scd_operations.altitude_upper >= $2, true)
			AND
				COALESCE(scd_operations.altitude_lower <= $3, true)
			AND
				COALESCE(scd_operations.ends_at >= $4, true)
			AND
				COALESCE(scd_operations.starts_at <= $5, true)
			AND
				scd_operations.ends_at >= $6`, operationFieldsWithPrefix)
	)

	if v4d.SpatialVolume == nil || v4d.SpatialVolume.Footprint == nil {
		return nil, dsserr.BadRequest("missing geospatial footprint for query")
	}
	cells, err := v4d.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return nil, dsserr.BadRequest(err.Error())
	}
	if len(cells) == 0 {
		return nil, dsserr.BadRequest("missing cell IDs for query")
	}

	cids := make([]int64, len(cells))
	for i, cid := range cells {
		cids[i] = int64(cid)
	}

	result, err := s.fetchOperations(
		ctx, q, operationsIntersectingVolumeQuery,
		pq.Array(cids),
		v4d.SpatialVolume.AltitudeLo,
		v4d.SpatialVolume.AltitudeHi,
		v4d.StartTime,
		v4d.EndTime,
		s.clock.Now(),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Store) SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error) {
	tx, err := s.Begin()
	if err != nil {
		return nil, err
	}

	result, err := s.searchOperations(ctx, tx, v4d, owner)
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
