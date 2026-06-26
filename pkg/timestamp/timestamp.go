package timestamp

import (
	"context"
	"net/http"
	"time"

	"github.com/interuss/stacktrace"
)

type timestampKey struct{}

// RequestTimestampFromContext returns the request timestamp from the context, or an error if the value is not present or if it is zero.
// The timestamp is set by the Middleware when a query is received then (on the receiver side) by the Raftstore when the query is applied.
// It is then used for deterministic execution of time-dependent queries.
func RequestTimestampFromContext(ctx context.Context) (time.Time, error) {
	timestamp, ok := ctx.Value(timestampKey{}).(time.Time)
	if !ok {
		return time.Time{}, stacktrace.NewError("timestamp not found in context")
	}

	if timestamp.IsZero() {
		return time.Time{}, stacktrace.NewError("timestamp is zero")
	}

	return timestamp, nil
}

// WithRequestTimestamp returns a new context with the given timestamp.
func WithRequestTimestamp(ctx context.Context, timestamp time.Time) context.Context {
	return context.WithValue(ctx, timestampKey{}, timestamp)
}

// RequestTimestampMiddleware is an HTTP middleware that stamps each incoming
// request with its received time. This timestamp is later used as the
// timestamp of the Raft proposal, so that time-dependent queries
// execute deterministically across nodes and contexts (catchup / restart etc.).
func RequestTimestampMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithRequestTimestamp(r.Context(), time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
