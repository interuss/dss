package sql

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Queryable abstracts common operations on sql.DB and sql.Tx instances.
type Queryable interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
}

// FetchIDs runs a query returning a single ID column and returns the IDs of all rows.
func FetchIDs(ctx context.Context, q Queryable, query string, args ...interface{}) ([]dssmodels.ID, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in query: %s", query)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[dssmodels.ID])
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning IDs")
	}
	return ids, nil
}

// IDStrings converts IDs to strings.
func IDStrings(ids []dssmodels.ID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}
