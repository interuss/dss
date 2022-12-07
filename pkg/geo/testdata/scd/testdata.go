package testdata

import (
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
)

var (
	Loop                           = `37.427636,-122.170502,37.408799,-122.064069,37.421265,-122.086504`
	LoopWithOddNumberOfCoordinates = `37.427636,-122.170502,37.408799`
	LoopWithOnlyTwoPoints          = `37.427636,-122.170502,37.408799,-122.064069`

	LoopPolygon = &restapi.Polygon{
		Vertices: []restapi.LatLngPoint{
			{
				Lat: 37.427636,
				Lng: -122.170502,
			},
			{
				Lat: 37.408799,
				Lng: -122.064069,
			},
			{
				Lat: 37.421265,
				Lng: -122.086504,
			},
		},
	}

	LoopVolume3D = restapi.Volume3D{
		AltitudeUpper: &restapi.Altitude{
			Value:     456,
			Reference: "W84",
		},
		AltitudeLower: &restapi.Altitude{
			Value:     123,
			Reference: "W84",
		},
		OutlinePolygon: LoopPolygon,
	}

	LoopVolume4D = &restapi.Volume4D{
		Volume: LoopVolume3D,
		TimeStart: &restapi.Time{
			Value:  time.Now().Format(time.RFC3339Nano),
			Format: "RFC3339",
		},
		TimeEnd: &restapi.Time{
			Value:  time.Now().Format(time.RFC3339Nano),
			Format: "RFC3339",
		},
	}
)
