package apiv2

import (
	"testing"
	"time"

	restapi "github.com/interuss/dss/pkg/api/ridv2"
	"github.com/stretchr/testify/require"
)

func triangleVolume4D() *restapi.Volume4D {
	return &restapi.Volume4D{
		Volume: restapi.Volume3D{
			OutlinePolygon: &restapi.Polygon{
				Vertices: []restapi.LatLngPoint{
					{Lat: 37.427636, Lng: -122.170502},
					{Lat: 37.408799, Lng: -122.064069},
					{Lat: 37.421265, Lng: -122.086504},
				},
			},
			AltitudeLower: &restapi.Altitude{Value: 100, Reference: "W84", Units: "M"},
			AltitudeUpper: &restapi.Altitude{Value: 200, Reference: "W84", Units: "M"},
		},
		TimeStart: &restapi.Time{Value: "2023-01-01T00:00:00Z", Format: "RFC3339"},
		TimeEnd:   &restapi.Time{Value: "2023-01-01T01:00:00Z", Format: "RFC3339"},
	}
}

func circleVolume4D() *restapi.Volume4D {
	return &restapi.Volume4D{
		Volume: restapi.Volume3D{
			OutlineCircle: &restapi.Circle{
				Center: &restapi.LatLngPoint{Lat: 37.427636, Lng: -122.170502},
				Radius: &restapi.Radius{Value: 300, Units: "M"},
			},
		},
	}
}

func TestCellsVolume4DFromRest_Polygon(t *testing.T) {
	vol4 := triangleVolume4D()

	got, err := CellsVolume4DFromRest(vol4)
	require.NoError(t, err)
	require.NotNil(t, got)

	wantCells, err := FromPolygon(vol4.Volume.OutlinePolygon).CalculateCovering()
	require.NoError(t, err)
	require.Equal(t, wantCells, got.Cells)

	require.NotNil(t, got.StartTime)
	require.Equal(t, "2023-01-01T00:00:00Z", got.StartTime.Format(time.RFC3339))
	require.NotNil(t, got.EndTime)
	require.Equal(t, "2023-01-01T01:00:00Z", got.EndTime.Format(time.RFC3339))

	require.NotNil(t, got.AltitudeLo)
	require.InDelta(t, float32(100), *got.AltitudeLo, 0)
	require.NotNil(t, got.AltitudeHi)
	require.InDelta(t, float32(200), *got.AltitudeHi, 0)
}

func TestCellsVolume4DFromRest_Circle(t *testing.T) {
	vol4 := circleVolume4D()

	got, err := CellsVolume4DFromRest(vol4)
	require.NoError(t, err)
	require.NotNil(t, got)

	circle, err := FromCircle(vol4.Volume.OutlineCircle)
	require.NoError(t, err)
	wantCells, err := circle.CalculateCovering()
	require.NoError(t, err)
	require.Equal(t, wantCells, got.Cells)
}

func TestCellsVolume4DFromRest_BothOutlinesSpecified(t *testing.T) {
	vol4 := triangleVolume4D()
	vol4.Volume.OutlineCircle = circleVolume4D().Volume.OutlineCircle

	_, err := CellsVolume4DFromRest(vol4)
	require.Error(t, err)
}

func TestCellsVolume4DFromRest_NoOutlineSpecified(t *testing.T) {
	vol4 := &restapi.Volume4D{}

	_, err := CellsVolume4DFromRest(vol4)
	require.Error(t, err)
}
