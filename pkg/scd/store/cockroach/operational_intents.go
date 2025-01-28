package cockroach

import (
	"context"
	"fmt"
	"strings"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

var (
	operationFieldsWithIndices   [14]string
	operationFieldsWithPrefix    string
	operationFieldsWithoutPrefix string
)

// TODO Update database schema and fields below.
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
	operationFieldsWithIndices[10] = "state"
	operationFieldsWithIndices[11] = "cells"
	operationFieldsWithIndices[12] = "uss_requested_ovn"
	operationFieldsWithIndices[13] = "past_ovns"

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

func (s *repo) fetchOperationalIntents(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*scdmodels.OperationalIntent, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", query)
	}
	defer rows.Close()

	var (
		payload         []*scdmodels.OperationalIntent
		cids            []int64
		ussRequestedOVN pgtype.Text
		pastOVNs        []string
	)
	ussAvailabilities := map[dssmodels.Manager]scdmodels.UssAvailabilityState{}
	for rows.Next() {
		var (
			o         = &scdmodels.OperationalIntent{}
			updatedAt time.Time
		)
		err := rows.Scan(
			&o.ID,
			&o.Manager,
			&o.Version,
			&o.USSBaseURL,
			&o.AltitudeLower,
			&o.AltitudeUpper,
			&o.StartTime,
			&o.EndTime,
			&o.SubscriptionID,
			&updatedAt,
			&o.State,
			&cids,
			&ussRequestedOVN,
			&pastOVNs,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Operation row")
		}

		// If the managing USS has requested a specific OVN on this operational intent, it will be persisted in DB.
		// If not, a default DSS-generated OVN based on the last update time is used.
		// See https://github.com/interuss/dss/issues/1078 for more details.
		if ussRequestedOVN.Valid {
			o.OVN = scdmodels.OVN(ussRequestedOVN.String)
		} else {
			o.OVN = scdmodels.NewOVNFromTime(updatedAt, o.ID.String())
		}

		o.PastOVNs = make([]scdmodels.OVN, 0, len(pastOVNs))
		for _, pastOVN := range pastOVNs {
			o.PastOVNs = append(o.PastOVNs, scdmodels.OVN(pastOVN))
		}

		o.SetCells(cids)
		ussAvailabilities[o.Manager] = scdmodels.UssAvailabilityStateUnknown
		payload = append(payload, o)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}

	for manager := range ussAvailabilities {
		ussAvailability, err := s.GetUssAvailability(ctx, manager)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, stacktrace.Propagate(err, "Error getting USS availability of %s", manager)
		}

		if ussAvailability != nil {
			ussAvailabilities[manager] = ussAvailability.Availability
		}
	}

	for _, op := range payload {
		op.UssAvailability = ussAvailabilities[op.Manager]
	}

	return payload, nil
}

func (s *repo) fetchOperationalIntent(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*scdmodels.OperationalIntent, error) {
	operations, err := s.fetchOperationalIntents(ctx, q, query, args...)
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

func (s *repo) fetchOperationByID(ctx context.Context, q dsssql.Queryable, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM
			scd_operations
		WHERE
			id = $1`, operationFieldsWithoutPrefix)
	uid, err := id.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return s.fetchOperationalIntent(ctx, q, query, uid)
}

// GetOperation implements repos.Operation.GetOperation.
func (s *repo) GetOperationalIntent(ctx context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	return s.fetchOperationByID(ctx, s.q, id)
}

// DeleteOperation implements repos.Operation.DeleteOperation.
func (s *repo) DeleteOperationalIntent(ctx context.Context, id dssmodels.ID) error {
	var (
		deleteOperationQuery = `
			DELETE FROM
				scd_operations
			WHERE
				id = $1
		`
	)

	uid, err := id.PgUUID()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	res, err := s.q.Exec(ctx, deleteOperationQuery, uid)
	if err != nil {
		return stacktrace.Propagate(err, "Error in query: %s", deleteOperationQuery)
	}

	if res.RowsAffected() == 0 {
		return stacktrace.NewError("Could not delete Operation that does not exist")
	}

	return nil
}

// UpsertOperation implements repos.Operation.UpsertOperation.
func (s *repo) UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	var (
		upsertOperationsQuery = fmt.Sprintf(`
			INSERT INTO
				scd_operations
				(%s)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9, transaction_timestamp(), $10, $11, $12, $13)
			ON CONFLICT (%s) DO UPDATE
				SET %s = $2,
					%s = $3,
					%s = $4,
					%s = $5,
					%s = $6,
					%s = $7,
					%s = $8,
					%s = $9,
					%s = transaction_timestamp(),
					%s = $10,
					%s = $11,
					%s = $12,
					%s = $13
				RETURNING
					%s`,
			operationFieldsWithoutPrefix,
			operationFieldsWithIndices[0],
			operationFieldsWithIndices[1],
			operationFieldsWithIndices[2],
			operationFieldsWithIndices[3],
			operationFieldsWithIndices[4],
			operationFieldsWithIndices[5],
			operationFieldsWithIndices[6],
			operationFieldsWithIndices[7],
			operationFieldsWithIndices[8],
			operationFieldsWithIndices[9],
			operationFieldsWithIndices[10],
			operationFieldsWithIndices[11],
			operationFieldsWithIndices[12],
			operationFieldsWithIndices[13],
			operationFieldsWithPrefix,
		)
	)

	cids := make([]int64, len(operation.Cells))
	// TODO clevels not used, can get rid of it?
	clevels := make([]int, len(operation.Cells))

	for i, cell := range operation.Cells {
		cids[i] = int64(cell)
		clevels[i] = cell.Level()
	}

	opid, err := operation.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	subid, err := operation.SubscriptionID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}

	var ussRequestedOVN pgtype.Text // if the OVN is to be generated by the DSS, it must be NULL in DB, not just an empty string
	if operation.OVN != "" {
		ussRequestedOVN = pgtype.Text{
			String: operation.OVN.String(),
			Valid:  true,
		}
	}

	pastOVNs := make([]string, 0, len(operation.PastOVNs))
	for _, pastOVN := range operation.PastOVNs {
		pastOVNs = append(pastOVNs, pastOVN.String())
	}

	operation, err = s.fetchOperationalIntent(ctx, s.q, upsertOperationsQuery,
		opid,
		operation.Manager,
		operation.Version,
		operation.USSBaseURL,
		operation.AltitudeLower,
		operation.AltitudeUpper,
		operation.StartTime,
		operation.EndTime,
		subid,
		operation.State,
		cids,
		ussRequestedOVN,
		pastOVNs,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Operation")
	}

	return operation, nil
}

func (s *repo) searchOperationalIntents(ctx context.Context, q dsssql.Queryable, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	var (
		operationsIntersectingVolumeQuery = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_operations
			WHERE
				cells && $1
			AND
				COALESCE(scd_operations.altitude_upper >= $2, true)
			AND
				COALESCE(scd_operations.altitude_lower <= $3, true)
			AND
				COALESCE(scd_operations.ends_at >= $4, true)
			AND
				COALESCE(scd_operations.starts_at <= $5, true)
			LIMIT $6`, operationFieldsWithPrefix)
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

	result, err := s.fetchOperationalIntents(
		ctx, q, operationsIntersectingVolumeQuery,
		dsssql.CellUnionToCellIds(cells),
		v4d.SpatialVolume.AltitudeLo,
		v4d.SpatialVolume.AltitudeHi,
		v4d.StartTime,
		v4d.EndTime,
		dssmodels.MaxResultLimit,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Operations")
	}

	return result, nil
}

// SearchOperations implements repos.Operation.SearchOperations.
func (s *repo) SearchOperationalIntents(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	return s.searchOperationalIntents(ctx, s.q, v4d)
}

// GetDependentOperations implements repos.Operation.GetDependentOperations.
func (s *repo) GetDependentOperationalIntents(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	var dependentOperationsQuery = `
      SELECT
        id
      FROM
        scd_operations
      WHERE
        subscription_id = $1`

	subid, err := subscriptionID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	rows, err := s.q.Query(ctx, dependentOperationsQuery, subid)
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

// ListExpiredOperationalIntents lists all operational intents older than the threshold.
// Their age is determined by their end time, or by their last update time if they do not have an end time.
func (s *repo) ListExpiredOperationalIntents(ctx context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	expiredOpIntentsQuery := fmt.Sprintf(`
        SELECT
            %s
        FROM
            scd_operations
        WHERE
            scd_operations.ends_at IS NOT NULL AND scd_operations.ends_at <= $1
            OR
            scd_operations.ends_at IS NULL AND scd_operations.updated_at <= $1 -- use last update time as reference if there is no end time
        LIMIT $2`, operationFieldsWithPrefix)

	result, err := s.fetchOperationalIntents(
		ctx, s.q, expiredOpIntentsQuery,
		threshold,
		dssmodels.MaxResultLimit,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Operations")
	}

	return result, nil
}
