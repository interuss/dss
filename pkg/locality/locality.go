package locality

import (
	"context"
	"net/http"

	"github.com/interuss/stacktrace"
)

type key struct{}

// MustFromContext returns the request locality from the context and panics if it is not
// present, which is a programming error.
func MustFromContext(ctx context.Context) string {
	locality, ok := ctx.Value(key{}).(string)
	if !ok {
		panic(stacktrace.NewError("request locality not present in context"))
	}

	return locality
}

// NewContext returns a new context with the given locality.
func NewContext(ctx context.Context, locality string) context.Context {
	return context.WithValue(ctx, key{}, locality)
}

// Middleware is an HTTP middleware that stamps each incoming request with this
// DSS instance's locality so that locality-dependent operations execute deterministically across nodes.
func Middleware(locality string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), locality)))
		})
	}
}
