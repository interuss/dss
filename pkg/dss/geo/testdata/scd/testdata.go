package testdata

import (
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	"time"
)

var (
	LoopPolygon = &dssmodels.GeoPolygon{
		Vertices: []*dssmodels.LatLngPoint{
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

	AltitudeHi = float32(200)
	AltitudeLo = float32(100)

	LoopVolume3D = &dssmodels.Volume3D{
		AltitudeHi: &AltitudeHi,
		AltitudeLo: &AltitudeLo,
		Footprint:  LoopPolygon,
	}

	LoopVolume4D = &dssmodels.Volume4D{
		SpatialVolume: LoopVolume3D,
		StartTime:     &[]time.Time{time.Unix(10000, 0)}[0],
		EndTime:       &[]time.Time{time.Unix(10060, 0)}[0],
	}

	LoopVolume4DWithOnlyTwoPoints = &dssmodels.Volume4D{
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeHi: &AltitudeHi,
			AltitudeLo: &AltitudeLo,
			Footprint: &dssmodels.GeoPolygon{
				Vertices: []*dssmodels.LatLngPoint{
					{
						Lat: 37.427636,
						Lng: -122.170502,
					},
					{
						Lat: 37.408799,
						Lng: -122.064069,
					},
				},
			},
		},
		StartTime: &[]time.Time{time.Unix(10000, 0)}[0],
		EndTime:   &[]time.Time{time.Unix(10060, 0)}[0],
	}
)
