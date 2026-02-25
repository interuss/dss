package models

import (
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/stacktrace"
)

type Volume4DOpts struct {
	RequireAltitudeBounds bool
	RequireTimeBounds     bool
}

func (opts *Volume4DOpts) to3DOpts() Volume3DOpts {
	return Volume3DOpts{
		RequireAltitudeBounds: opts.RequireAltitudeBounds,
	}
}

// Volume4DFromSCDRest converts vol4 SCD v1 REST model to a Volume4D
func Volume4DFromSCDRest(vol4 *restapi.Volume4D, opts Volume4DOpts) (*Volume4D, error) {
	vol3, err := Volume3DFromSCDRest(&vol4.Volume, opts.to3DOpts())
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
	} else if opts.RequireTimeBounds {
		return nil, stacktrace.NewError("Missing start time")
	}

	var endTime *time.Time
	if vol4.TimeEnd != nil {
		ts, err := time.Parse(time.RFC3339Nano, vol4.TimeEnd.Value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time")
		}
		endTime = &ts
	} else if opts.RequireTimeBounds {
		return nil, stacktrace.NewError("Missing end time")
	}

	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, stacktrace.NewError("Start time cannot be after end time")
	}

	return &Volume4D{
		SpatialVolume: vol3,
		StartTime:     startTime,
		EndTime:       endTime,
	}, nil
}

// UnionVolume4DFromSCDRest converts a slice of vol4 SCD v1 REST model to a single bounding Volume4D
func UnionVolume4DFromSCDRest(vol4s []restapi.Volume4D, opts Volume4DOpts) (*Volume4D, error) {
	volumes := make([]*Volume4D, len(vol4s))
	for idx, vol4 := range vol4s {
		volume, err := Volume4DFromSCDRest(&vol4, opts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to parse volume %d", idx)
		}
		volumes[idx] = volume
	}
	union, err := UnionVolumes4D(volumes...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to union volumes")
	}

	return union, nil
}

type Volume3DOpts struct {
	RequireAltitudeBounds bool
}

// Volume3DFromSCDRest converts a vol3 SCD v1 REST model to a Volume3D
func Volume3DFromSCDRest(vol3 *restapi.Volume3D, opts Volume3DOpts) (*Volume3D, error) {
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
		altLo = float32p(float32(vol3.AltitudeLower.Value))
	} else if opts.RequireAltitudeBounds {
		return nil, stacktrace.NewError("Missing lower altitude")
	}

	var altHi *float32
	if vol3.AltitudeUpper != nil {
		if vol3.AltitudeUpper.Units != UnitsM {
			return nil, stacktrace.NewError("Invalid upper altitude unit")
		}
		if vol3.AltitudeUpper.Reference != ReferenceW84 {
			return nil, stacktrace.NewError("Invalid upper altitude reference")
		}
		altHi = float32p(float32(vol3.AltitudeUpper.Value))
	} else if opts.RequireAltitudeBounds {
		return nil, stacktrace.NewError("Missing upper altitude")
	}

	if altLo != nil && altHi != nil && *altLo > *altHi {
		return nil, stacktrace.NewError("Lower altitude cannot be greater than upper altitude")
	}

	switch {
	case vol3.OutlineCircle != nil && vol3.OutlinePolygon != nil:
		return nil, stacktrace.NewError("Both circle and polygon specified in outline geometry")
	case vol3.OutlinePolygon != nil:
		return &Volume3D{
			Footprint:  GeoPolygonFromSCDRest(vol3.OutlinePolygon),
			AltitudeLo: altLo,
			AltitudeHi: altHi,
		}, nil
	case vol3.OutlineCircle != nil:
		return &Volume3D{
			Footprint:  GeoCircleFromSCDRest(vol3.OutlineCircle),
			AltitudeLo: altLo,
			AltitudeHi: altHi,
		}, nil
	}

	return &Volume3D{
		AltitudeLo: altLo,
		AltitudeHi: altHi,
	}, nil
}

// GeoCircleFromSCDRest converts a circle SCD v1 REST model to a GeoCircle
func GeoCircleFromSCDRest(c *restapi.Circle) *GeoCircle {
	return &GeoCircle{
		Center:      *LatLngPointFromSCDRest(c.Center),
		RadiusMeter: unitToMeterMultiplicativeFactors[unit(c.Radius.Units)] * c.Radius.Value,
	}
}

// GeoPolygonFromSCDRest converts a polygon SCD v1 REST model to a GeoPolygon
func GeoPolygonFromSCDRest(p *restapi.Polygon) *GeoPolygon {
	result := &GeoPolygon{}
	for _, ltlng := range p.Vertices {
		result.Vertices = append(result.Vertices, LatLngPointFromSCDRest(&ltlng))
	}

	return result
}

// LatLngPointFromSCDRest converts a point SCD v1 REST model to a latlngpoint
func LatLngPointFromSCDRest(p *restapi.LatLngPoint) *LatLngPoint {
	return &LatLngPoint{
		Lat: float64(p.Lat),
		Lng: float64(p.Lng),
	}
}

// ToSCDRest converts the Volume4D to a SCD v1 REST model
func (vol4 *Volume4D) ToSCDRest() *restapi.Volume4D {

	result := &restapi.Volume4D{}
	if vol4.SpatialVolume != nil {
		result.Volume = *vol4.SpatialVolume.ToSCDRest()
	}

	if vol4.StartTime != nil {
		result.TimeStart = &restapi.Time{
			Format: TimeFormatRFC3339,
			Value:  vol4.StartTime.Format(time.RFC3339Nano),
		}
	}

	if vol4.EndTime != nil {
		result.TimeEnd = &restapi.Time{
			Format: TimeFormatRFC3339,
			Value:  vol4.EndTime.Format(time.RFC3339Nano),
		}
	}

	return result
}

// ToSCDRest converts the Volume3D to a SCD v1 REST model
func (vol3 *Volume3D) ToSCDRest() *restapi.Volume3D {
	if vol3 == nil {
		return nil
	}

	result := &restapi.Volume3D{}

	if vol3.AltitudeLo != nil {
		result.AltitudeLower = &restapi.Altitude{
			Reference: altitudeReferenceWGS84.String(),
			Units:     unitMeter.String(),
			Value:     float64(*vol3.AltitudeLo),
		}
	}

	if vol3.AltitudeHi != nil {
		result.AltitudeUpper = &restapi.Altitude{
			Reference: altitudeReferenceWGS84.String(),
			Units:     unitMeter.String(),
			Value:     float64(*vol3.AltitudeHi),
		}
	}

	switch t := vol3.Footprint.(type) {
	case nil:
		// Empty on purpose
	case *GeoPolygon:
		result.OutlinePolygon = t.ToSCDRest()
	case *GeoCircle:
		result.OutlineCircle = t.ToSCDRest()
	}

	return result
}

// ToSCDRest converts the GeoCircle to a SCD v1 REST model
func (gc *GeoCircle) ToSCDRest() *restapi.Circle {
	if gc == nil {
		return nil
	}

	return &restapi.Circle{
		Center: gc.Center.ToSCDRest(),
		Radius: &restapi.Radius{
			Units: unitMeter.String(),
			Value: gc.RadiusMeter,
		},
	}
}

// ToSCDRest converts the GeoPolygon to a SCD v1 REST model
func (gp *GeoPolygon) ToSCDRest() *restapi.Polygon {
	if gp == nil {
		return nil
	}

	result := &restapi.Polygon{
		Vertices: make([]restapi.LatLngPoint, len(gp.Vertices)),
	}

	for _, pt := range gp.Vertices {
		result.Vertices = append(result.Vertices, *pt.ToSCDRest())
	}

	return result
}

// ToSCDRest converts the LatLngPoint to a SCD v1 REST model
func (pt *LatLngPoint) ToSCDRest() *restapi.LatLngPoint {
	result := &restapi.LatLngPoint{
		Lat: restapi.Latitude(pt.Lat),
		Lng: restapi.Longitude(pt.Lng),
	}

	return result
}
