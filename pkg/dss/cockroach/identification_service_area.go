package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"
	"go.uber.org/multierr"
)

var isaFields = "identification_service_areas.id, identification_service_areas.owner, identification_service_areas.url, identification_service_areas.starts_at, identification_service_areas.ends_at, identification_service_areas.updated_at"
var isaFieldsWithoutPrefix = "id, owner, url, starts_at, ends_at, updated_at"

func (c *Store) fetchISAs(ctx context.Context, q queryable, query string, args ...interface{}) ([]*models.IdentificationServiceArea, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*models.IdentificationServiceArea
	for rows.Next() {
		i := new(models.IdentificationServiceArea)

		err := rows.Scan(
			&i.ID,
			&i.Owner,
			&i.Url,
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

func (c *Store) fetchISA(ctx context.Context, q queryable, query string, args ...interface{}) (*models.IdentificationServiceArea, error) {
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

func (c *Store) fetchISAByID(ctx context.Context, q queryable, id models.ID) (*models.IdentificationServiceArea, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			identification_service_areas
		WHERE
			id = $1`, isaFields)
	return c.fetchISA(ctx, q, query, id)
}

func (c *Store) fetchISAByIDAndOwner(ctx context.Context, q queryable, id models.ID, owner models.Owner) (*models.IdentificationServiceArea, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			identification_service_areas
		WHERE
			id = $1
		AND
			owner = $2`, isaFields)
	return c.fetchISA(ctx, q, query, id, owner)
}

func (c *Store) populateISACells(ctx context.Context, q queryable, i *models.IdentificationServiceArea) error {
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
func (c *Store) pushISA(ctx context.Context, q queryable, isa *models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error) {
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
	isa, err := c.fetchISA(ctx, q, upsertAreasQuery, isa.ID, isa.Owner, isa.Url, isa.StartTime, isa.EndTime)
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

	subscriptions, err := c.fetchSubscriptionsByCells(ctx, q, cids)
	if err != nil {
		return nil, nil, err
	}

	return isa, subscriptions, nil
}

// Get returns the isa identified by "id".
func (c *Store) GetISA(ctx context.Context, id models.ID) (*models.IdentificationServiceArea, error) {
	return c.fetchISAByID(ctx, c.DB, id)
}

// InsertISA inserts the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
func (c *Store) InsertISA(ctx context.Context, isa models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error) {
	tx, err := c.Begin()
	if err != nil {
		return nil, nil, err
	}

	old, err := c.fetchISAByIDAndOwner(ctx, tx, isa.ID, isa.Owner)
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
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(c.clock.Now(), old); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	area, subscribers, err := c.pushISA(ctx, tx, &isa)
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
func (c *Store) DeleteISA(ctx context.Context, id models.ID, owner models.Owner, version *models.Version) (*models.IdentificationServiceArea, []*models.Subscription, error) {
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

	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := c.fetchISAByIDAndOwner(ctx, tx, id, owner)
	switch {
	case err == sql.ErrNoRows: // Return a 404 here.
		return nil, nil, multierr.Combine(dsserr.NotFound(id.String()), tx.Rollback())
	case err != nil:
		return nil, nil, multierr.Combine(err, tx.Rollback())
	case !version.Empty() && !version.Matches(old.Version):
		return nil, nil, multierr.Combine(dsserr.VersionMismatch("old version"), tx.Rollback())
	}
	if err := c.populateISACells(ctx, tx, old); err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	cids := make([]int64, len(old.Cells))
	for i, cell := range old.Cells {
		cids[i] = int64(cell)
	}
	subscriptions, err := c.fetchSubscriptionsByCells(ctx, tx, cids)
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
func (c *Store) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*models.IdentificationServiceArea, error) {
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
				COALESCE(identification_service_areas.starts_at >= $2, true)
			AND
				COALESCE(identification_service_areas.ends_at <= $3, true)`, isaFields)
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

	result, err := c.fetchISAs(ctx, tx, serviceAreasInCellsQuery, pq.Array(cids), earliest, latest)
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
