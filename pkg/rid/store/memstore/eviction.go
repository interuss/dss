package memstore

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
)

// defaultEvictionConfig is used while the eviction settings are not set.
var defaultEvictionConfig = dssmodels.EvictionConfig{
	Interval:   time.Hour,
	Parameters: dssmodels.EvictionParameters{TTL: 30 * time.Minute},
}

func newEviction(r *repo) *memstore.Eviction {
	return memstore.NewEviction(map[string]memstore.DeleteExpired{
		"isas":          r.DeleteExpiredISAs,
		"subscriptions": r.DeleteExpiredSubscriptions,
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
