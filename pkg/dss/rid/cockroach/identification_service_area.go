package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dssmodels "github.com/interuss/dss/pkg/dss/models"
	ridmodels "github.com/interuss/dss/pkg/dss/rid/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	dsssql "github.com/interuss/dss/pkg/sql"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var isaFields = "identification_service_areas.id, identification_service_areas.owner, identification_service_areas.url, identification_service_areas.starts_at, identification_service_areas.ends_at, identification_service_areas.updated_at"
var isaFieldsWithoutPrefix = "id, owner, url, starts_at, ends_at, updated_at"

func recoverRollbackRepanic(ctx context.Context, tx *sql.Tx) {
	if p := recover(); p != nil {
		if err := tx.Rollback(); err != nil {
			logging.WithValuesFromContext(ctx, logging.Logger).Error(
				"failed to rollback transaction", zap.Error(err),
			)
		}
	}
}

func (c *Store) fetchISAs(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*ridmodels.IdentificationServiceArea, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*ridmodels.IdentificationServiceArea
	for rows.Next() {
		i := new(ridmodels.IdentificationServiceArea)

		err := rows.Scan(
			&i.ID,
			&i.Owner,
			&i.URL,
			&i.StartTime,
			&i.EndTime,
			&i.Version,
		)
		if err != nil {
			return nil, err
		}
		payload = append(payload, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *Store) fetchISA(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*ridmodels.IdentificationServiceArea, error) {
	isas, err := c.fetchISAs(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(isas) > 1 {
		return nil, multierr.Combine(err, fmt.Errorf("query returned %d identification_service_areas", len(isas)))
	}
	if len(isas) == 0 {
		return nil, sql.ErrNoRows
	}
	return isas[0], nil
}

func (c *Store) fetchISAByID(ctx context.Context, q dsssql.Queryable, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			identification_service_areas
		WHERE
			id = $1
		AND
			ends_at >= $2`, isaFields)
	return c.fetchISA(ctx, q, query, id, c.clock.Now())
}

func (c *Store) populateISACells(ctx context.Context, q dsssql.Queryable, i *ridmodels.IdentificationServiceArea) error {
	const query = `
	SELECT
		cell_id
	FROM
		cells_identification_service_areas
	WHERE identification_service_area_id = $1`

	rows, err := q.QueryContext(ctx, query, i.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var cell int64
	i.Cells = s2.CellUnion{}

	for rows.Next() {
		if err := rows.Scan(&cell); err != nil {
			return err
		}
		i.Cells = append(i.Cells, s2.CellID(uint64(cell)))
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

// pushISA creates/updates the IdentificationServiceArea
// identified by "id" and owned by "owner", affecting "cells" in the time
// interval ["starts", "ends"].
//
// Returns the created/updated IdentificationServiceArea and all Subscriptions
// affected by the operation.
func (c *Store) pushISA(ctx context.Context, q dsssql.Queryable, isa *ridmodels.IdentificationServiceArea) (
	*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	var (
		upsertAreasQuery = fmt.Sprintf(`
			UPSERT INTO
				identification_service_areas
				(%s)
			VALUES
				($1, $2, $3, $4, $5, transaction_timestamp())
			RETURNING
				%s`, isaFieldsWithoutPrefix, isaFields)
		upsertCellsForAreaQuery = `
			UPSERT INTO
				cells_identification_service_areas
				(cell_id, cell_level, identification_service_area_id)
			VALUES
				($1, $2, $3)`
		deleteLeftOverCellsForAreaQuery = `
			DELETE FROM
				cells_identification_service_areas
			WHERE
				cell_id != ALL($1)
			AND
				identification_service_area_id = $2`
	)

	cids := make([]int64, len(isa.Cells))
	clevels := make([]int, len(isa.Cells))

	for i, cell := range isa.Cells {
		cids[i] = int64(cell)
		clevels[i] = cell.Level()
	}

	cells := isa.Cells
	isa, err := c.fetchISA(ctx, q, upsertAreasQuery, isa.ID, isa.Owner, isa.URL, isa.StartTime, isa.EndTime)
	if err != nil {
		return nil, nil, err
	}
	isa.Cells = cells

	for i := range cids {
		if _, err := q.ExecContext(ctx, upsertCellsForAreaQuery, cids[i], clevels[i], isa.ID); err != nil {
			return nil, nil, err
		}
	}

	if _, err := q.ExecContext(ctx, deleteLeftOverCellsForAreaQuery, pq.Array(cids), isa.ID); err != nil {
		return nil, nil, err
	}

	subscriptions, err := c.fetchSubscriptionsForNotification(ctx, q, cids)
	if err != nil {
		return nil, nil, err
	}

	return isa, subscriptions, nil
}

// GetISA returns the isa identified by "id".
func (c *Store) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	return c.fetchISAByID(ctx, c.DB, id)
}

// InsertISA inserts the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
func (c *Store) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	tx, err := c.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer recoverRollbackRepanic(ctx, tx)

	old, err := c.fetchISAByID(ctx, tx, isa.ID)
	switch {
	case err == sql.ErrNoRows:
		break
	case err != nil:
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	switch {
	case old == nil && !isa.Version.Empty():
		// The user wants to update an existing ISA, but one wasn't found.
		return nil, nil, multierr.Combine(dsserr.NotFound(isa.ID.String()), tx.Rollback())
	case old != nil && isa.Version.Empty():
		// The user wants to create a new ISA but it already exists.
		return nil, nil, multierr.Combine(dsserr.AlreadyExists(isa.ID.String()), tx.Rollback())
	case old != nil && !isa.Version.Matches(old.Version):
		// The user wants to update an ISA but the version doesn't match.
		return nil, nil, multierr.Combine(dsserr.VersionMismatch("old version"), tx.Rollback())
	case old != nil && old.Owner != isa.Owner:
		return nil, nil, multierr.Combine(dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner)), tx.Rollback())
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(c.clock.Now(), old); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	area, subscribers, err := c.pushISA(ctx, tx, isa)
	if err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return area, subscribers, nil
}

// DeleteISA deletes the IdentificationServiceArea identified by "id" and owned by "owner".
// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
func (c *Store) DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	var (
		deleteQuery = `
			DELETE FROM
				identification_service_areas
			WHERE
				id = $1
			AND
				owner = $2
			RETURNING
				*
		`
	)

	tx, err := c.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer recoverRollbackRepanic(ctx, tx)

	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := c.fetchISAByID(ctx, tx, id)
	switch {
	case err == sql.ErrNoRows: // Return a 404 here.
		return nil, nil, multierr.Combine(dsserr.NotFound(id.String()), tx.Rollback())
	case err != nil:
		return nil, nil, multierr.Combine(err, tx.Rollback())
	case !version.Empty() && !version.Matches(old.Version):
		return nil, nil, multierr.Combine(dsserr.VersionMismatch("old version"), tx.Rollback())
	case old != nil && old.Owner != owner:
		return nil, nil, multierr.Combine(dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner)), tx.Rollback())
	}
	if err := c.populateISACells(ctx, tx, old); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	cids := make([]int64, len(old.Cells))
	for i, cell := range old.Cells {
		cids[i] = int64(cell)
	}
	subscriptions, err := c.fetchSubscriptionsForNotification(ctx, tx, cids)
	if err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if _, err := tx.ExecContext(ctx, deleteQuery, id, owner); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return old, subscriptions, nil
}

// SearchISAs searches IdentificationServiceArea
// instances that intersect with "cells" and, if set, the temporal volume
// defined by "earliest" and "latest".
func (c *Store) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	var (
		serviceAreasInCellsQuery = fmt.Sprintf(`
			SELECT
				%s
			FROM
				identification_service_areas
			JOIN
				(SELECT DISTINCT
					cells_identification_service_areas.identification_service_area_id
				FROM
					cells_identification_service_areas
				WHERE
					cells_identification_service_areas.cell_id = ANY($1)
				)
			AS
				unique_identification_service_areas
			ON
				identification_service_areas.id = unique_identification_service_areas.identification_service_area_id
			WHERE
				COALESCE(identification_service_areas.ends_at >= $2, true)
			AND
				COALESCE(identification_service_areas.starts_at <= $3, true)
			AND
				identification_service_areas.ends_at >= $4`, isaFields)
	)

	if len(cells) == 0 {
		return nil, dsserr.BadRequest("missing cell IDs for query")
	}

	cids := make([]int64, len(cells))
	for i, cid := range cells {
		cids[i] = int64(cid)
	}

	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}
	defer recoverRollbackRepanic(ctx, tx)

	result, err := c.fetchISAs(
		ctx, tx, serviceAreasInCellsQuery, pq.Array(cids), earliest, latest,
		c.clock.Now())
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
