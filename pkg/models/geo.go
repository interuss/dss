package models

import (
	"slices"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/dss/pkg/memstore/utils"
)

func float32p(v float32) *float32 {
	return &v
}

func timeP(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// Volume4D is a Contiguous block of geographic spacetime.
type Volume4D struct {
	// Constant spatial extent of this volume.
	SpatialVolume *Volume3D
	// End time of this volume.
	EndTime *time.Time
	// Beginning time of this volume.
	StartTime *time.Time
}

// Volume3D is A three-dimensional geographic volume consisting of a vertically-extruded shape.
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
	// The returned CellUnion must be sorted.
	CalculateCovering() (s2.CellUnion, error)
}

// UnionCellsVolumes4D unions cells volumes and returns a volume that covers all the individual volumes in space (cells) and time.
func UnionCellsVolumes4D(volumes ...*CellsVolume4D) *CellsVolume4D {
	result := &CellsVolume4D{}
	unbounded := struct{ startTime, endTime, altitudeLo, altitudeHi bool }{}
	cellUnions := make([]s2.CellUnion, 0, len(volumes))

	for _, volume := range volumes {
		cellUnions = append(cellUnions, volume.Cells)

		if volume.EndTime == nil {
			unbounded.endTime = true
			result.EndTime = nil
		} else if !unbounded.endTime {
			if result.EndTime != nil {
				if volume.EndTime.After(*result.EndTime) {
					result.EndTime = timeP(*volume.EndTime)
				}
			} else {
				result.EndTime = timeP(*volume.EndTime)
			}
		}

		if volume.StartTime == nil {
			unbounded.startTime = true
			result.StartTime = nil
		} else if !unbounded.startTime {
			if result.StartTime != nil {
				if volume.StartTime.Before(*result.StartTime) {
					result.StartTime = timeP(*volume.StartTime)
				}
			} else {
				result.StartTime = timeP(*volume.StartTime)
			}
		}

		if volume.AltitudeLo == nil {
			unbounded.altitudeLo = true
			result.AltitudeLo = nil
		} else if !unbounded.altitudeLo {
			if result.AltitudeLo != nil {
				if *volume.AltitudeLo < *result.AltitudeLo {
					result.AltitudeLo = float32p(*volume.AltitudeLo)
				}
			} else {
				result.AltitudeLo = float32p(*volume.AltitudeLo)
			}
		}

		if volume.AltitudeHi == nil {
			unbounded.altitudeHi = true
			result.AltitudeHi = nil
		} else if !unbounded.altitudeHi {
			if result.AltitudeHi != nil {
				if *volume.AltitudeHi > *result.AltitudeHi {
					result.AltitudeHi = float32p(*volume.AltitudeHi)
				}
			} else {
				result.AltitudeHi = float32p(*volume.AltitudeHi)
			}
		}
	}

	result.Cells = s2.CellUnionFromUnion(cellUnions...)
	geo.Levelify(&result.Cells)
	return result
}

// CellsVolume4D is the cells-native counterpart to Volume4D
type CellsVolume4D struct {
	Cells      s2.CellUnion
	StartTime  *time.Time
	EndTime    *time.Time
	AltitudeLo *float32
	AltitudeHi *float32
}

// Clone returns a deep copy of v
func (v *CellsVolume4D) Clone() *CellsVolume4D {
	if v == nil {
		return nil
	}
	return &CellsVolume4D{
		Cells:      slices.Clone(v.Cells),
		StartTime:  utils.ClonePtr(v.StartTime),
		EndTime:    utils.ClonePtr(v.EndTime),
		AltitudeLo: utils.ClonePtr(v.AltitudeLo),
		AltitudeHi: utils.ClonePtr(v.AltitudeHi),
	}
}

// GeoCircle models a circular enclosed area on earth's surface.
type GeoCircle struct {
	Center      LatLngPoint
	RadiusMeter float32
}

// CalculateCovering returns the (sorted) spatial covering of gc.
func (gc *GeoCircle) CalculateCovering() (s2.CellUnion, error) {
	if err := geo.ValidateLatLng(gc.Center.Lat, gc.Center.Lng); err != nil {
		return nil, err
	}

	if !(gc.RadiusMeter > 0) {
		return nil, geo.ErrRadiusMustBeLargerThan0
	}

	// TODO: Use an S2 Cap as an inscribed polygon does not fully cover the defined circle
	return geo.RegionCoverer.Covering(s2.RegularLoop(
		s2.PointFromLatLng(s2.LatLngFromDegrees(gc.Center.Lat, gc.Center.Lng)),
		geo.DistanceMetersToAngle(float64(gc.RadiusMeter)),
		20,
	)), nil
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

// CalculateCovering returns the (sorted) spatial covering of gp.
func (gp *GeoPolygon) CalculateCovering() (s2.CellUnion, error) {
	var points []s2.Point
	if gp == nil {
		return nil, geo.ErrBadCoordSet
	}
	for _, v := range gp.Vertices {
		if err := geo.ValidateLatLng(v.Lat, v.Lng); err != nil {
			return nil, err
		}
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(v.Lat, v.Lng)))
	}
	if len(points) < 3 {
		return nil, geo.ErrNotEnoughPointsInPolygon
	}
	return geo.Covering(points)
}

// LatLngPoint models a point on the earth's surface.
type LatLngPoint struct {
	Lat float64
	Lng float64
}
