package memstore

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/timestamp"
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
	counts, err := r.eviction.DeleteExpired(ctx, params)
	if err != nil {
		return nil, err
	}
	r.state.LastEviction = timestamp.MustFromContext(ctx)
	return counts, nil
}

func (r *repo) GetEvictionConfig(_ context.Context) (dssmodels.EvictionConfig, error) {
	config := defaultEvictionConfig
	if r.state.EvictionConfig != nil {
		config = *r.state.EvictionConfig
	}
	config.LastRun = r.state.LastEviction
	return config, nil
}

func (r *repo) SetEvictionConfig(ctx context.Context, config dssmodels.EvictionConfig) (dssmodels.EvictionConfig, error) {
	r.state.EvictionConfig = &config
	return r.GetEvictionConfig(ctx)
}
