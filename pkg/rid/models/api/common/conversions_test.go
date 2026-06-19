package common

import (
	"testing"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestFromGeoPolygonSring(t *testing.T) {
	testCases := []struct {
		name    string
		area    string
		wantErr bool
		want    *dssmodels.GeoPolygon
	}{
		{
			name: "OddNumberOfPoints",
			area: `1.1,1.2,2.1,2.2,3.1,3.2`,
			want: &dssmodels.GeoPolygon{Vertices: []*dssmodels.LatLngPoint{
				{Lat: 1.1, Lng: 1.2},
				{Lat: 2.1, Lng: 2.2},
				{Lat: 3.1, Lng: 3.2},
			}},
		},
		{
			name: "EvenNumberOfPoints",
			area: `1.1,1.2,2.1,2.2,3.1,3.2,4.1,4.2`,
			want: &dssmodels.GeoPolygon{Vertices: []*dssmodels.LatLngPoint{
				{Lat: 1.1, Lng: 1.2},
				{Lat: 2.1, Lng: 2.2},
				{Lat: 3.1, Lng: 3.2},
				{Lat: 4.1, Lng: 4.2},
			}},
		},
		{
			name:    "Empty",
			area:    "",
			wantErr: true,
		},
		{
			name:    "TwoPoints",
			area:    "1,2,3,4",
			wantErr: true,
		},
		{
			name:    "OddNumberOfCoordinates",
			area:    "1,2,3",
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			p, err := FromGeoPolygonSring(testCase.area)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.want, p)
			}
		})
	}
}
