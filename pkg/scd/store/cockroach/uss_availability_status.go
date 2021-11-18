package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"strings"
	"time"
)

var (
	availabilityFieldsWithIndices   [3]string
	availabilityFieldsWithPrefix    string
	availabilityFieldsWithoutPrefix string
)

// TODO Update database schema and fields below.
func init() {
	availabilityFieldsWithIndices[0] = "id"
	availabilityFieldsWithIndices[1] = "availability"
	availabilityFieldsWithIndices[2] = "updated_at"

	availabilityFieldsWithoutPrefix = strings.Join(
		availabilityFieldsWithIndices[:], ",",
	)

	withPrefix := make([]string, len(availabilityFieldsWithIndices))
	for idx, field := range availabilityFieldsWithIndices {
		withPrefix[idx] = "scd_availabilities." + field
	}

	availabilityFieldsWithPrefix = strings.Join(
		withPrefix[:], ",",
	)
}

// GetussAvailability implements repos.Availability.GetUssAvailabilityStatus.
func (u *repo) GetUssAvailabilityStatus(ctx context.Context, ussID dssmodels.ID) (*scdmodels.UssAvailabilityStatus, error) {
	var ussAvailabilityQuery = `
      SELECT
        availability
      FROM
        uss_availability
      WHERE
        id = $1`

	return u.fetchAvailability(ctx, u.q, ussAvailabilityQuery, ussID)
}

// Implements repos.Availability.UpsertAvailability
func (u *repo) UpsertUssAvailability(ctx context.Context, s *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		UPSERT INTO
		uss_availability
		  (%s)
		VALUES
			($1, $2, transaction_timestamp())
		RETURNING
			%s`, availabilityFieldsWithoutPrefix, availabilityFieldsWithPrefix)
	)

	s, err := u.fetchAvailability(ctx, u.q, upsertQuery,
		s.Uss,
		s.Availability)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching Availability")
	}
	return s, nil
}

func (u *repo) fetchAvailabilities(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) ([]*scdmodels.UssAvailabilityStatus, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", query)
	}
	defer rows.Close()

	var payload []*scdmodels.UssAvailabilityStatus
	for rows.Next() {
		var (
			u         = new(scdmodels.UssAvailabilityStatus)
			updatedAt time.Time
		)
		err := rows.Scan(
			&u.Uss,
			&u.Availability,
			&updatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Constraint row")
		}
		payload = append(payload, u)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}
	return payload, nil
}

func (u *repo) fetchAvailability(ctx context.Context, q dsssql.Queryable, query string, args ...interface{}) (*scdmodels.UssAvailabilityStatus, error) {
	availabilities, err := u.fetchAvailabilities(ctx, q, query, args...)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}
	if len(availabilities) > 1 {
		return nil, stacktrace.NewError("Query returned %d availabilities when only 0 or 1 was expected", len(availabilities))
	}
	if len(availabilities) == 0 {
		return nil, sql.ErrNoRows
	}
	return availabilities[0], nil
}

// GetUssAvailability returns the Availability status identified by "id".
func (c *repo) GetUssAvailability(ctx context.Context, id dssmodels.ID) (*scdmodels.UssAvailabilityStatus, error) {
	sub, err := c.GetUssAvailabilityStatus(ctx, id)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	} else if sub == nil {
		return nil, nil
	}
	return sub, nil
}
