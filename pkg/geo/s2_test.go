package geo

import (
	"testing"

	"github.com/golang/geo/s2"

	"github.com/stretchr/testify/require"
)

func TestParseAreaSucceedsForValidLoop(t *testing.T) {
	// square shape polygon
	// d-c
	// | |
	// a-b
	pts := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.000)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.005, 0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.005, 0.000)),
	}
	cells, err := Covering(pts)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestGeoPolygonFromRestSuccessForOppositeWindingOrder(t *testing.T) {
	// square shape polygon
	// b-c
	// | |
	// a-d
	pts := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.000)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.005, 0.000)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.005, 0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.005)),
	}
	cells, err := Covering(pts)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestCoveringFailsForIntersectingLoop(t *testing.T) {
	// hourglass shape polygon
	// c-b
	//  X
	// a-d
	pts := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.000)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.005, 0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.005, 0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.005, 0.000)),
	}
	cells, err := Covering(pts)
	require.Error(t, err)
	require.Nil(t, cells)
}

func TestCoveringFailsForSharedVertexLoop(t *testing.T) {
	// L shape polygon
	// a-b+d
	//    |
	//    c
	pts := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.000)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(-0.005, -0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.005)),
	}
	cells, err := Covering(pts)
	require.Error(t, err)
	require.Nil(t, cells)
}

func TestCoveringSucceedsForColinearLoop(t *testing.T) {
	// s2 implements a consistent perturbation model such
	// that no three points are ever considered to be collinear
	// line shape polygon
	// a-b-c-d
	pts := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.000)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.005)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.010)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(0.000, 0.015)),
	}
	cells, err := Covering(pts)
	require.NoError(t, err)
	require.NotNil(t, cells)
}
