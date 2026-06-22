package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/stacktrace"
)

const (
	// TimeFormatRFC3339 is the string used for RFC3339
	TimeFormatRFC3339 = "RFC3339"
	minLat            = -90.0
	maxLat            = 90.0
	minLng            = -180.0
	maxLng            = 180.0
	UnitsM            = "M"
	ReferenceW84      = "W84"
)

var (
	unitToMeterMultiplicativeFactors = map[unit]float32{
		unitMeter: 1,
	}

	altitudeReferenceWGS84 altitudeReference = "W84"
	unitMeter              unit              = "M"
)

type (
	altitudeReference string
	unit              string
)

func (ar altitudeReference) String() string {
	return string(ar)
}

func (u unit) String() string {
	return string(u)
}

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

type Volume3DJSON struct {
	AltitudeHi *float32      `json:"AltitudeHi,omitempty"`
	AltitudeLo *float32      `json:"AltitudeLo,omitempty"`
	Footprint  *geometryJSON `json:"Footprint,omitempty"`
}

type geometryType string

const (
	circle  geometryType = "circle"
	polygon geometryType = "polygon"
	cells   geometryType = "cells"
)

// geometryJSON is a helper struct for marshaling and unmarshaling Geometry types to/from JSON.
type geometryJSON struct {
	Type    geometryType `json:"type"`
	Polygon *GeoPolygon  `json:"polygon,omitempty"`
	Circle  *GeoCircle   `json:"circle,omitempty"`
	Cells   []s2.CellID  `json:"cells,omitempty"`
}

func (v Volume3D) MarshalJSON() ([]byte, error) {
	w := Volume3DJSON{AltitudeHi: v.AltitudeHi, AltitudeLo: v.AltitudeLo}
	if v.Footprint != nil {
		switch f := v.Footprint.(type) {
		case *GeoPolygon:
			w.Footprint = &geometryJSON{Type: polygon, Polygon: f}

		case *GeoCircle:
			w.Footprint = &geometryJSON{Type: circle, Circle: f}

		case precomputedCellGeometry:
			cellsResult := make([]s2.CellID, 0, len(f))
			for id := range f {
				cellsResult = append(cellsResult, id)
			}
			w.Footprint = &geometryJSON{Type: cells, Cells: cellsResult}

		default:
			return nil, fmt.Errorf("Volume3D: unsupported Footprint type %T for JSON marshaling", v.Footprint)
		}
	}

	return json.Marshal(w)
}

func (v *Volume3D) UnmarshalJSON(data []byte) error {
	var w Volume3DJSON
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	v.AltitudeHi = w.AltitudeHi
	v.AltitudeLo = w.AltitudeLo
	if w.Footprint != nil {
		switch w.Footprint.Type {
		case polygon:
			v.Footprint = w.Footprint.Polygon

		case circle:
			v.Footprint = w.Footprint.Circle

		case cells:
			pcg := make(precomputedCellGeometry, len(w.Footprint.Cells))
			for _, id := range w.Footprint.Cells {
				pcg[id] = struct{}{}
			}

			v.Footprint = pcg
		default:
			return fmt.Errorf("Volume3D: unknown geometry type %q", w.Footprint.Type)
		}
	}

	return nil
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
// individual volumes in space and time, or one of these root causes:
// * geo.ErrMissingFootprint
// * geo.ErrNotEnoughPointsInPolygon
// * geo.ErrBadCoordSet
// * geo.ErrRadiusMustBeLargerThan0
func UnionVolumes4D(volumes ...*Volume4D) (*Volume4D, error) {
	result := &Volume4D{}
	unbounded := struct{ startTime, endTime, altitudeLo, altitudeHi bool }{}

	for _, volume := range volumes {
		if volume.EndTime == nil {
			unbounded.endTime = true
			result.EndTime = nil
		} else if !unbounded.endTime {
			if result.EndTime != nil {
				if volume.EndTime.After(*result.EndTime) {
					*result.EndTime = *volume.EndTime
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

			if volume.SpatialVolume.AltitudeLo == nil {
				unbounded.altitudeLo = true
				result.SpatialVolume.AltitudeLo = nil
			} else if !unbounded.altitudeLo {
				if result.SpatialVolume.AltitudeLo != nil {
					if *volume.SpatialVolume.AltitudeLo < *result.SpatialVolume.AltitudeLo {
						*result.SpatialVolume.AltitudeLo = *volume.SpatialVolume.AltitudeLo
					}
				} else {
					result.SpatialVolume.AltitudeLo = float32p(*volume.SpatialVolume.AltitudeLo)
				}
			}

			if volume.SpatialVolume.AltitudeHi == nil {
				unbounded.altitudeHi = true
				result.SpatialVolume.AltitudeHi = nil
			} else if !unbounded.altitudeHi {
				if result.SpatialVolume.AltitudeHi != nil {
					if *volume.SpatialVolume.AltitudeHi > *result.SpatialVolume.AltitudeHi {
						*result.SpatialVolume.AltitudeHi = *volume.SpatialVolume.AltitudeHi
					}
				} else {
					result.SpatialVolume.AltitudeHi = float32p(*volume.SpatialVolume.AltitudeHi)
				}
			}

			if volume.SpatialVolume.Footprint != nil {
				cells, err := volume.SpatialVolume.Footprint.CalculateCovering()
				if err != nil {
					return nil, stacktrace.Propagate(err, "Error calculating footprint covering")
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

// CalculateSpatialCovering returns the spatial covering of vol4, or one of:
// * geo.ErrMissingSpatialVolume
// * geo.ErrMissingFootprint
// * geo.ErrNotEnoughPointsInPolygon
// * geo.ErrBadCoordSet
// * geo.ErrRadiusMustBeLargerThan0
func (vol4 *Volume4D) CalculateSpatialCovering() (s2.CellUnion, error) {
	if vol4.SpatialVolume == nil {
		return nil, geo.ErrMissingSpatialVolume
	}
	return vol4.SpatialVolume.CalculateCovering()
}

// CalculateCovering returns the spatial covering of vol3, or one of:
// * geo.ErrMissingFootprint
// * geo.ErrNotEnoughPointsInPolygon
// * geo.ErrBadCoordSet
// * geo.ErrRadiusMustBeLargerThan0
func (vol3 *Volume3D) CalculateCovering() (s2.CellUnion, error) {
	if vol3.Footprint == nil {
		return nil, geo.ErrMissingFootprint
	}
	return vol3.Footprint.CalculateCovering()
}

// CalculateCovering returns the result of invoking gf, with possible errors:
// * geo.ErrNotEnoughPointsInPolygon
// * geo.ErrBadCoordSet
// * geo.ErrRadiusMustBeLargerThan0
func (gf GeometryFunc) CalculateCovering() (s2.CellUnion, error) {
	return gf()
}

// GeoCircle models a circular enclosed area on earth's surface.
type GeoCircle struct {
	Center      LatLngPoint
	RadiusMeter float32
}

// CalculateCovering returns the spatial covering of gc.
func (gc *GeoCircle) CalculateCovering() (s2.CellUnion, error) {
	if (gc.Center.Lat > maxLat) || (gc.Center.Lat < minLat) || (gc.Center.Lng > maxLng) || (gc.Center.Lng < minLng) {
		return nil, geo.ErrBadCoordSet
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

// CalculateCovering returns the spatial covering of gp.
func (gp *GeoPolygon) CalculateCovering() (s2.CellUnion, error) {
	var points []s2.Point
	if gp == nil {
		return nil, geo.ErrBadCoordSet
	}
	for _, v := range gp.Vertices {
		// ensure that coordinates passed are actually on earth
		if (v.Lat > maxLat) || (v.Lat < minLat) || (v.Lng > maxLng) || (v.Lng < minLng) {
			return nil, geo.ErrBadCoordSet
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
