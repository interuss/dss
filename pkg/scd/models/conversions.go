package models

import (
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

const (
	// TimeFormatRFC3339 is the string used for RFC3339
	TimeFormatRFC3339 = "RFC3339"
	UnitsM            = "M"
	ReferenceW84      = "W84"
)

type Volume4DValidator func(*dssmodels.Volume4D) error

func WithRequireTimeBounds() Volume4DValidator {
	return func(v *dssmodels.Volume4D) error {
		if v.StartTime == nil {
			return stacktrace.NewError("Missing start time")
		}
		if v.EndTime == nil {
			return stacktrace.NewError("Missing end time")
		}
		return nil
	}
}

func WithRequireEndTimeAfter(now time.Time) Volume4DValidator {
	return func(v *dssmodels.Volume4D) error {
		if v.EndTime != nil && v.EndTime.Before(now) {
			return stacktrace.NewError("End time may not be in the past")
		}
		return nil
	}
}

func WithRequireAltitudeBounds() Volume4DValidator {
	return func(v *dssmodels.Volume4D) error {
		if v.SpatialVolume.AltitudeLo == nil {
			return stacktrace.NewError("Missing lower altitude")
		}
		if v.SpatialVolume.AltitudeHi == nil {
			return stacktrace.NewError("Missing upper altitude")
		}
		return nil
	}
}

// UnionVolumes4DFromSCDRest converts a slice of vol4 SCD v1 REST model to a single bounding Volume4D
// Validation is applied on the resulting volume union
func UnionVolumes4DFromSCDRest(vol4s []restapi.Volume4D, validators ...Volume4DValidator) (*dssmodels.Volume4D, error) {
	volumes := make([]*dssmodels.Volume4D, len(vol4s))
	for idx, vol4 := range vol4s {
		volume, err := Volume4DFromSCDRest(&vol4)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to parse volume %d", idx)
		}
		volumes[idx] = volume
	}
	union, err := dssmodels.UnionVolumes4D(volumes...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to union volumes")
	}

	for _, validator := range validators {
		if err := validator(union); err != nil {
			return nil, stacktrace.Propagate(err, "Invalid volume union")
		}
	}

	return union, nil
}

// Volume4DFromSCDRest converts vol4 SCD v1 REST model to a Volume4D
func Volume4DFromSCDRest(vol4 *restapi.Volume4D) (*dssmodels.Volume4D, error) {
	vol3, err := Volume3DFromSCDRest(&vol4.Volume)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	var startTime *time.Time
	if vol4.TimeStart != nil {
		ts, err := time.Parse(time.RFC3339Nano, vol4.TimeStart.Value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time")
		}
		startTime = &ts
	}

	var endTime *time.Time
	if vol4.TimeEnd != nil {
		ts, err := time.Parse(time.RFC3339Nano, vol4.TimeEnd.Value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time")
		}
		endTime = &ts
	}

	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, stacktrace.NewError("Start time cannot be after end time")
	}

	volume := &dssmodels.Volume4D{
		SpatialVolume: vol3,
		StartTime:     startTime,
		EndTime:       endTime,
	}

	return volume, nil
}

// Volume3DFromSCDRest converts a vol3 SCD v1 REST model to a Volume3D
func Volume3DFromSCDRest(vol3 *restapi.Volume3D) (*dssmodels.Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	var altLo *float32
	if vol3.AltitudeLower != nil {
		if vol3.AltitudeLower.Units != UnitsM {
			return nil, stacktrace.NewError("Invalid lower altitude unit")
		}
		if vol3.AltitudeLower.Reference != ReferenceW84 {
			return nil, stacktrace.NewError("Invalid lower altitude reference")
		}
		altLo = new(float32(vol3.AltitudeLower.Value))
	}

	var altHi *float32
	if vol3.AltitudeUpper != nil {
		if vol3.AltitudeUpper.Units != UnitsM {
			return nil, stacktrace.NewError("Invalid upper altitude unit")
		}
		if vol3.AltitudeUpper.Reference != ReferenceW84 {
			return nil, stacktrace.NewError("Invalid upper altitude reference")
		}
		altHi = new(float32(vol3.AltitudeUpper.Value))
	}

	if altLo != nil && altHi != nil && *altLo > *altHi {
		return nil, stacktrace.NewError("Lower altitude cannot be greater than upper altitude")
	}

	var (
		footprint dssmodels.Geometry
		err       error
	)
	switch {
	case vol3.OutlineCircle != nil && vol3.OutlinePolygon != nil:
		err = stacktrace.NewError("Both circle and polygon specified in outline geometry")
	case vol3.OutlinePolygon != nil:
		footprint, err = GeoPolygonFromSCDRest(vol3.OutlinePolygon)
	case vol3.OutlineCircle != nil:
		footprint, err = GeoCircleFromSCDRest(vol3.OutlineCircle)
	}
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	return &dssmodels.Volume3D{
		Footprint:  footprint,
		AltitudeLo: altLo,
		AltitudeHi: altHi,
	}, nil
}

// GeoCircleFromSCDRest converts a circle SCD v1 REST model to a GeoCircle
func GeoCircleFromSCDRest(c *restapi.Circle) (*dssmodels.GeoCircle, error) {
	center, err := LatLngPointFromSCDRest(c.Center)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid circle center")
	}
	return &dssmodels.GeoCircle{
		Center:      *center,
		RadiusMeter: c.Radius.Value,
	}, nil
}

// GeoPolygonFromSCDRest converts a polygon SCD v1 REST model to a GeoPolygon
func GeoPolygonFromSCDRest(p *restapi.Polygon) (*dssmodels.GeoPolygon, error) {
	result := &dssmodels.GeoPolygon{}
	for _, ltlng := range p.Vertices {
		v, err := LatLngPointFromSCDRest(&ltlng)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Invalid polygon vertex")
		}
		result.Vertices = append(result.Vertices, v)
	}

	return result, nil
}

// LatLngPointFromSCDRest converts a point SCD v1 REST model to a latlngpoint
func LatLngPointFromSCDRest(p *restapi.LatLngPoint) (*dssmodels.LatLngPoint, error) {
	return dssmodels.NewLatLngPoint(float64(p.Lat), float64(p.Lng))
}
