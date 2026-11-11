package memstore

import (
	"context"
	"maps"
	"slices"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

// DeleteExpired deletes up to limit expired entities of an entity type, 0 meaning unlimited, and returns their IDs.
type DeleteExpired func(ctx context.Context, threshold time.Time, limit int) ([]dssmodels.ID, error)

// Eviction deletes the expired entities of the entity types of a memstore, registered by name.
type Eviction struct {
	entityTypes map[string]DeleteExpired
}

func NewEviction(entityTypes map[string]DeleteExpired) *Eviction {
	return &Eviction{entityTypes: entityTypes}
}

// DeleteExpired deletes the expired entities of every entity type according to p and returns their number per entity
// type. The entity types are processed in sorted order and the threshold computed from the proposal timestamp of ctx,
// to ensure deterministic behavior.
func (e *Eviction) DeleteExpired(ctx context.Context, p dssmodels.EvictionParameters) (map[string]int, error) {
	threshold := timestamp.MustFromContext(ctx).Add(-p.TTL)
	counts := make(map[string]int, len(e.entityTypes))
	for _, name := range slices.Sorted(maps.Keys(e.entityTypes)) {
		ids, err := e.entityTypes[name](ctx, threshold, p.Limit)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to delete the expired %s", name)
		}
		counts[name] = len(ids)
	}
	return counts, nil
}
