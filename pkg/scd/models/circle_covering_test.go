package models

import (
	"testing"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/stretchr/testify/require"
)

func TestSCDCircleCoveringDoesNotMissOverlap(t *testing.T) {
	mk := func(lat, lng, r float64) *restapi.Circle {
		return &restapi.Circle{
			Center: &restapi.LatLngPoint{Lat: restapi.Latitude(lat), Lng: restapi.Longitude(lng)},
			Radius: &restapi.Radius{Value: float32(r)},
		}
	}
	a, err := circleCellsFromSCDRest(mk(37.427636, -122.170502, 15000))
	require.NoError(t, err)
	b, err := circleCellsFromSCDRest(mk(37.450505047, -122.003926497, 20))
	require.NoError(t, err)

	require.True(t, a.Intersects(b),
		"SCD must index overlapping circular areas into intersecting cells")
}
