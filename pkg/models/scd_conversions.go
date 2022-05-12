package models

import (
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/stacktrace"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// Volume4DFromSCDProto converts vol4 proto to a Volume4D
func Volume4DFromSCDProto(vol4 *scdpb.Volume4D) (*Volume4D, error) {
	vol3, err := Volume3DFromSCDProto(vol4.GetVolume())
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	result := &Volume4D{
		SpatialVolume: vol3,
	}

	if startTime := vol4.GetTimeStart(); startTime != nil {
		st := startTime.GetValue()
		ts := st.AsTime()
		err := st.CheckValid()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time from proto")
		}
		result.StartTime = &ts
	}

	if endTime := vol4.GetTimeEnd(); endTime != nil {
		et := endTime.GetValue()
		ts := et.AsTime()
		err := et.CheckValid()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time from proto")
		}
		result.EndTime = &ts
	}

	return result, nil
}

// Volume3DFromSCDProto converts a vol3 proto to a Volume3D
func Volume3DFromSCDProto(vol3 *scdpb.Volume3D) (*Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	altitudeLower := vol3.GetAltitudeLower()
	var altLo *float32
	if altitudeLower != nil {
		if altitudeLower.Units != UnitsM {
			return nil, stacktrace.NewError("Invalid lower altitude unit")
		}
		if altitudeLower.Reference != ReferenceW84 {
			return nil, stacktrace.NewError("Invalid lower altitude reference")
		}
		altLo = float32p(float32(altitudeLower.GetValue()))
	}

	altitudeUpper := vol3.GetAltitudeUpper()
	var altHi *float32
	if altitudeUpper != nil {
		if altitudeUpper.Units != UnitsM {
			return nil, stacktrace.NewError("Invalid upper altitude unit")
		}
		if altitudeUpper.Reference != ReferenceW84 {
			return nil, stacktrace.NewError("Invalid upper altitude reference")
		}
		altHi = float32p(float32(altitudeUpper.GetValue()))
	}

	switch {
	case vol3.GetOutlineCircle() != nil && vol3.GetOutlinePolygon() != nil:
		return nil, stacktrace.NewError("Both circle and polygon specified in outline geometry")
	case vol3.GetOutlinePolygon() != nil:
		return &Volume3D{
			Footprint:  GeoPolygonFromSCDProto(vol3.GetOutlinePolygon()),
			AltitudeLo: altLo,
			AltitudeHi: altHi,
		}, nil
	case vol3.GetOutlineCircle() != nil:
		return &Volume3D{
			Footprint:  GeoCircleFromSCDProto(vol3.GetOutlineCircle()),
			AltitudeLo: altLo,
			AltitudeHi: altHi,
		}, nil
	}

	return &Volume3D{
		AltitudeLo: altLo,
		AltitudeHi: altHi,
	}, nil
}

// GeoCircleFromSCDProto converts a circle proto to a GeoCircle
func GeoCircleFromSCDProto(c *scdpb.Circle) *GeoCircle {
	return &GeoCircle{
		Center:      *LatLngPointFromSCDProto(c.GetCenter()),
		RadiusMeter: unitToMeterMultiplicativeFactors[unit(c.GetRadius().GetUnits())] * c.GetRadius().GetValue(),
	}
}

// GeoPolygonFromSCDProto converts a polygon proto to a GeoPolygon
func GeoPolygonFromSCDProto(p *scdpb.Polygon) *GeoPolygon {
	result := &GeoPolygon{}
	for _, ltlng := range p.GetVertices() {
		result.Vertices = append(result.Vertices, LatLngPointFromSCDProto(ltlng))
	}

	return result
}

// LatLngPointFromSCDProto converts a point proto to a latlngpoint
func LatLngPointFromSCDProto(p *scdpb.LatLngPoint) *LatLngPoint {
	return &LatLngPoint{
		Lat: p.GetLat(),
		Lng: p.GetLng(),
	}
}

// ToSCDProto converts the Volume4D to a proto
func (vol4 *Volume4D) ToSCDProto() (*scdpb.Volume4D, error) {
	vol3, err := vol4.SpatialVolume.ToSCDProto()
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	result := &scdpb.Volume4D{
		Volume: vol3,
	}

	if vol4.StartTime != nil {
		ts := tspb.New(*vol4.StartTime)
		result.TimeStart = &scdpb.Time{
			Format: TimeFormatRFC3339,
			Value:  ts,
		}
	}

	if vol4.EndTime != nil {
		ts := tspb.New(*vol4.EndTime)
		result.TimeEnd = &scdpb.Time{
			Format: TimeFormatRFC3339,
			Value:  ts,
		}
	}

	return result, nil
}

// ToSCDProto converts the Volume3D to a proto
func (vol3 *Volume3D) ToSCDProto() (*scdpb.Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	result := &scdpb.Volume3D{}

	if vol3.AltitudeLo != nil {
		result.AltitudeLower = &scdpb.Altitude{
			Reference: altitudeReferenceWGS84.String(),
			Units:     unitMeter.String(),
			Value:     float64(*vol3.AltitudeLo),
		}
	}

	if vol3.AltitudeHi != nil {
		result.AltitudeUpper = &scdpb.Altitude{
			Reference: altitudeReferenceWGS84.String(),
			Units:     unitMeter.String(),
			Value:     float64(*vol3.AltitudeHi),
		}
	}

	switch t := vol3.Footprint.(type) {
	case nil:
		// Empty on purpose
	case *GeoPolygon:
		result.OutlinePolygon = t.ToSCDProto()
	case *GeoCircle:
		result.OutlineCircle = t.ToSCDProto()
	}

	return result, nil
}

// ToSCDProto converts the GeoCircle to a proto
func (gc *GeoCircle) ToSCDProto() *scdpb.Circle {
	if gc == nil {
		return nil
	}

	return &scdpb.Circle{
		Center: gc.Center.ToSCDProto(),
		Radius: &scdpb.Radius{
			Units: unitMeter.String(),
			Value: gc.RadiusMeter,
		},
	}
}

// ToSCDProto converts the GeoPolygon to a proto
func (gp *GeoPolygon) ToSCDProto() *scdpb.Polygon {
	if gp == nil {
		return nil
	}

	result := &scdpb.Polygon{}

	for _, pt := range gp.Vertices {
		result.Vertices = append(result.Vertices, pt.ToSCDProto())
	}

	return result
}

// ToSCDProto converts the LatLngPoint to a proto
func (pt *LatLngPoint) ToSCDProto() *scdpb.LatLngPoint {
	result := &scdpb.LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}

	return result
}
