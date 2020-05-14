package models

import (
	"errors"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/dss/geo"
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
