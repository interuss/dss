package memstore

import (
	"context"
	"testing"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestEvictionConfigNotSet(t *testing.T) {
	got, err := setUpStore(t).GetEvictionConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, defaultEvictionConfig, got)
}

func TestSetEvictionConfig(t *testing.T) {
	ctx := context.Background()
	r := setUpStore(t)

	config := dssmodels.EvictionConfig{Interval: time.Hour, Parameters: dssmodels.EvictionParameters{TTL: 30 * time.Minute}}
	got, err := r.SetEvictionConfig(ctx, config)
	require.NoError(t, err)
	require.Equal(t, config, got)

	replacement := dssmodels.EvictionConfig{Interval: time.Minute, Parameters: dssmodels.EvictionParameters{TTL: time.Hour, Limit: 3}}
	_, err = r.SetEvictionConfig(ctx, replacement)
	require.NoError(t, err)
	stored, err := r.GetEvictionConfig(ctx)
	require.NoError(t, err)
	require.Equal(t, replacement, stored)
}
