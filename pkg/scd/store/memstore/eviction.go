package memstore

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/timestamp"
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

func (r *repo) SetEvictionDisabled(_ context.Context, disabled bool) error {
	r.state.EvictionDisabled = disabled
	return nil
}
