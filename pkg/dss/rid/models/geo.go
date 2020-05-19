package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/interuss/dss/pkg/dss/geo"

	// TODO: all of the uses of ridpb should use protos that can be used
	// by both rid and scd.
	"github.com/interuss/dss/pkg/api/v1/ridpb"
)

const (
	minLat = -90.0
	maxLat = 90.0
	minLng = -180.0
	maxLng = 180.0
)

var (
	// ErrMissingSpatialVolume indicates that a spatial volume is required but
	// missing to complete an operation.
	ErrMissingSpatialVolume = errors.New("missing spatial volume")
	// ErrMissingFootprint indicates that a geometry footprint is required but
	// missing to complete an operation.
	ErrMissingFootprint = errors.New("missing footprint")

	errNotEnoughPointsInPolygon = errors.New("not enough points in polygon")
	errBadCoordSet              = errors.New("coordinates did not create a well formed area")
	errRadiusMustBeLargerThan0  = errors.New("radius must be larger than 0")
)

func timeP(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func float32P(v float32) *float32 {
	return &v
}

// Contiguous block of geographic spacetime.
type Volume4D struct {
	// Constant spatial extent of this volume.
	SpatialVolume *Volume3D
	// End time of this volume.
	EndTime *time.Time
	// Beginning time of this volume.
	StartTime *time.Time
}

// A three-dimensional geographic volume consisting of a vertically-extruded shape.
type Volume3D struct {
	// Maximum bounding altitude (meters above the WGS84 ellipsoid) of this volume.
	AltitudeHi *float32
	// Minimum bounding altitude (meters above the WGS84 ellipsoid) of this volume.
	AltitudeLo *float32
	// Projection of this volume onto the earth's surface.
	Footprint Geometry
}

// Geometry models a geometry.
type Geometry interface {
	// CalculateCovering returns an s2 cell covering for a geometry.
	CalculateCovering() (s2.CellUnion, error)
}

// GeometryFunc is an implementation of Geometry
type GeometryFunc func() (s2.CellUnion, error)

type precomputedCellGeometry map[s2.CellID]struct{}

func (pcg precomputedCellGeometry) merge(ids ...s2.CellID) precomputedCellGeometry {
	for _, id := range ids {
		pcg[id] = struct{}{}
	}
	return pcg
}

func (pcg precomputedCellGeometry) CalculateCovering() (s2.CellUnion, error) {
	var (
		result = make(s2.CellUnion, len(pcg))
		idx    int
	)

	for id := range pcg {
		result[idx] = id
		idx++
	}

	return result, nil
}

// UnionVolumes4D unions volumes and returns a volume that covers all the
// individual volumes in space and time.
func UnionVolumes4D(volumes ...*Volume4D) (*Volume4D, error) {
	result := &Volume4D{}

	for _, volume := range volumes {
		if volume.EndTime != nil {
			if result.EndTime != nil {
				if volume.EndTime.After(*result.EndTime) {
					*result.EndTime = *volume.EndTime
				}
			} else {
				result.EndTime = timeP(*volume.EndTime)
			}
		}

		if volume.StartTime != nil {
			if result.StartTime != nil {
				if volume.StartTime.Before(*result.StartTime) {
					*result.StartTime = *volume.StartTime
				}
			} else {
				result.StartTime = timeP(*volume.StartTime)
			}
		}

		if volume.SpatialVolume != nil {
			if result.SpatialVolume == nil {
				result.SpatialVolume = &Volume3D{}
			}

			if volume.SpatialVolume.AltitudeLo != nil {
				if result.SpatialVolume.AltitudeLo != nil {
					if *volume.SpatialVolume.AltitudeLo < *result.SpatialVolume.AltitudeLo {
						*result.SpatialVolume.AltitudeLo = *volume.SpatialVolume.AltitudeLo
					}
				} else {
					result.SpatialVolume.AltitudeLo = float32P(*volume.SpatialVolume.AltitudeLo)
				}
			}

			if volume.SpatialVolume.AltitudeHi != nil {
				if result.SpatialVolume.AltitudeHi != nil {
					if *volume.SpatialVolume.AltitudeHi > *result.SpatialVolume.AltitudeHi {
						*result.SpatialVolume.AltitudeHi = *volume.SpatialVolume.AltitudeHi
					}
				} else {
					result.SpatialVolume.AltitudeHi = float32P(*volume.SpatialVolume.AltitudeHi)
				}
			}

			if volume.SpatialVolume.Footprint != nil {
				cells, err := volume.SpatialVolume.Footprint.CalculateCovering()
				if err != nil {
					return nil, err
				}

				if result.SpatialVolume.Footprint == nil {
					result.SpatialVolume.Footprint = precomputedCellGeometry{}
				}
				result.SpatialVolume.Footprint.(precomputedCellGeometry).merge(cells...)
			}
		}
	}

	return result, nil
}

// CalculateSpatialCovering returns the spatial covering of v4d.
func (v4d *Volume4D) CalculateSpatialCovering() (s2.CellUnion, error) {
	switch {
	case v4d.SpatialVolume == nil:
		return nil, ErrMissingSpatialVolume
	default:
		return v4d.SpatialVolume.CalculateCovering()
	}
}

// CalculateCovering returns the spatial covering of v3d.
func (v3d *Volume3D) CalculateCovering() (s2.CellUnion, error) {
	switch {
	case v3d.Footprint == nil:
		return nil, ErrMissingFootprint
	default:
		return v3d.Footprint.CalculateCovering()
	}
}

// CalculateCovering returns the result of invoking gf.
func (gf GeometryFunc) CalculateCovering() (s2.CellUnion, error) {
	return gf()
}

// GeoCircle models a circular enclosed area on earth's surface.
type GeoCircle struct {
	Center      LatLngPoint
	RadiusMeter float32
}

func (gc *GeoCircle) CalculateCovering() (s2.CellUnion, error) {
	if (gc.Center.Lat > maxLat) || (gc.Center.Lat < minLat) || (gc.Center.Lng > maxLng) || (gc.Center.Lng < minLng) {
		return nil, errBadCoordSet
	}

	if !(gc.RadiusMeter > 0) {
		return nil, errRadiusMustBeLargerThan0
	}

	return geo.CoveringForLoop(s2.RegularLoop(
		s2.PointFromLatLng(s2.LatLngFromDegrees(gc.Center.Lat, gc.Center.Lng)),
		geo.DistanceMetersToAngle(float64(gc.RadiusMeter)),
		20,
	))
}

// GeoPolygon models an enclosed area on the earth.
// The bounding edges of this polygon shall be the shortest paths between connected vertices.  This means, for instance, that the edge between two points both defined at a particular latitude is not generally contained at that latitude.
// The winding order shall be interpreted as the order which produces the smaller area.
// The path between two vertices shall be the shortest possible path between those vertices.
// Edges may not cross.
// Vertices may not be duplicated.  In particular, the final polygon vertex shall not be identical to the first vertex.
type GeoPolygon struct {
	Vertices []*LatLngPoint
}

func (gp *GeoPolygon) CalculateCovering() (s2.CellUnion, error) {
	var points []s2.Point
	if gp == nil {
		return nil, errBadCoordSet
	}
	for _, v := range gp.Vertices {
		// ensure that coordinates passed are actually on earth
		if (v.Lat > maxLat) || (v.Lat < minLat) || (v.Lng > maxLng) || (v.Lng < minLng) {
			return nil, errBadCoordSet
		}
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(v.Lat, v.Lng)))
	}
	if len(points) < 3 {
		return nil, errNotEnoughPointsInPolygon
	}
	return geo.Covering(points)
}

// LatLngPoint models a point on the earth's surface.
type LatLngPoint struct {
	Lat float64
	Lng float64
}

// Volume4DFromProto convert proto to model object
func Volume4DFromProto(vol4 *ridpb.Volume4D) (*Volume4D, error) {
	vol3, err := Volume3DFromProto(vol4.GetSpatialVolume())
	if err != nil {
		return nil, err
	}

	result := &Volume4D{
		SpatialVolume: vol3,
	}

	if startTime := vol4.GetTimeStart(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return nil, err
		}
		result.StartTime = &ts
	}

	if endTime := vol4.GetTimeEnd(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return nil, err
		}
		result.EndTime = &ts
	}

	return result, nil
}

// Volume3DFromProto convert proto to model object
func Volume3DFromProto(vol3 *ridpb.Volume3D) (*Volume3D, error) {
	footprint := vol3.GetFootprint()
	if footprint == nil {
		return nil, errors.New("spatial_volume missing required footprint")
	}
	polygonFootprint := GeoPolygonFromProto(footprint)

	result := &Volume3D{
		Footprint:  polygonFootprint,
		AltitudeLo: proto.Float32(vol3.GetAltitudeLo()),
		AltitudeHi: proto.Float32(vol3.GetAltitudeHi()),
	}

	return result, nil
}

// GeoPolygonFromProto convert proto to model object
func GeoPolygonFromProto(footprint *ridpb.GeoPolygon) *GeoPolygon {
	result := &GeoPolygon{}

	for _, ltlng := range footprint.Vertices {
		result.Vertices = append(result.Vertices, PointFromProto(ltlng))
	}

	return result
}

// PointFromProto convert proto to model object
func PointFromProto(pt *ridpb.LatLngPoint) *LatLngPoint {
	return &LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}
}

// Business -> RID

// ToProto converts Volume4D model obj to proto
func (vol4 *Volume4D) ToProto() (*ridpb.Volume4D, error) {
	vol3, err := vol4.SpatialVolume.ToProto()
	if err != nil {
		return nil, err
	}

	result := &ridpb.Volume4D{
		SpatialVolume: vol3,
	}

	if vol4.StartTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = ts
	}

	if vol4.EndTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd = ts
	}

	return result, nil
}

// ToProto converts Volume3D model obj to proto
func (vol3 *Volume3D) ToProto() (*ridpb.Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	result := &ridpb.Volume3D{}

	if vol3.AltitudeLo != nil {
		result.AltitudeLo = *vol3.AltitudeLo
	}

	if vol3.AltitudeHi != nil {
		result.AltitudeHi = *vol3.AltitudeHi
	}

	switch t := vol3.Footprint.(type) {
	case nil:
		// Empty on purpose
	case *GeoPolygon:
		result.Footprint = t.ToProto()
	default:
		return nil, fmt.Errorf("unsupported geometry type: %T", vol3.Footprint)
	}

	return result, nil
}

// ToProto converts GeoPolygon model obj to proto
func (polygon *GeoPolygon) ToProto() *ridpb.GeoPolygon {
	if polygon == nil {
		return nil
	}

	result := &ridpb.GeoPolygon{}

	for _, pt := range polygon.Vertices {
		result.Vertices = append(result.Vertices, pt.ToProto())
	}

	return result
}

// ToProto converts latlngpoint model obj to proto
func (pt *LatLngPoint) ToProto() *ridpb.LatLngPoint {
	result := &ridpb.LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}

	return result
}
