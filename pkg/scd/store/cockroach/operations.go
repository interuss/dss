package cockroach

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/lib/pq"
	"github.com/palantir/stacktrace"
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

func (s *repo) fetchOperations(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*scdmodels.Operation, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", query)
	}
	defer rows.Close()

	var payload []*scdmodels.Operation
	for rows.Next() {
		var (
			o         = &scdmodels.Operation{}
			updatedAt time.Time
		)
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Operation row")
		}
		o.OVN = scdmodels.NewOVNFromTime(updatedAt, o.ID.String())
		payload = append(payload, o)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}

	for _, op := range payload {
		if err := s.populateOperationCells(ctx, q, op); err != nil {
			return nil, stacktrace.Propagate(err, "Error populating cells for Operation %s", op.ID)
		}
	}

	return payload, nil
}

func (s *repo) fetchOperation(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*scdmodels.Operation, error) {
	operations, err := s.fetchOperations(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(operations) > 1 {
		return nil, stacktrace.NewError("Query returned %d Operations when only 0 or 1 was expected", len(operations))
	}
	if len(operations) == 0 {
		return nil, nil
	}
	return operations[0], nil
}

func (s *repo) fetchOperationByID(ctx context.Context, q dsssql.Queryable, id dssmodels.ID) (*scdmodels.Operation, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM
			scd_operations
		WHERE
			id = $1`, operationFieldsWithoutPrefix)
	return s.fetchOperation(ctx, q, query, id)
}

func (s *repo) populateOperationCells(ctx context.Context, q dsssql.Queryable, o *scdmodels.Operation) error {
	const query = `
	SELECT
		cell_id
	FROM
		scd_cells_operations
	WHERE operation_id = $1`

	rows, err := q.QueryContext(ctx, query, o.ID)
	if err != nil {
		return stacktrace.Propagate(err, "Error in query: %s", query)
	}
	defer rows.Close()

	var cell int64
	o.Cells = s2.CellUnion{}

	for rows.Next() {
		if err := rows.Scan(&cell); err != nil {
			return stacktrace.Propagate(err, "Error scanning cell ID row")
		}
		o.Cells = append(o.Cells, s2.CellID(uint64(cell)))
	}
	if err := rows.Err(); err != nil {
		return stacktrace.Propagate(err, "Error in rows query result")
	}

	return nil
}

// GetOperation implements repos.Operation.GetOperation.
func (s *repo) GetOperation(ctx context.Context, id dssmodels.ID) (*scdmodels.Operation, error) {
	return s.fetchOperationByID(ctx, s.q, id)
}

// DeleteOperation implements repos.Operation.DeleteOperation.
func (s *repo) DeleteOperation(ctx context.Context, id dssmodels.ID) error {
	var (
		deleteOperationQuery = `
			DELETE FROM
				scd_operations
			WHERE
				id = $1
		`
	)

	res, err := s.q.ExecContext(ctx, deleteOperationQuery, id)
	if err != nil {
		return stacktrace.Propagate(err, "Error in query: %s", deleteOperationQuery)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get RowsAffected")
	}
	if rows == 0 {
		return stacktrace.NewError("Could not delete Operation that does not exist")
	}

	return nil
}

// UpsertOperation implements repos.Operation.UpsertOperation.
func (s *repo) UpsertOperation(ctx context.Context, operation *scdmodels.Operation) (*scdmodels.Operation, error) {
	var (
		upsertOperationsQuery = fmt.Sprintf(`
			UPSERT INTO
				scd_operations
				(%s)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9, transaction_timestamp())
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
	operation, err := s.fetchOperation(ctx, s.q, upsertOperationsQuery,
		operation.ID,
		operation.Owner,
		operation.Version,
		operation.USSBaseURL,
		operation.AltitudeLower,
		operation.AltitudeUpper,
		operation.StartTime,
		operation.EndTime,
		operation.SubscriptionID,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Operation")
	}
	operation.Cells = cells

	for i := range cids {
		if _, err := s.q.ExecContext(ctx, upsertCellsForOperationQuery, cids[i], clevels[i], operation.ID); err != nil {
			return nil, stacktrace.Propagate(err, "Error in query: %s", upsertCellsForOperationQuery)
		}
	}

	if _, err := s.q.ExecContext(ctx, deleteLeftOverCellsForOperationQuery, pq.Array(cids), operation.ID); err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", deleteLeftOverCellsForOperationQuery)
	}

	return operation, nil
}

func (s *repo) searchOperations(ctx context.Context, q dsssql.Queryable, v4d *dssmodels.Volume4D) ([]*scdmodels.Operation, error) {
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
				COALESCE(scd_operations.starts_at <= $5, true)`, operationFieldsWithPrefix)
	)

	if v4d.SpatialVolume == nil || v4d.SpatialVolume.Footprint == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing geospatial footprint for query")
	}
	cells, err := v4d.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to calculate footprint covering")
	}
	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing cell IDs for query")
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
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Operations")
	}

	return result, nil
}

// SearchOperations implements repos.Operation.SearchOperations.
func (s *repo) SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Operation, error) {
	return s.searchOperations(ctx, s.q, v4d)
}

// GetDependentOperations implements repos.Operation.GetDependentOperations.
func (s *repo) GetDependentOperations(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	var dependentOperationsQuery = `
      SELECT
        id
      FROM
        scd_operations
      WHERE
        subscription_id = $1`

	rows, err := s.q.QueryContext(ctx, dependentOperationsQuery, subscriptionID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", dependentOperationsQuery)
	}
	defer rows.Close()
	var opID dssmodels.ID
	var dependentOps []dssmodels.ID
	for rows.Next() {
		err = rows.Scan(&opID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning dependent Operation ID")
		}
		dependentOps = append(dependentOps, opID)
	}

	return dependentOps, nil
}
