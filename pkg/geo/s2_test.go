package geo_test

import (
	"testing"

	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/dss/pkg/geo/testdata"

	"github.com/stretchr/testify/require"
)

func TestParseAreaSuccessForOddNumberOfPoints(t *testing.T) {
	cells, err := geo.AreaToCellIDs(`37.4047,-122.1474,37.4037,-122.1485,37.4035,-122.1466`)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaSuccessForOppositeWindingOrder(t *testing.T) {
	cells, err := geo.AreaToCellIDs(`0.000,0.000, 0.000,0.005, -0.005,0.0025`)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaSuccessForEvenNumberOfPoints(t *testing.T) {
	cells, err := geo.AreaToCellIDs(`37.4047,-122.1474,37.4037,-122.1485,37.4035,-122.1466,37.4035,-122.1466`)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaSucceedsForValidLoop(t *testing.T) {
	cells, err := geo.AreaToCellIDs(testdata.Loop)
	require.NoError(t, err)
	require.NotNil(t, cells)
}

func TestParseAreaFailsForEmptyString(t *testing.T) {
	cells, err := geo.AreaToCellIDs("")
	require.Error(t, err)
	require.Nil(t, cells)
}

func TestParseAreaFailsForLoopWithOnlyTwoPoints(t *testing.T) {
	cells, err := geo.AreaToCellIDs(testdata.LoopWithOnlyTwoPoints)
	require.Error(t, err)
	require.Nil(t, cells)
}

func TestParseAreaFailsForLoopWithOddNumberOfCoordinates(t *testing.T) {
	cells, err := geo.AreaToCellIDs(testdata.LoopWithOddNumberOfCoordinates)
	require.Error(t, err)
	require.Nil(t, cells)
}
