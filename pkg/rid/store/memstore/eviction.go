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

func (r *repo) GetEvictionState(_ context.Context) (dssmodels.EvictionState, error) {
	state := dssmodels.EvictionState{Config: defaultEvictionConfig, LastRun: r.state.LastEviction, Disabled: r.state.EvictionDisabled}
	if r.state.EvictionConfig != nil {
		state.Config = *r.state.EvictionConfig
	}
	return state, nil
}

func (r *repo) SetEvictionConfig(_ context.Context, config dssmodels.EvictionConfig) (dssmodels.EvictionConfig, error) {
	r.state.EvictionConfig = &config
	return config, nil
}

func (r *repo) SetEvictionDisabled(ctx context.Context, disabled bool) (dssmodels.EvictionState, error) {
	r.state.EvictionDisabled = disabled
	return r.GetEvictionState(ctx)
}
