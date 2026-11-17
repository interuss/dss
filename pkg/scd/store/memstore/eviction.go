package memstore

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
)

// defaultEvictionConfig is used while the eviction settings are not set.
// The TTL of 2*56 days is the loosest that can be argued for, F3548-21 needing flight information at most 56 days in advance,
// and is not lower as USSs may treat expired operational intents still present as hazards
// (https://github.com/interuss/dss/issues/1461#issuecomment-4426817959).
var defaultEvictionConfig = dssmodels.EvictionConfig{
	Interval:   24 * time.Hour,
	Parameters: dssmodels.EvictionParameters{TTL: 2688 * time.Hour},
}

func newEviction(r *repo) *memstore.Eviction {
	return memstore.NewEviction(map[string]memstore.DeleteExpired{
		"operational_intents": r.DeleteExpiredOperationalIntents,
		"subscriptions":       r.DeleteExpiredSubscriptions,
	})
}

func (r *repo) DeleteExpired(ctx context.Context, params dssmodels.EvictionParameters) (map[string]int, error) {
	return r.eviction.DeleteExpired(ctx, params)
}

func (r *repo) GetEvictionConfig(_ context.Context) (dssmodels.EvictionConfig, error) {
	if r.state.EvictionConfig == nil {
		return defaultEvictionConfig, nil
	}
	return *r.state.EvictionConfig, nil
}

func (r *repo) SetEvictionConfig(_ context.Context, config dssmodels.EvictionConfig) (dssmodels.EvictionConfig, error) {
	r.state.EvictionConfig = &config
	return config, nil
}
