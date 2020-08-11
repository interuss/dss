package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"

	"github.com/lib/pq"
	"github.com/palantir/stacktrace"
)

const (
	nConstraintFields = 10
)

var (
	constraintFieldsWithIndices   [nConstraintFields]string
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
	constraintFieldsWithIndices[8] = "cells"
	constraintFieldsWithIndices[9] = "updated_at"

	constraintFieldsWithoutPrefix = strings.Join(
		constraintFieldsWithIndices[:], ",",
	)

	withPrefix := make([]string, nConstraintFields)
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
		return nil, stacktrace.Propagate(err, "Error in query: %s", query)
	}
	defer rows.Close()

	var payload []*scdmodels.Constraint
	cids := pq.Int64Array{}
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
			&cids,
			&updatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Constraint row")
		}
		c.Cells = geo.CellUnionFromInt64(cids)
		c.OVN = scdmodels.NewOVNFromTime(updatedAt, c.ID.String())
		payload = append(payload, c)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}
	return payload, nil
}

func (c *repo) fetchConstraint(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*scdmodels.Constraint, error) {
	constraints, err := c.fetchConstraints(ctx, q, query, args...)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}
	if len(constraints) > 1 {
		return nil, stacktrace.NewError("Query returned %d Constraints when only 0 or 1 was expected", len(constraints))
	}
	if len(constraints) == 0 {
		return nil, sql.ErrNoRows
	}
	return constraints[0], nil
}

// Implements scd.repos.Constraint.GetConstraint
func (c *repo) GetConstraint(ctx context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
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

// Implements scd.repos.Constraint.UpsertConstraint
func (c *repo) UpsertConstraint(ctx context.Context, s *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		UPSERT INTO
		  scd_constraints
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, transaction_timestamp())
		RETURNING
			%s`, constraintFieldsWithoutPrefix, constraintFieldsWithPrefix)
	)

	cids := make([]int64, len(s.Cells))

	for i, cell := range s.Cells {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, stacktrace.Propagate(err, "Error validating cell")
		}
		cids[i] = int64(cell)
	}

	s, err := c.fetchConstraint(ctx, c.q, upsertQuery,
		s.ID,
		s.Owner,
		s.Version,
		s.USSBaseURL,
		s.AltitudeLower,
		s.AltitudeUpper,
		s.StartTime,
		s.EndTime,
		pq.Int64Array(cids))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Constraint")
	}

	return s, nil
}

// Implements scd.repos.Constraint.DeleteConstraint
func (c *repo) DeleteConstraint(ctx context.Context, id dssmodels.ID) error {
	const (
		query = `
		DELETE FROM
			scd_constraints
		WHERE
			id = $1`
	)

	res, err := c.q.ExecContext(ctx, query, id)
	if err != nil {
		return stacktrace.Propagate(err, "Error in query: %s", query)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get RowsAffected")
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
			WHERE
			  cells && $1
			AND
				COALESCE(starts_at <= $3, true)
			AND
				COALESCE(ends_at >= $2, true)`, constraintFieldsWithoutPrefix)
	)

	// TODO: Lazily calculate & cache spatial covering so that it is only ever
	// computed once on a particular Volume4D
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not calculate spatial covering")
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
		return nil, stacktrace.Propagate(err, "Error fetching Constraints")
	}

	return constraints, nil
}
