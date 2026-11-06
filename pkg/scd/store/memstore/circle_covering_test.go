package memstore

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCircleOverlappingIntentIsDiscoverable(t *testing.T) {
	ctx := writeCtx()
	repo := setUpStore(t)
	primary := dssmodels.GeoCircle{Center: dssmodels.LatLngPoint{Lat: 37.427636, Lng: -122.170502}, RadiusMeter: 15000}
	secondary := dssmodels.GeoCircle{Center: dssmodels.LatLngPoint{Lat: 37.450505047, Lng: -122.003926497}, RadiusMeter: 20}
	a, err := primary.CalculateCovering()
	require.NoError(t, err)
	b, err := secondary.CalculateCovering()
	require.NoError(t, err)
	op := sampleOperationalIntent()
	op.Cells = a
	_, err = repo.UpsertOperationalIntent(ctx, op)
	require.NoError(t, err)
	found, err := repo.SearchOperationalIntents(ctx, cellsVolume4D(b, &startTime, &endTime, &altLow, &altHigh))
	require.NoError(t, err)
	require.Len(t, found, 1, "An overlapping operational intent must be discoverable")
	require.Equal(t, op.ID, found[0].ID)
}
