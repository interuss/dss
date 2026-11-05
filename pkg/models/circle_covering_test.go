package models

import (
	"math"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// Independent great-circle distance oracle: do not rebuild the production S2 Cap
// and compare it with the output of CalculateCovering.
func circleCenterDistanceMeters(a, b LatLngPoint) float64 {
	const earthRadiusMeters = 6371000.0
	lat1, lat2 := a.Lat*math.Pi/180, b.Lat*math.Pi/180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180
	h := math.Pow(math.Sin(dLat/2), 2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dLng/2), 2)
	return 2 * earthRadiusMeters * math.Asin(math.Sqrt(h))
}

func TestCircleCoveringIncludesBoundaryCells(t *testing.T) {
	primary := GeoCircle{Center: LatLngPoint{Lat: 37.427636, Lng: -122.170502}, RadiusMeter: 15000}
	secondary := GeoCircle{Center: LatLngPoint{Lat: 37.450505047, Lng: -122.003926497}, RadiusMeter: 20}

	distance := circleCenterDistanceMeters(primary.Center, secondary.Center)
	require.Less(t, distance, float64(primary.RadiusMeter+secondary.RadiusMeter),
		"Test fixture must consist of geographically overlapping circles")

	a, err := primary.CalculateCovering()
	require.NoError(t, err)
	b, err := secondary.CalculateCovering()
	require.NoError(t, err)
	require.True(t, slices.IsSorted(a))
	require.True(t, slices.IsSorted(b))
	require.True(t, a.Intersects(b),
		"Geographically overlapping circles must have intersecting S2 coverings")
}
