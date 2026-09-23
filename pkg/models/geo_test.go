package models

import (
	"slices"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/geo"
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
	require.True(t, slices.IsSorted(got), "Geometry.CalculateCovering must return a sorted CellUnion")
}

func TestGeoCircleCoveringIsSorted(t *testing.T) {
	got, err := (&GeoCircle{
		Center:      LatLngPoint{Lat: 37.427636, Lng: -122.170502},
		RadiusMeter: 300,
	}).CalculateCovering()

	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.True(t, slices.IsSorted(got), "Geometry.CalculateCovering must return a sorted CellUnion")
}

func TestPrecomputedCellGeometryCoveringIsSorted(t *testing.T) {
	parent := s2.CellIDFromToken("808fb0ac").Parent(geo.DefaultMinimumCellLevel - 1)
	siblings := parent.Children()
	extra := s2.CellIDFromToken("808fb744").Parent(geo.DefaultMinimumCellLevel)

	pcg := precomputedCellGeometry{}.merge(siblings[3], siblings[1], extra, siblings[0], siblings[2])

	got, err := pcg.CalculateCovering()

	require.NoError(t, err)
	require.True(t, slices.IsSorted(got), "Geometry.CalculateCovering must return a sorted CellUnion")

	want := s2.CellUnion{siblings[0], siblings[1], siblings[2], siblings[3], extra}
	slices.Sort(want)
	require.Equal(t, want, got)

	for _, id := range got {
		require.Equal(t, geo.DefaultMinimumCellLevel, id.Level())
	}
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

func TestUnionCellsVolumes4D_Time(t *testing.T) {
	now := time.Now()
	start := now.Add(time.Hour)
	end := start.Add(time.Hour)

	nextStart := end.Add(time.Hour)
	nextEnd := nextStart.Add(time.Hour)

	tests := []struct {
		name          string
		volumes       []*CellsVolume4D
		wantStartTime *time.Time
		wantEndTime   *time.Time
	}{
		{
			name:          "unbounded",
			volumes:       []*CellsVolume4D{{}},
			wantStartTime: nil,
			wantEndTime:   nil,
		},
		{
			name:          "single bounded",
			volumes:       []*CellsVolume4D{{StartTime: &start, EndTime: &end}},
			wantStartTime: &start,
			wantEndTime:   &end,
		},
		{
			name:          "single unbounded start",
			volumes:       []*CellsVolume4D{{EndTime: &end}},
			wantStartTime: nil,
			wantEndTime:   &end,
		},
		{
			name:          "single unbounded end",
			volumes:       []*CellsVolume4D{{StartTime: &start}},
			wantStartTime: &start,
			wantEndTime:   nil,
		},
		{
			name:          "multiple bounded",
			volumes:       []*CellsVolume4D{{StartTime: &start, EndTime: &end}, {StartTime: &nextStart, EndTime: &nextEnd}},
			wantStartTime: &start,
			wantEndTime:   &nextEnd,
		},
		{
			name:          "multiple unbounded",
			volumes:       []*CellsVolume4D{{StartTime: &start, EndTime: &end}, {}},
			wantStartTime: nil,
			wantEndTime:   nil,
		},
		{
			name:          "multiple unbounded combination",
			volumes:       []*CellsVolume4D{{StartTime: &start}, {EndTime: &nextEnd}},
			wantStartTime: nil,
			wantEndTime:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			union := UnionCellsVolumes4D(tt.volumes...)
			assert.Equal(t, tt.wantStartTime, union.StartTime)
			assert.Equal(t, tt.wantEndTime, union.EndTime)
		})
	}
}

func TestUnionCellsVolumes4D_Altitude(t *testing.T) {
	var (
		lower  float32 = 100.0
		low    float32 = 200.0
		high   float32 = 300.0
		higher float32 = 400.0
	)

	tests := []struct {
		name    string
		volumes []*CellsVolume4D
		wantLo  *float32
		wantHi  *float32
	}{
		{
			name:    "unbounded",
			volumes: []*CellsVolume4D{{}},
			wantLo:  nil,
			wantHi:  nil,
		},
		{
			name:    "single bounded",
			volumes: []*CellsVolume4D{{AltitudeLo: &low, AltitudeHi: &high}},
			wantLo:  &low,
			wantHi:  &high,
		},
		{
			name:    "single unbounded low",
			volumes: []*CellsVolume4D{{AltitudeHi: &high}},
			wantLo:  nil,
			wantHi:  &high,
		},
		{
			name:    "single unbounded high",
			volumes: []*CellsVolume4D{{AltitudeLo: &low}},
			wantLo:  &low,
			wantHi:  nil,
		},
		{
			name:    "multiple bounded",
			volumes: []*CellsVolume4D{{AltitudeLo: &lower, AltitudeHi: &low}, {AltitudeLo: &high, AltitudeHi: &higher}},
			wantLo:  &lower,
			wantHi:  &higher,
		},
		{
			name:    "multiple unbounded",
			volumes: []*CellsVolume4D{{AltitudeLo: &lower, AltitudeHi: &low}, {}},
			wantLo:  nil,
			wantHi:  nil,
		},
		{
			name:    "multiple unbounded combination",
			volumes: []*CellsVolume4D{{AltitudeLo: &low}, {AltitudeHi: &high}},
			wantLo:  nil,
			wantHi:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			union := UnionCellsVolumes4D(tt.volumes...)
			assert.Equal(t, tt.wantLo, union.AltitudeLo)
			assert.Equal(t, tt.wantHi, union.AltitudeHi)
		})
	}
}

func TestUnionCellsVolumes4D_Cells(t *testing.T) {
	a := s2.CellIDFromToken("808fb0ac").Parent(geo.DefaultMinimumCellLevel)
	b := s2.CellIDFromToken("808fb744").Parent(geo.DefaultMinimumCellLevel)
	c := s2.CellIDFromToken("808fb754").Parent(geo.DefaultMinimumCellLevel)

	got := UnionCellsVolumes4D(
		&CellsVolume4D{Cells: s2.CellUnion{b, a}},
		&CellsVolume4D{Cells: s2.CellUnion{c, a}},
	)

	want := s2.CellUnion{a, b, c}
	slices.Sort(want)
	require.Equal(t, want, got.Cells)
	require.True(t, slices.IsSorted(got.Cells))
}

func TestUnionCellsVolumes4D_LevelifiesMergedCells(t *testing.T) {
	parent := s2.CellIDFromToken("808fb0ac").Parent(geo.DefaultMinimumCellLevel - 1)
	siblings := parent.Children()

	got := UnionCellsVolumes4D(
		&CellsVolume4D{Cells: s2.CellUnion{siblings[0], siblings[1]}},
		&CellsVolume4D{Cells: s2.CellUnion{siblings[2], siblings[3]}},
	)

	want := s2.CellUnion{siblings[0], siblings[1], siblings[2], siblings[3]}
	slices.Sort(want)
	require.Equal(t, want, got.Cells)

	for _, id := range got.Cells {
		require.Equal(t, geo.DefaultMinimumCellLevel, id.Level())
	}
}
