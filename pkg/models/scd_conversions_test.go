package models

import (
	"testing"
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/stretchr/testify/assert"
)

func TestVolume4DFromSCDRest(t *testing.T) {
	start := time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	timeStart := restapi.Time{Value: start.Format(time.RFC3339), Format: TimeFormatRFC3339}
	timeEnd := restapi.Time{Value: end.Format(time.RFC3339), Format: TimeFormatRFC3339}
	invalid := restapi.Time{Value: start.Format(time.ANSIC)}
	testCases := []struct {
		name    string
		opts    Volume4DOpts
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
			opts: Volume4DOpts{RequireTimeBounds: true},
			rest: &restapi.Volume4D{TimeStart: &timeStart, TimeEnd: &timeEnd},
			want: &Volume4D{
				SpatialVolume: &Volume3D{},
				StartTime:     &start,
				EndTime:       &end,
			},
		},
		{
			name:    "InvalidTimeStart",
			rest:    &restapi.Volume4D{TimeStart: &invalid},
			wantErr: true,
		},
		{
			name:    "InvalidTimeEnd",
			rest:    &restapi.Volume4D{TimeEnd: &invalid},
			wantErr: true,
		},
		{
			name:    "TimeStartAfterTimeEnd",
			rest:    &restapi.Volume4D{TimeStart: &timeEnd, TimeEnd: &timeStart},
			wantErr: true,
		},
		{
			name:    "MissingTimeStart",
			opts:    Volume4DOpts{RequireTimeBounds: true},
			rest:    &restapi.Volume4D{TimeEnd: &timeEnd},
			wantErr: true,
		},
		{
			name:    "MissingTimeEnd",
			opts:    Volume4DOpts{RequireTimeBounds: true},
			rest:    &restapi.Volume4D{TimeStart: &timeStart},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := Volume4DFromSCDRest(testCase.rest, testCase.opts)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, testCase.want, actual)
			}
		})
	}
}

func TestVolume3DFromSCDRest(t *testing.T) {
	lower := restapi.Altitude{
		Value:     100.0,
		Reference: ReferenceW84,
		Units:     UnitsM,
	}
	lo := float32(100.0)
	upper := restapi.Altitude{
		Value:     200.0,
		Reference: ReferenceW84,
		Units:     UnitsM,
	}
	hi := float32(200.0)
	invalid := restapi.Altitude{Value: 0}

	testCases := []struct {
		name    string
		opts    Volume3DOpts
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
			rest: &restapi.Volume3D{AltitudeLower: &lower, AltitudeUpper: &upper},
			opts: Volume3DOpts{RequireAltitudeBounds: true},
			want: &Volume3D{AltitudeLo: &lo, AltitudeHi: &hi},
		},
		{
			name:    "InvalidLowerAltitude",
			rest:    &restapi.Volume3D{AltitudeLower: &invalid},
			wantErr: true,
		},
		{
			name:    "InvalidUpperAltitude",
			rest:    &restapi.Volume3D{AltitudeUpper: &invalid},
			wantErr: true,
		},
		{
			name:    "LowerAltitudeGreaterThanUpperAltitude",
			rest:    &restapi.Volume3D{AltitudeLower: &upper, AltitudeUpper: &lower},
			wantErr: true,
		},
		{
			name:    "MissingLowerAltitude",
			opts:    Volume3DOpts{RequireAltitudeBounds: true},
			rest:    &restapi.Volume3D{AltitudeUpper: &upper},
			wantErr: true,
		},
		{
			name:    "MissingUpperAltitude",
			opts:    Volume3DOpts{RequireAltitudeBounds: true},
			rest:    &restapi.Volume3D{AltitudeLower: &lower},
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
			actual, err := Volume3DFromSCDRest(testCase.rest, testCase.opts)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, testCase.want, actual)
			}
		})
	}
}
