package testdata

import (
	"github.com/steeling/InterUSS-Platform/pkg/dssv1"

	tspb "github.com/golang/protobuf/ptypes/timestamp"
)

var (
	Loop                           = `37.427636,-122.170502,37.408799,-122.064069,37.421265,-122.086504`
	LoopWithOddNumberOfCoordinates = `37.427636,-122.170502,37.408799`
	LoopWithOnlyTwoPoints          = `37.427636,-122.170502,37.408799,-122.064069`

	LoopPolygon = &dssv1.GeoPolygon{
		Vertices: []*dssv1.LatLngPoint{
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

	LoopVolume3D = &dssv1.Volume3D{
		AltitudeHi: 456,
		AltitudeLo: 123,
		Footprint:  LoopPolygon,
	}

	LoopVolume4D = &dssv1.Volume4D{
		SpatialVolume: LoopVolume3D,
		TimeStart: &tspb.Timestamp{
			Seconds: 10000,
		},
		TimeEnd: &tspb.Timestamp{
			Seconds: 10060,
		},
	}
)
