package models

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolygonCovering(t *testing.T) {
	got, err := (&GeoPolygon{
		Vertices: []*LatLngPoint{
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
		},
	}).CalculateCovering()

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

func TestUnionVolumes4D_Time(t *testing.T) {
	now := time.Now()
	start := now.Add(time.Hour)
	end := start.Add(time.Hour)

	nextStart := end.Add(time.Hour)
	nextEnd := nextStart.Add(time.Hour)

	tests := []struct {
		name          string
		volumes       []*Volume4D
		wantStartTime *time.Time
		wantEndTime   *time.Time
	}{
		{
			name:          "unbounded",
			volumes:       []*Volume4D{{}},
			wantStartTime: nil,
			wantEndTime:   nil,
		},
		{
			name:          "single bounded",
			volumes:       []*Volume4D{{StartTime: &start, EndTime: &end}},
			wantStartTime: &start,
			wantEndTime:   &end,
		},
		{
			name:          "single unbounded start",
			volumes:       []*Volume4D{{EndTime: &end}},
			wantStartTime: nil,
			wantEndTime:   &end,
		},
		{
			name:          "single unbounded end",
			volumes:       []*Volume4D{{StartTime: &start}},
			wantStartTime: &start,
			wantEndTime:   nil,
		},
		{
			name:          "multiple bounded",
			volumes:       []*Volume4D{{StartTime: &start, EndTime: &end}, {StartTime: &nextStart, EndTime: &nextEnd}},
			wantStartTime: &start,
			wantEndTime:   &nextEnd,
		},
		{
			name:          "multiple unbounded",
			volumes:       []*Volume4D{{StartTime: &start, EndTime: &end}, {}},
			wantStartTime: nil,
			wantEndTime:   nil,
		},
		{
			name:          "multiple unbounded combination",
			volumes:       []*Volume4D{{StartTime: &start}, {EndTime: &nextEnd}},
			wantStartTime: nil,
			wantEndTime:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			union, err := UnionVolumes4D(tt.volumes...)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStartTime, union.StartTime)
			assert.Equal(t, tt.wantEndTime, union.EndTime)
		})
	}
}

func TestUnionVolumes4D_Altitude(t *testing.T) {
	var (
		lower  float32 = 100.0
		low    float32 = 200.0
		high   float32 = 300.0
		higher float32 = 400.0
	)

	tests := []struct {
		name    string
		volumes []*Volume4D
		wantLo  *float32
		wantHi  *float32
	}{
		{
			name:    "unbounded",
			volumes: []*Volume4D{{SpatialVolume: &Volume3D{}}},
			wantLo:  nil,
			wantHi:  nil,
		},
		{
			name:    "single bounded",
			volumes: []*Volume4D{{SpatialVolume: &Volume3D{AltitudeLo: &low, AltitudeHi: &high}}},
			wantLo:  &low,
			wantHi:  &high,
		},
		{
			name:    "single unbounded low",
			volumes: []*Volume4D{{SpatialVolume: &Volume3D{AltitudeHi: &high}}},
			wantLo:  nil,
			wantHi:  &high,
		},
		{
			name:    "single unbounded high",
			volumes: []*Volume4D{{SpatialVolume: &Volume3D{AltitudeLo: &low}}},
			wantLo:  &low,
			wantHi:  nil,
		},
		{
			name:    "single bounded",
			volumes: []*Volume4D{{SpatialVolume: &Volume3D{AltitudeLo: &lower, AltitudeHi: &low}}, {SpatialVolume: &Volume3D{AltitudeLo: &high, AltitudeHi: &higher}}},
			wantLo:  &lower,
			wantHi:  &higher,
		},
		{
			name:    "multiple unbounded",
			volumes: []*Volume4D{{SpatialVolume: &Volume3D{AltitudeLo: &lower, AltitudeHi: &low}}, {SpatialVolume: &Volume3D{}}},
			wantLo:  nil,
			wantHi:  nil,
		},
		{
			name:    "multiple unbounded combination",
			volumes: []*Volume4D{{SpatialVolume: &Volume3D{AltitudeLo: &low}}, {SpatialVolume: &Volume3D{AltitudeHi: &high}}},
			wantLo:  nil,
			wantHi:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			union, err := UnionVolumes4D(tt.volumes...)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLo, union.SpatialVolume.AltitudeLo)
			assert.Equal(t, tt.wantHi, union.SpatialVolume.AltitudeHi)
		})
	}
}
