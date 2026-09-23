package models

import (
	"testing"
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
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

func TestUnionCellsVolume4DFromSCDRest(t *testing.T) {
	timeStart := time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC)
	timeMid := timeStart.Add(time.Minute)
	timeEnd := timeStart.Add(time.Hour)

	altLo := float32(100.0)
	altMid := float32(150.0)
	altHi := float32(200.0)

	testCases := []struct {
		name       string
		validators []CellsVolume4DValidator
		rest       []restapi.Volume4D
		wantStart  *time.Time
		wantEnd    *time.Time
		wantAltLo  *float32
		wantAltHi  *float32
		wantErr    bool
	}{
		{
			name: "Time",
			validators: []CellsVolume4DValidator{
				WithRequireCellsTimeBounds(),
				WithRequireCellsEndTimeAfter(timeEnd.Add(-time.Minute)),
			},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeMid)},
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeEnd)},
			},
			wantStart: &timeStart,
			wantEnd:   &timeEnd,
		},
		{
			name:       "TimeEndExpired",
			validators: []CellsVolume4DValidator{WithRequireCellsEndTimeAfter(timeEnd.Add(time.Minute))},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeEnd: newRestTime(timeMid)},
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeEnd: newRestTime(timeEnd)},
			},
			wantErr: true,
		},
		{
			name:       "MissingTimeStart",
			validators: []CellsVolume4DValidator{WithRequireCellsTimeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeEnd: newRestTime(timeMid)},
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeEnd)},
			},
			wantErr: true,
		},
		{
			name:       "MissingTimeEnd",
			validators: []CellsVolume4DValidator{WithRequireCellsTimeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeStart: newRestTime(timeStart), TimeEnd: newRestTime(timeMid)},
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint}, TimeStart: newRestTime(timeStart)},
			},
			wantErr: true,
		},
		{
			name:       "Altitude",
			validators: []CellsVolume4DValidator{WithRequireCellsAltitudeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint, AltitudeLower: newRestAlt(altLo), AltitudeUpper: newRestAlt(altMid)}},
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint, AltitudeLower: newRestAlt(altMid), AltitudeUpper: newRestAlt(altHi)}},
			},
			wantAltLo: &altLo,
			wantAltHi: &altHi,
		},
		{
			name:       "MissingLowerAltitude",
			validators: []CellsVolume4DValidator{WithRequireCellsAltitudeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint, AltitudeUpper: newRestAlt(altMid)}},
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint, AltitudeLower: newRestAlt(altMid), AltitudeUpper: newRestAlt(altHi)}},
			},
			wantErr: true,
		},
		{
			name:       "MissingUpperAltitude",
			validators: []CellsVolume4DValidator{WithRequireCellsAltitudeBounds()},
			rest: []restapi.Volume4D{
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint, AltitudeLower: newRestAlt(altLo), AltitudeUpper: newRestAlt(altMid)}},
				{Volume: restapi.Volume3D{OutlinePolygon: &footprint, AltitudeLower: newRestAlt(altMid)}},
			},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := UnionCellsVolume4DFromSCDRest(testCase.rest, testCase.validators...)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.wantStart, actual.StartTime)
				assert.Equal(t, testCase.wantEnd, actual.EndTime)
				assert.Equal(t, testCase.wantAltLo, actual.AltitudeLo)
				assert.Equal(t, testCase.wantAltHi, actual.AltitudeHi)
				assert.NotEmpty(t, actual.Cells)
			}
		})
	}
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
			name: "Empty",
			rest: &restapi.Volume4D{},
			// no footprint: ErrMissingFootprint, but time/altitude are still populated (nil here).
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
