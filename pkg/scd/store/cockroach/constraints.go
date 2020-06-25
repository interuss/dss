package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/multierr"
)

var (
	constraintFieldsWithIndices   [9]string
	constraintFieldsWithPrefix    string
	constraintFieldsWithoutPrefix string
)

func init() {
	constraintFieldsWithIndices[0] = "id"
	constraintFieldsWithIndices[1] = "owner"
	constraintFieldsWithIndices[2] = "version"
	constraintFieldsWithIndices[3] = "url"
	constraintFieldsWithIndices[4] = "altitude_lower"
	constraintFieldsWithIndices[5] = "altitude_upper"
	constraintFieldsWithIndices[6] = "starts_at"
	constraintFieldsWithIndices[7] = "ends_at"
	constraintFieldsWithIndices[8] = "updated_at"

	constraintFieldsWithoutPrefix = strings.Join(
		constraintFieldsWithIndices[:], ",",
	)

	withPrefix := make([]string, 9)
	for idx, field := range constraintFieldsWithIndices {
		withPrefix[idx] = "scd_constraints." + field
	}

	constraintFieldsWithPrefix = strings.Join(
		withPrefix[:], ",",
	)
}

func (c *repo) fetchConstraints(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*scdmodels.Constraint, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*scdmodels.Constraint
	for rows.Next() {
		var (
			c         = new(scdmodels.Constraint)
			updatedAt time.Time
		)
		err := rows.Scan(
			&c.ID,
			&c.Owner,
			&c.Version,
			&c.USSBaseURL,
			&c.AltitudeLower,
			&c.AltitudeUpper,
			&c.StartTime,
			&c.EndTime,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.OVN = scdmodels.NewOVNFromTime(updatedAt, c.ID.String())
		payload = append(payload, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *repo) fetchConstraint(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*scdmodels.Constraint, error) {
	constraints, err := c.fetchConstraints(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(constraints) > 1 {
		return nil, multierr.Combine(err, fmt.Errorf("query returned %d constraints", len(constraints)))
	}
	if len(constraints) == 0 {
		return nil, sql.ErrNoRows
	}
	return constraints[0], nil
}

// Implements scd.repos.Constraint.GetConstraint
func (c *repo) GetConstraint(ctx context.Context, id scdmodels.ID) (*scdmodels.Constraint, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_constraints
			WHERE
				id = $1`, constraintFieldsWithoutPrefix)
	)
	return c.fetchConstraint(ctx, c.q, query, id)
}

// Implements scd.repos.Constraint.GetConstraintCells
func (c *repo) GetConstraintCells(ctx context.Context, id scdmodels.ID) (s2.CellUnion, error) {
	var (
		cellsQuery = `
			SELECT
				cell_id
			FROM
				scd_cells_constraints
			WHERE
				constraint_id = $1
		`
	)

	rows, err := c.q.QueryContext(ctx, cellsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("GetConstraintCells Query error: %s", err)
	}
	defer rows.Close()

	var (
		cu   s2.CellUnion
		cidi int64
	)
	for rows.Next() {
		if err := rows.Scan(&cidi); err != nil {
			return nil, fmt.Errorf("GetConstraintCells row scan error: %s", err)
		}
		cu = append(cu, s2.CellID(cidi))
	}

	return cu, rows.Err()
}

// Implements scd.repos.Constraint.UpsertConstraint
func (c *repo) UpsertConstraint(ctx context.Context, s *scdmodels.Constraint, cells s2.CellUnion) (*scdmodels.Constraint, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		UPSERT INTO
		  scd_constraints
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, transaction_timestamp())
		RETURNING
			%s`, constraintFieldsWithoutPrefix, constraintFieldsWithPrefix)

		constraintCellQuery = `
		UPSERT INTO
			scd_cells_constraints
			(cell_id, cell_level, constraint_id)
		VALUES
			($1, $2, $3)
		`

		deleteLeftOverCellsForConstraintQuery = `
			DELETE FROM
				scd_cells_constraints
			WHERE
				cell_id != ALL($1)
			AND
				constraint_id = $2`
	)

	cids := make([]int64, len(cells))
	clevels := make([]int, len(cells))

	for i, cell := range cells {
		cids[i] = int64(cell)
		clevels[i] = cell.Level()
	}

	s, err := c.fetchConstraint(ctx, c.q, upsertQuery,
		s.ID,
		s.Owner,
		s.Version,
		s.USSBaseURL,
		s.AltitudeLower,
		s.AltitudeUpper,
		s.StartTime,
		s.EndTime)
	if err != nil {
		return nil, err
	}

	for i := range cids {
		if _, err := c.q.ExecContext(ctx, constraintCellQuery, cids[i], clevels[i], s.ID); err != nil {
			return nil, err
		}
	}

	if _, err := c.q.ExecContext(ctx, deleteLeftOverCellsForConstraintQuery, pq.Array(cids), s.ID); err != nil {
		return nil, err
	}

	return s, nil
}

// Implements scd.repos.Constraint.DeleteConstraint
func (c *repo) DeleteConstraint(ctx context.Context, id scdmodels.ID) error {
	const (
		query = `
		DELETE FROM
			scd_constraints
		WHERE
			id = $1`
	)

	res, err := c.q.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Implements scd.repos.Constraint.SearchConstraints
func (c *repo) SearchConstraints(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				scd_constraints
			JOIN
				(SELECT DISTINCT
					scd_cells_constraints.constraint_id
				FROM
					scd_cells_constraints
				WHERE
					scd_cells_constraints.cell_id = ANY($1)
				)
			AS
				unique_constraint_ids
			ON
				scd_constraints.id = unique_constraint_ids.constraint_id
			WHERE
				COALESCE(starts_at <= $3, true)
			AND
				COALESCE(ends_at >= $2, true)`, constraintFieldsWithPrefix)
	)

	// TODO: Lazily calculate & cache spatial covering so that it is only ever
	// computed once on a particular Volume4D
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, err
	}

	if len(cells) == 0 {
		return []*scdmodels.Constraint{}, nil
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	constraints, err := c.fetchConstraints(
		ctx, c.q, query, pq.Array(cids), v4d.StartTime, v4d.EndTime)
	if err != nil {
		return nil, err
	}

	return constraints, nil
}
