package locality

import (
	"context"

	"github.com/interuss/stacktrace"
)

type localityKey struct{}

// LocalityFromContext returns the locality from the context, or an error if not present.
func LocalityFromContext(ctx context.Context) (string, error) {
	locality, ok := ctx.Value(localityKey{}).(string)
	if !ok {
		return "", stacktrace.NewError("locality not found in context")
	}
	return locality, nil
}

// WithLocality returns a new context with the given locality.
func WithLocality(ctx context.Context, locality string) context.Context {
	return context.WithValue(ctx, localityKey{}, locality)
}
