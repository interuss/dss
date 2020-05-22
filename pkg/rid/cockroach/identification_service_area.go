package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	isaFields = "id, owner, url, cells, starts_at, ends_at, updated_at"
)

// ISAStore is an implementation of the ISARepo for CRDB.
type ISAStore struct {
	*cockroach.DB

	clock  clockwork.Clock
	logger *zap.Logger
}

func (c *ISAStore) process(ctx context.Context, query string, args ...interface{}) ([]*ridmodels.IdentificationServiceArea, error) {
	rows, err := c.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*ridmodels.IdentificationServiceArea
	cids := pq.Int64Array{}

	for rows.Next() {
		i := new(ridmodels.IdentificationServiceArea)

		err := rows.Scan(
			&i.ID,
			&i.Owner,
			&i.URL,
			&cids,
			&i.StartTime,
			&i.EndTime,
			&i.Version,
		)
		if err != nil {
			return nil, err
		}
		i.SetCells(cids)
		payload = append(payload, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *ISAStore) processOne(ctx context.Context, query string, args ...interface{}) (*ridmodels.IdentificationServiceArea, error) {
	isas, err := c.process(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(isas) > 1 {
		return nil, fmt.Errorf("query returned %d identification_service_areas", len(isas))
	}
	if len(isas) == 0 {
		return nil, sql.ErrNoRows
	}
	return isas[0], nil
}

// GetISA returns the isa identified by "id".
func (c *ISAStore) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	var query = fmt.Sprintf(`
		SELECT %s FROM
			identification_service_areas
		WHERE
			id = $1
		AND
			ends_at > $2`, isaFields)
	return c.processOne(ctx, query, id, c.clock.Now())
}

// InsertISA inserts the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
// TODO: Simplify the logic to insert without a query, such that the insert fails
// if there's an existing entity.
func (c *ISAStore) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		updateAreasQuery = fmt.Sprintf(`
			UPDATE
				identification_service_areas
			SET	(%s) = ($1, $2, $3, $4, $5, $6, transaction_timestamp())
			WHERE id = $1 AND updated_at = $7
			RETURNING
				%s`, isaFields, isaFields)
		insertAreasQuery = fmt.Sprintf(`
			INSERT INTO
				identification_service_areas
				(%s)
			VALUES
				($1, $2, $3, $4, $5, $6, transaction_timestamp())
			RETURNING
				%s`, isaFields, isaFields)
	)

	cids := make([]int64, len(isa.Cells))

	for i, cell := range isa.Cells {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, err
		}
		cids[i] = int64(cell)
	}

	var err error
	if isa.Version.Empty() {
		ret, err := c.processOne(ctx, insertAreasQuery, isa.ID, isa.Owner, isa.URL, pq.Array(cids), isa.StartTime, isa.EndTime)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}

	ret, err := c.processOne(ctx, updateAreasQuery, isa.ID, isa.Owner, isa.URL, pq.Array(cids), isa.StartTime, isa.EndTime, isa.Version.ToTimestamp())
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// UpdateISA updates the IdentificationServiceArea identified by "id" and owned
// by "owner", affecting "cells" in the time interval ["starts", "ends"].
//
// Returns the created IdentificationServiceArea and all Subscriptions affected
// by it.
// TODO: simplify the logic to just update, without the primary query.
func (c *ISAStore) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// DeleteISA deletes the IdentificationServiceArea identified by "id" and owned by "owner".
// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
func (c *ISAStore) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	var (
		deleteQuery = fmt.Sprintf(`
			DELETE FROM
				identification_service_areas
			WHERE
				id = $1
			AND
				owner = $2
			AND
				updated_at = $3
			RETURNING %s`, isaFields)
	)
	// Get the cells since the ISA might not have them set.
	cids := make([]int64, len(isa.Cells))
	for i, cell := range isa.Cells {
		cids[i] = int64(cell)
	}

	return c.processOne(ctx, deleteQuery, isa.ID, isa.Owner, isa.Version.ToTimestamp())
}

// SearchISAs searches IdentificationServiceArea
// instances that intersect with "cells" and, if set, the temporal volume
// defined by "earliest" and "latest".
func (c *ISAStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
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
		return nil, dsserr.BadRequest("missing cell IDs for query")
	}

	if earliest == nil {
		return nil, dsserr.Internal("must call with an earliest start time.")
	}

	cids := make([]int64, len(cells))
	for i, cid := range cells {
		cids[i] = int64(cid)
	}

	return c.process(ctx, isasInCellsQuery, earliest, latest, pq.Int64Array(cids))
}
