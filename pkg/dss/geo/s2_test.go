package geo

import (
	"testing"

	"github.com/interuss/dss/pkg/dss/geo/testdata"
	dspb "github.com/interuss/dss/pkg/dssproto"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/require"
)

func TestGeoPolygonToCellIDs(t *testing.T) {
	got, err := GeoPolygonToCellIDs(&dspb.GeoPolygon{Vertices: []*dspb.LatLngPoint{
		// Stanford
		{
			Lat: 37.427636,
			Lng: -122.170502,
		},
		// NASA Ames
		{
			Lat: 37.408799,
			Lng: -122.064069,
		},
		// Googleplex
		{
			Lat: 37.421265,
			Lng: -122.086504,
		},
	}})

	want := s2.CellUnion{
		s2.CellIDFromToken("808fb0ac"),
		s2.CellIDFromToken("808fb744"),
		s2.CellIDFromToken("808fb754"),
		s2.CellIDFromToken("808fb75c"),
		s2.CellIDFromToken("808fb9fc"),
		s2.CellIDFromToken("808fba04"),
		s2.CellIDFromToken("808fba0c"),
		s2.CellIDFromToken("808fba14"),
		s2.CellIDFromToken("808fba1c"),
		s2.CellIDFromToken("808fba5c"),
		s2.CellIDFromToken("808fba64"),
		s2.CellIDFromToken("808fba6c"),
		s2.CellIDFromToken("808fba74"),
		s2.CellIDFromToken("808fba8c"),
		s2.CellIDFromToken("808fbad4"),
		s2.CellIDFromToken("808fbadc"),
		s2.CellIDFromToken("808fbae4"),
		s2.CellIDFromToken("808fbaec"),
		s2.CellIDFromToken("808fbaf4"),
		s2.CellIDFromToken("808fbb2c"),
	}
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestParseAreaSuccessForOddNumberOfPoints(t *testing.T) {
	cells, err := AreaToCellIDs(`37.4047,-122.1474,37.4037,-122.1485,37.4035,-122.1466`)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaSuccessForOppositeWindingOrder(t *testing.T) {
	cells, err := AreaToCellIDs(`0.000,0.000, 0.000,0.005, -0.005,0.0025`)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaSuccessForEvenNumberOfPoints(t *testing.T) {
	cells, err := AreaToCellIDs(`37.4047,-122.1474,37.4037,-122.1485,37.4035,-122.1466,37.4035,-122.1466`)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaSucceedsForValidLoop(t *testing.T) {
	cells, err := AreaToCellIDs(testdata.Loop)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaFailsForEmptyString(t *testing.T) {
	cells, err := AreaToCellIDs("")
	require.Error(t, err)
	require.Nil(t, cells)
}

func TestParseAreaFailsForLoopWithOnlyTwoPoints(t *testing.T) {
	cells, err := AreaToCellIDs(testdata.LoopWithOnlyTwoPoints)
	require.Error(t, err)
	require.Nil(t, cells)
}

func TestParseAreaFailsForLoopWithOddNumberOfCoordinates(t *testing.T) {
	cells, err := AreaToCellIDs(testdata.LoopWithOddNumberOfCoordinates)
	require.Error(t, err)
	require.Nil(t, cells)
}
