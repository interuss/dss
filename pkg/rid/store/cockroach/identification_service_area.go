package cockroach

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/jackc/pgtype"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"

	"github.com/golang/geo/s2"
	repos "github.com/interuss/dss/pkg/rid/repos"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

const (
	isaFields       = "id, owner, url, cells, starts_at, ends_at, writer, updated_at"
	updateISAFields = "id, url, cells, starts_at, ends_at, writer, updated_at"
)

func NewISARepo(ctx context.Context, db dssql.Queryable, dbVersion semver.Version, logger *zap.Logger) repos.ISA {
	if dbVersion.Compare(v400) >= 0 {
		return &isaRepo{
			Queryable: db,
			logger:    logger,
		}
	}
	return &isaRepoV3{
		Queryable: db,
		logger:    logger,
	}
}

// isaRepo is an implementation of the ISARepo for CRDB.
type isaRepo struct {
	dssql.Queryable

	logger *zap.Logger
}

func (c *isaRepo) process(ctx context.Context, query string, args ...interface{}) ([]*ridmodels.IdentificationServiceArea, error) {
	rows, err := c.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, fmt.Sprintf("Error in query: %s", query))
	}
	defer rows.Close()

	var payload []*ridmodels.IdentificationServiceArea
	pgCids := pgtype.Int8Array{}

	var writer pgtype.Varchar
	for rows.Next() {
		i := new(ridmodels.IdentificationServiceArea)

		var updateTime time.Time
		if i.StartTime.IsZero() {
			*i.StartTime = time.Now()
		}

		err := rows.Scan(
			&i.ID,
			&i.Owner,
			&i.URL,
			&pgCids,
			&i.StartTime,
			&i.EndTime,
			&writer,
			&updateTime,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning ISA row")
		}
		i.Writer = writer.String
		var cids []int64
		if err := pgCids.AssignTo(&cids); err != nil {
			return nil, stacktrace.Propagate(err, "Error Converting jackc/pgtype to array")
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

func (c *isaRepo) processOne(ctx context.Context, query string, args ...interface{}) (*ridmodels.IdentificationServiceArea, error) {
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
func (c *isaRepo) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			identification_service_areas
		WHERE
			id = $1`, isaFields)
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
func (c *isaRepo) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		insertAreasQuery = fmt.Sprintf(`
			INSERT INTO
				identification_service_areas
				(%s)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, transaction_timestamp())
			RETURNING
				%s`, isaFields, isaFields)
	)

	cids := make([]int64, len(isa.Cells))

	for i, cell := range isa.Cells {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, stacktrace.Propagate(err, "Error validating cell")
		}
		cids[i] = int64(cell)
	}
	var pgCids pgtype.Int8Array
	if err := pgCids.Set(cids); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}
	id, err := isa.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, insertAreasQuery, id, isa.Owner, isa.URL, pgCids, isa.StartTime, isa.EndTime, isa.Writer)

}

// UpdateISA updates the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
// TODO: simplify the logic to just update, without the primary query.
// Returns nil, nil if ID, version not found
func (c *isaRepo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		updateAreasQuery = fmt.Sprintf(`
			UPDATE
				identification_service_areas
			SET	(%s) = ($1, $2, $3, $4, $5, $7, transaction_timestamp())
			WHERE id = $1 AND updated_at = $6
			RETURNING
				%s`, updateISAFields, isaFields)
	)

	cids := make([]int64, len(isa.Cells))

	for i, cell := range isa.Cells {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, stacktrace.Propagate(err, "Error validating cell")
		}
		cids[i] = int64(cell)
	}

	var pgCids pgtype.Int8Array
	err := pgCids.Set(cids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}
	id, err := isa.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, updateAreasQuery, id, isa.URL, pgCids, isa.StartTime, isa.EndTime, isa.Version.ToTimestamp(), isa.Writer)
}

// DeleteISA deletes the IdentificationServiceArea identified by "id" and owned by "owner".
// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
// Returns nil, nil if ID, version not found
func (c *isaRepo) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		deleteQuery = fmt.Sprintf(`
			DELETE FROM
				identification_service_areas
			WHERE
				id = $1
			AND
				updated_at = $2
			RETURNING %s`, isaFields)
	)
	id, err := isa.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, deleteQuery, id, isa.Version.ToTimestamp())
}

// SearchISAs searches IdentificationServiceArea
// instances that intersect with "cells" and, if set, the temporal volume
// defined by "earliest" and "latest".
func (c *isaRepo) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
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
				cells && $3`, isaFields)
	)

	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing cell IDs for query")
	}

	if earliest == nil {
		return nil, stacktrace.NewError("Earliest start time is missing")
	}

	cids := make([]int64, len(cells))
	for i, cid := range cells {
		cids[i] = int64(cid)
	}

	var pgCids pgtype.Int8Array
	err := pgCids.Set(cids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	return c.process(ctx, isasInCellsQuery, earliest, latest, pgCids)
}

// ListExpiredISAs lists all expired ISAs based on writer.
// Records expire if current time is <expiredDurationInMin> minutes more than records' endTime.
// The function queries both empty writer and null writer when passing empty string as a writer.
func (c *isaRepo) ListExpiredISAs(ctx context.Context, writer string) ([]*ridmodels.IdentificationServiceArea, error) {
	writerQuery := "'" + writer + "'"
	if len(writer) == 0 {
		writerQuery = "'' OR writer = NULL"
	}

	var (
		isasInCellsQuery = fmt.Sprintf(`
	SELECT
		%s
	FROM
		identification_service_areas
	WHERE
		ends_at + INTERVAL '%d' MINUTE <= CURRENT_TIMESTAMP
	AND
		(writer = %s)`, isaFields, expiredDurationInMin, writerQuery)
	)

	return c.process(ctx, isasInCellsQuery)
}
