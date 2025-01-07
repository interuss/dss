package cockroach

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"

	"github.com/jackc/pgx/v5"
)

const (
	nConstraintFields = 10
)

var (
	constraintFieldsWithIndices   [nConstraintFields]string
	constraintFieldsWithPrefix    string
	constraintFieldsWithoutPrefix string
)

// TODO Update database schema and fields below.
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
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", query)
	}
	defer rows.Close()

	var payload []*scdmodels.Constraint
	var cids []int64
	for rows.Next() {
		var (
			c         = new(scdmodels.Constraint)
			updatedAt time.Time
		)
		err := rows.Scan(
			&c.ID,
			&c.Manager,
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
		return nil, pgx.ErrNoRows
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
	uid, err := id.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.fetchConstraint(ctx, c.q, query, uid)
}

// Implements scd.repos.Constraint.UpsertConstraint
func (c *repo) UpsertConstraint(ctx context.Context, s *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		INSERT INTO
		  scd_constraints
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, transaction_timestamp())
		ON CONFLICT (%s) DO UPDATE 
		SET %s = $2,
			%s = $3,
			%s = $4,
			%s = $5,
			%s = $6,
			%s = $7,
			%s = $8,
			%s = $9,
			%s = transaction_timestamp()
		WHERE scd_constraints.%s = $1
		RETURNING
			%s`,
			constraintFieldsWithoutPrefix,
			constraintFieldsWithIndices[0],
			constraintFieldsWithIndices[1],
			constraintFieldsWithIndices[2],
			constraintFieldsWithIndices[3],
			constraintFieldsWithIndices[4],
			constraintFieldsWithIndices[5],
			constraintFieldsWithIndices[6],
			constraintFieldsWithIndices[7],
			constraintFieldsWithIndices[8],
			constraintFieldsWithIndices[9],
			constraintFieldsWithIndices[0],
			constraintFieldsWithPrefix,
		)
	)

	cids, err := dsssql.CellUnionToCellIdsWithValidation(s.Cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")

	}

	id, err := s.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	s, err = c.fetchConstraint(ctx, c.q, upsertQuery,
		id,
		s.Manager,
		s.Version,
		s.USSBaseURL,
		s.AltitudeLower,
		s.AltitudeUpper,
		s.StartTime,
		s.EndTime,
		cids)
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

	uid, err := id.PgUUID()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	res, err := c.q.Exec(ctx, query, uid)
	if err != nil {
		return stacktrace.Propagate(err, "Error in query: %s", query)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
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
				COALESCE(ends_at >= $2, true)
			LIMIT $4;
			`, constraintFieldsWithoutPrefix)
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

	constraints, err := c.fetchConstraints(
		ctx, c.q, query, dsssql.CellUnionToCellIds(cells), v4d.StartTime, v4d.EndTime, dssmodels.MaxResultLimit)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Constraints")
	}

	return constraints, nil
}
