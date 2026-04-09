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
		want       *Volume4D
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
			want: &Volume4D{SpatialVolume: &Volume3D{}, StartTime: &timeStart, EndTime: &timeEnd},
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
			want: &Volume4D{SpatialVolume: &Volume3D{AltitudeLo: &altLo, AltitudeHi: &altHi}},
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
		want    *Volume4D
		wantErr bool
	}{
		{
			name: "Empty",
			rest: &restapi.Volume4D{},
			want: &Volume4D{SpatialVolume: &Volume3D{}},
		},
		{
			name: "Times",
			rest: &restapi.Volume4D{TimeStart: newRestTime(start), TimeEnd: newRestTime(end)},
			want: &Volume4D{SpatialVolume: &Volume3D{}, StartTime: &start, EndTime: &end},
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
			actual, err := Volume4DFromSCDRest(testCase.rest)
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
		want    *Volume3D
		wantErr bool
	}{
		{
			name: "Empty",
			rest: &restapi.Volume3D{},
			want: &Volume3D{},
		},
		{
			name: "Polygon",
			rest: &restapi.Volume3D{
				OutlinePolygon: &restapi.Polygon{},
			},
			want: &Volume3D{
				Footprint: &GeoPolygon{},
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
			want: &Volume3D{
				Footprint: &GeoCircle{},
			},
		},
		{
			name: "Altitudes",
			rest: &restapi.Volume3D{AltitudeLower: newRestAlt(lo), AltitudeUpper: newRestAlt(hi)},
			want: &Volume3D{AltitudeLo: &lo, AltitudeHi: &hi},
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
