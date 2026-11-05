package models

import (
	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSCDCircleCoveringDoesNotMissOverlap(t *testing.T) {
	mk := func(lat, lng, r float64) *restapi.Circle {
		return &restapi.Circle{Center: &restapi.LatLngPoint{Lat: restapi.Latitude(lat), Lng: restapi.Longitude(lng)}, Radius: &restapi.Radius{Value: float32(r)}}
	}
	a, err := circleCellsFromSCDRest(mk(37.427636, -122.170502, 15000))
	require.NoError(t, err)
	b, err := circleCellsFromSCDRest(mk(37.450505047, -122.003926497, 20))
	require.NoError(t, err)
	set := make(map[s2.CellID]bool, len(a))
	for _, id := range a {
		set[id] = true
	}
	matched := false
	for _, id := range b {
		if set[id] {
			matched = true
		}
	}
	require.True(t, matched, "SCD must index overlapping circular areas into shared cells")
}
