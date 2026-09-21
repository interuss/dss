package models

import (
	"testing"
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRestTime(t time.Time) *restapi.Time {
	return &restapi.Time{Value: t.Format(time.RFC3339), Format: TimeFormatRFC3339}
}

func newRestAlt(a float32) *restapi.Altitude {
	return &restapi.Altitude{Value: float64(a), Reference: ReferenceW84, Units: UnitsM}
}

var footprint = restapi.Polygon{
	Vertices: []restapi.LatLngPoint{
		{Lat: 37.427636, Lng: -122.170502},
		{Lat: 37.408799, Lng: -122.064069},
		{Lat: 37.421265, Lng: -122.086504},
	},
}

func TestCellsVolume4DFromSCDRest(t *testing.T) {
	start := time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	restInvalidTime := &restapi.Time{Value: start.Format(time.ANSIC)}
	lo := float32(100.0)
	hi := float32(200.0)
	restInvalidAlt := &restapi.Altitude{Value: 0}

	testCases := []struct {
		name      string
		rest      *restapi.Volume4D
		wantStart *time.Time
		wantEnd   *time.Time
		wantAltLo *float32
		wantAltHi *float32
		wantCells bool
		wantErr   bool
	}{
		{
			name:    "Empty",
			rest:    &restapi.Volume4D{},
			wantErr: true,
		},
		{
			name:      "Times",
			rest:      &restapi.Volume4D{TimeStart: newRestTime(start), TimeEnd: newRestTime(end)},
			wantStart: &start,
			wantEnd:   &end,
			wantErr:   true,
		},
		{
			name:    "InvalidTimeStart",
			rest:    &restapi.Volume4D{TimeStart: restInvalidTime},
			wantErr: true,
		},
		{
			name:    "InvalidTimeEnd",
			rest:    &restapi.Volume4D{TimeEnd: restInvalidTime},
			wantErr: true,
		},
		{
			name:    "TimeStartAfterTimeEnd",
			rest:    &restapi.Volume4D{TimeStart: newRestTime(end), TimeEnd: newRestTime(start)},
			wantErr: true,
		},
		{
			name:      "Polygon",
			rest:      &restapi.Volume4D{Volume: restapi.Volume3D{OutlinePolygon: &footprint}},
			wantCells: true,
		},
		{
			name: "Circle",
			rest: &restapi.Volume4D{Volume: restapi.Volume3D{
				OutlineCircle: &restapi.Circle{
					Center: &restapi.LatLngPoint{Lat: 37.427636, Lng: -122.170502},
					Radius: &restapi.Radius{Value: 300},
				},
			}},
			wantCells: true,
		},
		{
			name:      "Altitudes",
			rest:      &restapi.Volume4D{Volume: restapi.Volume3D{OutlinePolygon: &footprint, AltitudeLower: newRestAlt(lo), AltitudeUpper: newRestAlt(hi)}},
			wantAltLo: &lo,
			wantAltHi: &hi,
			wantCells: true,
		},
		{
			name:    "InvalidLowerAltitude",
			rest:    &restapi.Volume4D{Volume: restapi.Volume3D{AltitudeLower: restInvalidAlt}},
			wantErr: true,
		},
		{
			name:    "InvalidUpperAltitude",
			rest:    &restapi.Volume4D{Volume: restapi.Volume3D{AltitudeUpper: restInvalidAlt}},
			wantErr: true,
		},
		{
			name:    "LowerAltitudeGreaterThanUpperAltitude",
			rest:    &restapi.Volume4D{Volume: restapi.Volume3D{AltitudeLower: newRestAlt(hi), AltitudeUpper: newRestAlt(lo)}},
			wantErr: true,
		},
		{
			name:    "MultiGeom",
			rest:    &restapi.Volume4D{Volume: restapi.Volume3D{OutlineCircle: &restapi.Circle{}, OutlinePolygon: &restapi.Polygon{}}},
			wantErr: true,
		},
		{
			name: "BadCoordSet",
			rest: &restapi.Volume4D{Volume: restapi.Volume3D{OutlinePolygon: &restapi.Polygon{
				Vertices: []restapi.LatLngPoint{{Lat: 200, Lng: 0}, {Lat: 0, Lng: 0}, {Lat: 0, Lng: 1}},
			}}},
			wantErr: true,
		},
		{
			name: "NotEnoughPointsInPolygon",
			rest: &restapi.Volume4D{Volume: restapi.Volume3D{OutlinePolygon: &restapi.Polygon{
				Vertices: []restapi.LatLngPoint{{Lat: 0, Lng: 0}, {Lat: 0, Lng: 1}},
			}}},
			wantErr: true,
		},
		{
			name: "RadiusMustBeLargerThan0",
			rest: &restapi.Volume4D{Volume: restapi.Volume3D{OutlineCircle: &restapi.Circle{
				Center: &restapi.LatLngPoint{Lat: 37.427636, Lng: -122.170502},
				Radius: &restapi.Radius{Value: 0},
			}}},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := CellsVolume4DFromSCDRest(testCase.rest)
			if testCase.wantErr {
				require.Error(t, err)
				if actual != nil {
					assert.Equal(t, testCase.wantStart, actual.StartTime)
					assert.Equal(t, testCase.wantEnd, actual.EndTime)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.wantStart, actual.StartTime)
			assert.Equal(t, testCase.wantEnd, actual.EndTime)
			assert.Equal(t, testCase.wantAltLo, actual.AltitudeLo)
			assert.Equal(t, testCase.wantAltHi, actual.AltitudeHi)
			if testCase.wantCells {
				assert.NotEmpty(t, actual.Cells)
			} else {
				assert.Empty(t, actual.Cells)
			}
		})
	}
}

func TestUnionVolumes4DFromSCDRest(t *testing.T) {
	timeStart := time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC)
	timeMid := timeStart.Add(time.Minute)
	timeEnd := timeStart.Add(time.Hour)

	altLo := float32(100.0)
	altMid := float32(150.0)
	altHi := float32(200.0)

	testCases := []struct {
		name       string
		validators []Volume4DValidator
		rest       []restapi.Volume4D
		want       *dssmodels.Volume4D
		wantErr    bool
	}{
		{
			name: "Time",
			validators: []Volume4DValidator{
				WithRequireTimeBounds(),
				WithRequireEndTimeAfter(timeEnd.Add(-time.Minute)),
			},
			rest: []restapi.Volume4D{
				{TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeMid)},
				{TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeEnd)},
			},
			want: &dssmodels.Volume4D{SpatialVolume: &dssmodels.Volume3D{}, StartTime: &timeStart, EndTime: &timeEnd},
		},
		{
			name:       "TimeEndExpired",
			validators: []Volume4DValidator{WithRequireEndTimeAfter(timeEnd.Add(time.Minute))},
			rest: []restapi.Volume4D{
				{TimeEnd: newRestTime(timeMid)},
				{TimeEnd: newRestTime(timeEnd)},
			},
			wantErr: true,
		},
		{
			name:       "MissingTimeStart",
			validators: []Volume4DValidator{WithRequireTimeBounds()},
			rest: []restapi.Volume4D{
				{TimeEnd: newRestTime(timeMid)},
				{TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeEnd)},
			},
			wantErr: true,
		},
		{
			name:       "MissingTimeEnd",
			validators: []Volume4DValidator{WithRequireTimeBounds()},
			rest: []restapi.Volume4D{
				{TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeMid)},
				{TimeStart: newRestTime(timeStart)},
			},
			wantErr: true,
		},
		{
			name:       "Altitude",
			validators: []Volume4DValidator{WithRequireAltitudeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{AltitudeLower: newRestAlt(altLo), AltitudeUpper: newRestAlt(altMid)}},
				{Volume: restapi.Volume3D{AltitudeLower: newRestAlt(altMid), AltitudeUpper: newRestAlt(altHi)}},
			},
			want: &dssmodels.Volume4D{SpatialVolume: &dssmodels.Volume3D{AltitudeLo: &altLo, AltitudeHi: &altHi}},
		},
		{
			name:       "MissingLowerAltitude",
			validators: []Volume4DValidator{WithRequireAltitudeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{AltitudeUpper: newRestAlt(altMid)}},
				{Volume: restapi.Volume3D{AltitudeLower: newRestAlt(altMid), AltitudeUpper: newRestAlt(altHi)}},
			},
			wantErr: true,
		},
		{
			name:       "MissingUpperAltitude",
			validators: []Volume4DValidator{WithRequireAltitudeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{AltitudeLower: newRestAlt(altLo), AltitudeUpper: newRestAlt(altMid)}},
				{Volume: restapi.Volume3D{AltitudeLower: newRestAlt(altMid)}},
			},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := UnionVolumes4DFromSCDRest(testCase.rest, testCase.validators...)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, actual)
			}
		})
	}
}

func TestVolume4DFromSCDRest(t *testing.T) {
	start := time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	restInvalid := &restapi.Time{Value: start.Format(time.ANSIC)}

	testCases := []struct {
		name    string
		rest    *restapi.Volume4D
		want    *dssmodels.Volume4D
		wantErr bool
	}{
		{
			name: "Empty",
			rest: &restapi.Volume4D{},
			want: &dssmodels.Volume4D{SpatialVolume: &dssmodels.Volume3D{}},
		},
		{
			name: "Times",
			rest: &restapi.Volume4D{TimeStart: newRestTime(start), TimeEnd: newRestTime(end)},
			want: &dssmodels.Volume4D{SpatialVolume: &dssmodels.Volume3D{}, StartTime: &start, EndTime: &end},
		},
		{
			name:    "InvalidTimeStart",
			rest:    &restapi.Volume4D{TimeStart: restInvalid},
			wantErr: true,
		},
		{
			name:    "InvalidTimeEnd",
			rest:    &restapi.Volume4D{TimeEnd: restInvalid},
			wantErr: true,
		},
		{
			name:    "TimeStartAfterTimeEnd",
			rest:    &restapi.Volume4D{TimeStart: newRestTime(end), TimeEnd: newRestTime(start)},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := volume4DFromSCDRest(testCase.rest)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, actual)
			}
		})
	}
}

func TestVolume3DFromSCDRest(t *testing.T) {
	lo := float32(100.0)
	hi := float32(200.0)
	restInvalid := &restapi.Altitude{Value: 0}

	testCases := []struct {
		name    string
		rest    *restapi.Volume3D
		want    *dssmodels.Volume3D
		wantErr bool
	}{
		{
			name: "Empty",
			rest: &restapi.Volume3D{},
			want: &dssmodels.Volume3D{},
		},
		{
			name: "Polygon",
			rest: &restapi.Volume3D{
				OutlinePolygon: &restapi.Polygon{},
			},
			want: &dssmodels.Volume3D{
				Footprint: &dssmodels.GeoPolygon{},
			},
		},
		{
			name: "Circle",
			rest: &restapi.Volume3D{
				OutlineCircle: &restapi.Circle{
					Center: &restapi.LatLngPoint{},
					Radius: &restapi.Radius{},
				},
			},
			want: &dssmodels.Volume3D{
				Footprint: &dssmodels.GeoCircle{},
			},
		},
		{
			name: "Altitudes",
			rest: &restapi.Volume3D{AltitudeLower: newRestAlt(lo), AltitudeUpper: newRestAlt(hi)},
			want: &dssmodels.Volume3D{AltitudeLo: &lo, AltitudeHi: &hi},
		},
		{
			name:    "InvalidLowerAltitude",
			rest:    &restapi.Volume3D{AltitudeLower: restInvalid},
			wantErr: true,
		},
		{
			name:    "InvalidUpperAltitude",
			rest:    &restapi.Volume3D{AltitudeUpper: restInvalid},
			wantErr: true,
		},
		{
			name:    "LowerAltitudeGreaterThanUpperAltitude",
			rest:    &restapi.Volume3D{AltitudeLower: newRestAlt(hi), AltitudeUpper: newRestAlt(lo)},
			wantErr: true,
		},
		{
			name:    "MuiltiGeom",
			rest:    &restapi.Volume3D{OutlineCircle: &restapi.Circle{}, OutlinePolygon: &restapi.Polygon{}},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := Volume3DFromSCDRest(testCase.rest)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, actual)
			}
		})
	}
}
