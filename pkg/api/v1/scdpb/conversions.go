// package scdpb provides functions to convert between scd geo objects and
// common model geo objects.
package scdpb

import (
	"errors"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	dssmodels "github.com/interuss/dss/pkg/dss/models"
)

var (
	errEmptyTime              = errors.New("time must not be empty")
	errPolygonNotSupportedYet = errors.New("polygon is not supported yet")

	unitToMeterMultiplicativeFactors = map[string]float32{
		"M": 1,
	}
)

func float32p(v float32) *float32 {
	return &v
}

func (t *Time) ToProto() (*timestamp.Timestamp, error) {
	if t == nil {
		return nil, errEmptyTime
	}
	return t.GetValue(), nil
}

func (vol4 *Volume4D) ToCommon() (*dssmodels.Volume4D, error) {
	vol3, err := vol4.GetVolume().ToCommon()
	if err != nil {
		return nil, err
	}

	result := &dssmodels.Volume4D{
		SpatialVolume: vol3,
	}

	if startTime := vol4.GetTimeStart(); startTime != nil {
		st, err := startTime.ToProto()
		if err != nil {
			return nil, err
		}
		ts, err := ptypes.Timestamp(st)
		if err != nil {
			return nil, err
		}
		result.StartTime = &ts
	}

	if endTime := vol4.GetTimeEnd(); endTime != nil {
		et, err := endTime.ToProto()
		if err != nil {
			return nil, err
		}
		ts, err := ptypes.Timestamp(et)
		if err != nil {
			return nil, err
		}
		result.EndTime = &ts
	}

	return result, nil
}

func (vol3 *Volume3D) ToCommon() (*dssmodels.Volume3D, error) {
	switch {
	case vol3.GetOutlineCircle() == nil && vol3.GetOutlinePolygon() == nil:
		return nil, errors.New("missing outline geometry")
	case vol3.GetOutlineCircle() != nil && vol3.GetOutlinePolygon() != nil:
		return nil, errors.New("both circle and polygon specified in outline geometry")
	case vol3.GetOutlinePolygon() != nil:
		return &dssmodels.Volume3D{
			Footprint: vol3.GetOutlinePolygon().ToCommon(),
		}, nil
	case vol3.GetOutlineCircle() != nil:
		return &dssmodels.Volume3D{
			Footprint: vol3.GetOutlineCircle().ToCommon(),
		}, nil
	}

	footprint := vol3.GetOutlinePolygon()
	if footprint == nil {
		return nil, errors.New("spatial_volume missing required footprint")
	}
	polygonFootprint := footprint.ToCommon()

	result := &dssmodels.Volume3D{
		Footprint:  polygonFootprint,
		AltitudeLo: float32p(float32(vol3.GetAltitudeLower().GetValue())),
		AltitudeHi: float32p(float32(vol3.GetAltitudeUpper().GetValue())),
	}

	return result, nil
}

func (c *Circle) ToCommon() *dssmodels.GeoCircle {
	return &dssmodels.GeoCircle{
		Center: dssmodels.LatLngPoint{
			Lat: 0., // FIXME(tvoss): Replace with c.GetGeometry().GetCoordinates().GetCoordinates()[1]
			Lng: 0., // FIXME(tvoss): Replace with c.GetGeometry().GetCoordinates().GetCoordinates()[0]
		},
		RadiusMeter: unitToMeterMultiplicativeFactors[c.GetProperties().GetRadius().GetUnits()] * c.GetProperties().GetRadius().GetValue(),
	}
}

func (p *Polygon) ToCommon() *dssmodels.GeoPolygon {
	result := &dssmodels.GeoPolygon{}
	for _, ltlng := range p.GetCoordinates() {
		result.Vertices = append(result.Vertices, ltlng.ToCommon())
	}

	return result
}

func (p *Point) ToCommon() *dssmodels.LatLngPoint {
	return &dssmodels.LatLngPoint{
		Lat: 0, // FIXME(tvoss): p.GetCoordinates()[1]
		Lng: 0, // FIXME(tvoss): p.GetCoordinates()[0]
	}
}

func FromVolume4D(vol4 *dssmodels.Volume4D) (*Volume4D, error) {
	vol3, err := FromVolume3D(vol4.SpatialVolume)
	if err != nil {
		return nil, err
	}

	result := &Volume4D{
		Volume: vol3,
	}

	if vol4.StartTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = &Time{
			Format: "RFC3339",
			Value:  ts,
		}
	}

	if vol4.EndTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd = &Time{
			Format: "RFC3339",
			Value:  ts,
		}
	}

	return result, nil
}

func FromVolume3D(vol3 *dssmodels.Volume3D) (*Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	result := &Volume3D{}

	if vol3.AltitudeLo != nil {
		result.AltitudeLower = &Altitude{
			Reference: "W84",
			Units:     "M",
			Value:     float64(*vol3.AltitudeLo),
		}
	}

	if vol3.AltitudeHi != nil {
		result.AltitudeUpper = &Altitude{
			Reference: "W84",
			Units:     "M",
			Value:     float64(*vol3.AltitudeHi),
		}
	}

	switch t := vol3.Footprint.(type) {
	case nil:
		// Empty on purpose
	case *dssmodels.GeoPolygon:
		result.OutlinePolygon = FromGeoPolygon(t)
	case *dssmodels.GeoCircle:
		result.OutlineCircle = FromGeoCircle(t)
	}

	return result, nil
}

func FromGeoCircle(circle *dssmodels.GeoCircle) *Circle {
	if circle == nil {
		return nil
	}

	return &Circle{
		Geometry: &Circle_GeometryMessage{
			Coordinates: FromLatLngPoint(&circle.Center),
			Type:        "Point",
		},
		Properties: &CircleProperties{
			Radius: &Radius{
				Units: "M",
				Value: circle.RadiusMeter,
			},
		},
	}
}
func FromGeoPolygon(polygon *dssmodels.GeoPolygon) *Polygon {
	if polygon == nil {
		return nil
	}

	result := &Polygon{
		Type: "Polygon",
	}

	for _, pt := range polygon.Vertices {
		result.Coordinates = append(result.Coordinates, FromLatLngPoint(pt))
	}

	return result
}

func FromLatLngPoint(pt *dssmodels.LatLngPoint) *Point {
	result := &Point{
		Type:        "Point",
		Coordinates: 0., // FIXME(tvoss): Replace with []float64{pt.Lng, pt.Lat}
	}

	return result
}
