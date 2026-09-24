package apiv1

import (
	"testing"
	"time"

	restapi "github.com/interuss/dss/pkg/api/ridv1"
	"github.com/stretchr/testify/require"
)

func triangleVolume4D() *restapi.Volume4D {
	altLo := restapi.Altitude(100)
	altHi := restapi.Altitude(200)
	timeStart := "2023-01-01T00:00:00Z"
	timeEnd := "2023-01-01T01:00:00Z"
	return &restapi.Volume4D{
		SpatialVolume: restapi.Volume3D{
			Footprint: restapi.GeoPolygon{
				Vertices: []restapi.LatLngPoint{
					{Lat: 37.427636, Lng: -122.170502},
					{Lat: 37.408799, Lng: -122.064069},
					{Lat: 37.421265, Lng: -122.086504},
				},
			},
			AltitudeLo: &altLo,
			AltitudeHi: &altHi,
		},
		TimeStart: &timeStart,
		TimeEnd:   &timeEnd,
	}
}

func TestCellsVolume4DFromRest(t *testing.T) {
	vol4 := triangleVolume4D()

	got, err := CellsVolume4DFromRest(vol4)
	require.NoError(t, err)
	require.NotNil(t, got)

	wantCells, err := FromGeoPolygon(&vol4.SpatialVolume.Footprint).CalculateCovering()
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

func TestCellsVolume4DFromRest_InvalidTime(t *testing.T) {
	vol4 := triangleVolume4D()
	badTime := "not-a-time"
	vol4.TimeStart = &badTime

	_, err := CellsVolume4DFromRest(vol4)
	require.Error(t, err)
}

func TestCellsVolume4DFromRest_InvalidFootprint(t *testing.T) {
	vol4 := triangleVolume4D()
	vol4.SpatialVolume.Footprint.Vertices = vol4.SpatialVolume.Footprint.Vertices[:2]

	_, err := CellsVolume4DFromRest(vol4)
	require.Error(t, err)
}
