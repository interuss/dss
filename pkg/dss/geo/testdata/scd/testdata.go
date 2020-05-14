package testdata

import (
	"github.com/interuss/dss/pkg/api/v1/scdpb"

	"github.com/golang/protobuf/ptypes"
)

var (
	Loop                           = `37.427636,-122.170502,37.408799,-122.064069,37.421265,-122.086504`
	LoopWithOddNumberOfCoordinates = `37.427636,-122.170502,37.408799`
	LoopWithOnlyTwoPoints          = `37.427636,-122.170502,37.408799,-122.064069`

	LoopPolygon = &scdpb.Polygon{
		Vertices: []*scdpb.LatLngPoint{
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

	LoopVolume3D = &scdpb.Volume3D{
		AltitudeUpper: &scdpb.Altitude{
			Value:     456,
			Reference: "W84",
		},
		AltitudeLower: &scdpb.Altitude{
			Value:     123,
			Reference: "W84",
		},
		OutlinePolygon: LoopPolygon,
	}

	LoopVolume4D = &scdpb.Volume4D{
		Volume: LoopVolume3D,
		TimeStart: &scdpb.Time{
			Value:  ptypes.TimestampNow(),
			Format: "RFC3339",
		},
		TimeEnd: &scdpb.Time{
			Value:  ptypes.TimestampNow(),
			Format: "RFC3339",
		},
	}
)
