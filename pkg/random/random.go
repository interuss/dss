// Package random provides a per-request seed.
//
// Used by business logic operations that need to generate pseudo-random data (e.g. UUIDs)
// while still executing deterministically (every Raft node / replay must derive the same value).
package random

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"hash/fnv"
	mrand "math/rand"
	"net/http"

	"github.com/interuss/stacktrace"
)

type key struct{}

func newSeed() (int64, error) {
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to generate random seed")
	}
	return int64(binary.BigEndian.Uint64(buf[:])), nil
}

// MustFromContext returns the request seed from the context and panics if it is not present,
// which is a programming error.
func MustFromContext(ctx context.Context) int64 {
	seed, ok := ctx.Value(key{}).(int64)
	if !ok {
		panic(stacktrace.NewError("seed not found in context"))
	}
	return seed
}

// NewContext returns a new context with the given seed.
func NewContext(ctx context.Context, seed int64) context.Context {
	return context.WithValue(ctx, key{}, seed)
}

// Generator deterministically builds a pseudo-random generator from the given seed and label.
// The same (seed, label) pair always yields a generator producing the same sequence of values.
// Distinct labels let a single request derive multiple, independent values.
func Generator(seed int64, label string) (*mrand.Rand, error) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(seed))

	h := fnv.New64a()
	_, err := h.Write(buf[:])
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash seed")
	}
	_, err = h.Write([]byte(label))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash label")
	}

	return mrand.New(mrand.NewSource(int64(h.Sum64()))), nil
}

// Middleware is an HTTP middleware that stamps each incoming request with a fresh seed, so that
// any pseudo-random data (e.g. generated UUIDs) needed while executing it can be derived
// deterministically via Generator.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seed, err := newSeed()
		if err != nil {
			http.Error(w, "failed to generate request seed", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), seed)))
	})
}
