package cockroach

import (
	"context"
	"fmt"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
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
		withPrefix[idx] = "scd_uss_availability." + field
	}

	availabilityFieldsWithPrefix = strings.Join(
		withPrefix[:], ",",
	)
}

// Implements repos.UssAvailability.UpsertAvailability
func (u *repo) UpsertUssAvailability(ctx context.Context, s *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		INSERT INTO
		scd_uss_availability
		  (%s)
		VALUES
			($1, $2, transaction_timestamp())
		ON CONFLICT (id) DO UPDATE
			SET availability = $2, updated_at = transaction_timestamp()
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
	rows, err := q.Query(ctx, query, args...)
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
			return nil, stacktrace.Propagate(err, "Error scanning UssAvailability row")
		}
		u.Version = scdmodels.NewOVNFromTime(updatedAt, u.Uss.String())
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
		return nil, pgx.ErrNoRows
	}
	return availabilities[0], nil
}

// GetUssAvailability returns the Availability status identified by "ussID".
func (u *repo) GetUssAvailability(ctx context.Context, ussID dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	var ussAvailabilityQuery = fmt.Sprintf(`
      SELECT %s
      FROM
        scd_uss_availability
      WHERE
        id = $1`, availabilityFieldsWithoutPrefix)

	ussa, err := u.fetchAvailability(ctx, u.q, ussAvailabilityQuery, ussID)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	} else if ussa == nil {
		return nil, nil
	}
	return ussa, nil
}
