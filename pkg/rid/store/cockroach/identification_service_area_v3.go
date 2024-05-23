package cockroach

import (
	"context"
	"fmt"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"

	"github.com/golang/geo/s2"
	"go.uber.org/zap"
)

const (
	isaFieldsV3       = "id, owner, url, cells, starts_at, ends_at, updated_at"
	updateISAFieldsV3 = "id, url, cells, starts_at, ends_at, updated_at"
)

// The purpose od isaRepoV3 is solely to support backwards compatibility
// It will be deleted from the codebase when all existing production deployments have been upgraded to 3.1.0+.
type isaRepoV3 struct {
	dssql.Queryable

	logger *zap.Logger
}

func (c *isaRepoV3) process(ctx context.Context, query string, args ...interface{}) ([]*ridmodels.IdentificationServiceArea, error) {
	rows, err := c.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, fmt.Sprintf("Error in query: %s", query))
	}
	defer rows.Close()

	var payload []*ridmodels.IdentificationServiceArea
	var cids []int64

	for rows.Next() {
		i := new(ridmodels.IdentificationServiceArea)

		var updateTime time.Time

		err := rows.Scan(
			&i.ID,
			&i.Owner,
			&i.URL,
			&cids,
			&i.StartTime,
			&i.EndTime,
			&updateTime,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning ISA row")
		}
		i.SetCells(cids)
		i.Version = dssmodels.VersionFromTime(updateTime)
		payload = append(payload, i)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}

	return payload, nil
}

func (c *isaRepoV3) processOne(ctx context.Context, query string, args ...interface{}) (*ridmodels.IdentificationServiceArea, error) {
	isas, err := c.process(ctx, query, args...)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}
	if len(isas) > 1 {
		return nil, stacktrace.NewError("Query returned %d identification_service_areas when only 0 or 1 was expected", len(isas))
	}
	if len(isas) == 0 {
		return nil, nil
	}
	return isas[0], nil
}

// GetISA returns the isa identified by "id".
// Returns nil, nil if not found
func (c *isaRepoV3) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			identification_service_areas
		WHERE
			id = $1`, isaFieldsV3)
	uid, err := id.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, query, uid)
}

// InsertISA inserts the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
// TODO: Simplify the logic to insert without a query, such that the insert fails
// if there's an existing entity.
func (c *isaRepoV3) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		insertAreasQuery = fmt.Sprintf(`
			INSERT INTO
				identification_service_areas
				(%s)
			VALUES
				($1, $2, $3, $4, $5, $6, transaction_timestamp())
			RETURNING
				%s`, isaFieldsV3, isaFieldsV3)
	)

	cids, err := dssql.CellUnionToCellIdsWithValidation(isa.Cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}
	uid, err := isa.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, insertAreasQuery, uid, isa.Owner, isa.URL, cids, isa.StartTime, isa.EndTime)
}

// UpdateISA updates the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
// TODO: simplify the logic to just update, without the primary query.
// Returns nil, nil if ID, version not found
func (c *isaRepoV3) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		updateAreasQuery = fmt.Sprintf(`
			UPDATE
				identification_service_areas
			SET	(%s) = ($1, $2, $3, $4, $5, transaction_timestamp())
			WHERE id = $1 AND updated_at = $6
			RETURNING
				%s`, updateISAFieldsV3, isaFieldsV3)
	)

	cids, err := dssql.CellUnionToCellIdsWithValidation(isa.Cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	uid, err := isa.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, updateAreasQuery, uid, isa.URL, cids, isa.StartTime, isa.EndTime, isa.Version.ToTimestamp())
}

// DeleteISA deletes the IdentificationServiceArea identified by "id" and owned by "owner".
// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
// Returns nil, nil if ID, version not found
func (c *isaRepoV3) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		deleteQuery = fmt.Sprintf(`
			DELETE FROM
				identification_service_areas
			WHERE
				id = $1
			AND
				updated_at = $2
			RETURNING %s`, isaFieldsV3)
	)
	uid, err := isa.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, deleteQuery, uid, isa.Version.ToTimestamp())
}

// SearchISAs searches IdentificationServiceArea
// instances that intersect with "cells" and, if set, the temporal volume
// defined by "earliest" and "latest".
func (c *isaRepoV3) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	var (
		// TODO: make earliest and latest required (NOT NULL) and remove coalesce.
		// Make them real values (not pointers), on the model layer.
		isasInCellsQuery = fmt.Sprintf(`
			SELECT
				%s
			FROM
				identification_service_areas
			WHERE
				ends_at >= $1
			AND
				COALESCE(starts_at <= $2, true)
			AND
				cells && $3
			LIMIT $4`, isaFieldsV3)
	)

	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing cell IDs for query")
	}

	if earliest == nil {
		return nil, stacktrace.NewError("Earliest start time is missing")
	}

	return c.process(ctx, isasInCellsQuery, earliest, latest, dssql.CellUnionToCellIds(cells), dssmodels.MaxResultLimit)
}

// ListExpiredISAs returns empty. We don't support thi function in store v3.0 because db doesn't have 'writer' field.
func (c *isaRepoV3) ListExpiredISAs(ctx context.Context, writer string) ([]*ridmodels.IdentificationServiceArea, error) {
	return make([]*ridmodels.IdentificationServiceArea, 0), nil
}
