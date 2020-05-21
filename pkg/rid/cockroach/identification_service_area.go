package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/cockroach"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	dsssql "github.com/interuss/dss/pkg/sql"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var isaFields = "identification_service_areas.id, identification_service_areas.owner, identification_service_areas.url, identification_service_areas.starts_at, identification_service_areas.ends_at, identification_service_areas.updated_at"
var isaFieldsWithoutPrefix = "id, owner, url, starts_at, ends_at, updated_at"

// ISAStore is an implementation of the ISARepo for CRDB.
type ISAStore struct {
	*cockroach.DB

	clock  clockwork.Clock
	logger *zap.Logger
}

func (c *ISAStore) fetchSubscriptionsForNotification(
	ctx context.Context, q dsssql.Queryable, cells []int64) ([]*ridmodels.Subscription, error) {
	var updateQuery = fmt.Sprintf(`
			UPDATE subscriptions
			SET notification_index = notification_index + 1
			WHERE id = ANY(
				SELECT DISTINCT subscription_id
				FROM cells_subscriptions
				WHERE cell_id = ANY($1)
			)
			AND ends_at >= $2
			RETURNING %s`, subscriptionFieldsWithoutPrefix)
	return c.processSubscriptions(
		ctx, q, updateQuery, pq.Array(cells), c.clock.Now())
}

//TODO remove this from this store.. only here to support incrementing sub notification index, but the logic should be placed in application
func (c *ISAStore) processSubscriptions(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*ridmodels.Subscription, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*ridmodels.Subscription
	for rows.Next() {
		s := new(ridmodels.Subscription)

		err := rows.Scan(
			&s.ID,
			&s.Owner,
			&s.URL,
			&s.NotificationIndex,
			&s.StartTime,
			&s.EndTime,
			&s.Version,
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

func (c *ISAStore) process(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*ridmodels.IdentificationServiceArea, error) {
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

func (c *ISAStore) processOne(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*ridmodels.IdentificationServiceArea, error) {
	isas, err := c.process(ctx, q, query, args...)
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

// push creates/updates the IdentificationServiceArea
// identified by "id" and owned by "owner", affecting "cells" in the time
// interval ["starts", "ends"].
//
// Returns the created/updated IdentificationServiceArea and all Subscriptions
// affected by the operation.
func (c *ISAStore) push(ctx context.Context, q dsssql.Queryable, isa *ridmodels.IdentificationServiceArea) (
	*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	var (
		updateAreasQuery = fmt.Sprintf(`
			UPDATE
				identification_service_areas
			SET	(%s) = ($1, $2, $3, $4, $5, transaction_timestamp())
			WHERE id = $1 AND updated_at = $6 
			RETURNING
				%s`, isaFieldsWithoutPrefix, isaFields)
		insertAreasQuery = fmt.Sprintf(`
			INSERT INTO
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
	var err error
	if isa.Version.Empty() {
		isa, err = c.processOne(ctx, q, insertAreasQuery, isa.ID, isa.Owner, isa.URL, isa.StartTime, isa.EndTime)
		if err != nil {
			return nil, nil, err
		}
	} else {
		isa, err = c.processOne(ctx, q, updateAreasQuery, isa.ID, isa.Owner, isa.URL, isa.StartTime, isa.EndTime, isa.Version.ToTimestamp())
		if err != nil {
			return nil, nil, err
		}
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

func (c *ISAStore) getCells(ctx context.Context, id dssmodels.ID) (s2.CellUnion, error) {
	const query = `
	SELECT
		cell_id
	FROM
		cells_identification_service_areas
	WHERE identification_service_area_id = $1`

	rows, err := c.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cell int64
	cells := s2.CellUnion{}

	for rows.Next() {
		if err := rows.Scan(&cell); err != nil {
			return nil, err
		}
		cells = append(cells, s2.CellID(uint64(cell)))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cells, nil
}

// Get returns the isa identified by "id".
func (c *ISAStore) Get(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			identification_service_areas
		WHERE
			id = $1
		AND
			ends_at > $2`, isaFields)
	return c.processOne(ctx, c.DB, query, id, c.clock.Now())
}

// Insert inserts the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
// TODO: Simplify the logic to insert without a query, such that the insert fails
// if there's an existing entity.
func (c *ISAStore) Insert(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	tx, err := c.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer recoverRollbackRepanic(ctx, tx)

	area, subscribers, err := c.push(ctx, tx, isa)
	if err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	return area, subscribers, nil
}

// Update updates the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
// TODO: simplify the logic to just update, without the primary query.
func (c *ISAStore) Update(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	return nil, nil, dsserr.BadRequest("not yet implemented")
}

// Delete deletes the IdentificationServiceArea identified by "id" and owned by "owner".
// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
func (c *ISAStore) Delete(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	var (
		deleteQuery = `
			DELETE FROM
				identification_service_areas
			WHERE
				id = $1
			AND
				owner = $2
			AND
				updated_at = $3
			RETURNING
				*
		`
	)
	// Get the cells since the ISA might not have them set.
	cells, err := c.getCells(ctx, isa.ID)
	if err != nil {
		return nil, nil, err
	}
	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	tx, err := c.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer recoverRollbackRepanic(ctx, tx)

	subscriptions, err := c.fetchSubscriptionsForNotification(ctx, tx, cids)
	if err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	isa, err = c.processOne(ctx, tx, deleteQuery, isa.ID, isa.Owner, isa.Version.ToTimestamp())
	if err != nil {
		return nil, nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return isa, subscriptions, nil
}

// Search searches IdentificationServiceArea
// instances that intersect with "cells" and, if set, the temporal volume
// defined by "earliest" and "latest".
func (c *ISAStore) Search(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	var (
		isasInCellsQuery = fmt.Sprintf(`
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

	return c.process(ctx, c.DB, isasInCellsQuery, pq.Array(cids), earliest,
		latest, c.clock.Now())
}
