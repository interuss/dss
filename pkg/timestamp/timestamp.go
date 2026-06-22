package timestamp

import (
	"context"
	"net/http"
	"time"
)

type timestampKey struct{}

// NowFromContext returns the timestamp from the context, or zero if not present.
func NowFromContext(ctx context.Context) time.Time {
	t, ok := ctx.Value(timestampKey{}).(time.Time)
	if !ok {
		return time.Time{}
	}

	return t
}

// WithTimestamp returns a new context with the given timestamp.
func WithTimestamp(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, timestampKey{}, t)
}

// Middleware is an HTTP middleware that adds a timestamp to the request context.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithTimestamp(r.Context(), time.Now().UTC())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
