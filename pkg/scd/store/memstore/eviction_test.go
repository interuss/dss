package memstore

import (
	"context"
	"testing"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestEvictionConfigNotSet(t *testing.T) {
	got, err := setUpStore(t).GetEvictionState(context.Background())
	require.NoError(t, err)
	require.Equal(t, defaultEvictionConfig, got.Config)
}

func TestSetEvictionConfig(t *testing.T) {
	ctx := context.Background()
	r := setUpStore(t)

	config := dssmodels.EvictionConfig{Interval: 24 * time.Hour, Parameters: dssmodels.EvictionParameters{TTL: 2688 * time.Hour}}
	got, err := r.SetEvictionConfig(ctx, config)
	require.NoError(t, err)
	require.Equal(t, config, got)

	replacement := dssmodels.EvictionConfig{Interval: time.Hour, Parameters: dssmodels.EvictionParameters{TTL: time.Hour, Limit: 3}}
	_, err = r.SetEvictionConfig(ctx, replacement)
	require.NoError(t, err)
	stored, err := r.GetEvictionState(ctx)
	require.NoError(t, err)
	require.Equal(t, replacement, stored.Config)
}
