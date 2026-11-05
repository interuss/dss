package models

import (
	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/geo"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
)

func TestCircleCoveringIncludesBoundaryCells(t *testing.T) {
	primary := GeoCircle{Center: LatLngPoint{Lat: 37.427636, Lng: -122.170502}, RadiusMeter: 15000}
	secondary := GeoCircle{Center: LatLngPoint{Lat: 37.450505047, Lng: -122.003926497}, RadiusMeter: 20}
	a, err := primary.CalculateCovering()
	require.NoError(t, err)
	b, err := secondary.CalculateCovering()
	require.NoError(t, err)
	require.True(t, slices.IsSorted(a))
	require.True(t, slices.IsSorted(b))
	matched := false
	for _, id := range a {
		for _, other := range b {
			if id == other {
				matched = true
				break
			}
		}
	}
	require.True(t, matched, "Geographically overlapping circles must share an indexed cell")
	for _, circle := range []GeoCircle{primary, secondary} {
		covered, err := circle.CalculateCovering()
		require.NoError(t, err)
		center := s2.PointFromLatLng(s2.LatLngFromDegrees(circle.Center.Lat, circle.Center.Lng))
		cap := s2.CapFromCenterAngle(center, geo.DistanceMetersToAngle(float64(circle.RadiusMeter)))
		expected := geo.RegionCoverer.Covering(cap)
		require.Equal(t, expected, covered)
	}
}
