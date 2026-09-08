package timestamp

import (
	"context"
	"net/http"
	"time"

	"github.com/interuss/stacktrace"
)

type key struct{}

// fromContext returns the request timestamp from the context, or an error if the value is not present or if it is zero.
// The timestamp is set by the Middleware when a query is received then (on the receiver side) by the Raftstore when the query is applied.
// It is then used for deterministic execution of time-dependent queries.
func fromContext(ctx context.Context) (time.Time, error) {
	timestamp, ok := ctx.Value(key{}).(time.Time)
	if !ok {
		return time.Time{}, stacktrace.NewError("timestamp not found in context")
	}

	if timestamp.IsZero() {
		return time.Time{}, stacktrace.NewError("timestamp is zero")
	}

	return timestamp, nil
}

// MustFromContext returns the request timestamp from the context and panics if it is not
// present or invalid, which is a programming error.
func MustFromContext(ctx context.Context) time.Time {
	timestamp, err := fromContext(ctx)
	if err != nil {
		panic(err)
	}

	return timestamp
}

// NewContext returns a new context with the given timestamp.
func NewContext(ctx context.Context, timestamp time.Time) context.Context {
	return context.WithValue(ctx, key{}, timestamp)
}

// Middleware is an HTTP middleware that stamps each incoming
// request with its received time. This timestamp is later used as the
// timestamp of the Raft proposal, so that time-dependent queries
// execute deterministically across nodes and contexts (catchup / restart etc.).
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r.Context(), time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
