package testdata

import (
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
)

var (
	Owner       = "foo"
	Version, _  = dssmodels.VersionFromString("bar")
	CallbackURL = restapi.IdentificationServiceAreaURL("https://example.com")

	Loop                           = `37.427636,-122.170502,37.408799,-122.064069,37.421265,-122.086504`
	LoopWithOddNumberOfCoordinates = `37.427636,-122.170502,37.408799`
	LoopWithOnlyTwoPoints          = `37.427636,-122.170502,37.408799,-122.064069`

	LoopPolygon = restapi.GeoPolygon{
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

	AltitudeHi = restapi.Altitude(456)
	AltitudeLo = restapi.Altitude(123)

	LoopVolume3D = restapi.Volume3D{
		AltitudeHi: &AltitudeHi,
		AltitudeLo: &AltitudeLo,
		Footprint:  LoopPolygon,
	}

	TimeStart = time.Unix(10000, 0).Format(time.RFC3339Nano)
	TimeEnd   = time.Unix(10060, 0).Format(time.RFC3339Nano)

	LoopVolume4D = restapi.Volume4D{
		SpatialVolume: LoopVolume3D,
		TimeStart:     &TimeStart,
		TimeEnd:       &TimeEnd,
	}
)
