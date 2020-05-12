package models

import (
	"time"
)

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
	isGeometry()
}

// GeoCircle models a circular enclosed area on earth's surface.
type GeoCircle struct {
	Center      LatLngPoint
	RadiusMeter float32
}

func (*GeoCircle) isGeometry() {}

// GeoPolygon models an enclosed area on the earth.
// The bounding edges of this polygon shall be the shortest paths between connected vertices.  This means, for instance, that the edge between two points both defined at a particular latitude is not generally contained at that latitude.
// The winding order shall be interpreted as the order which produces the smaller area.
// The path between two vertices shall be the shortest possible path between those vertices.
// Edges may not cross.
// Vertices may not be duplicated.  In particular, the final polygon vertex shall not be identical to the first vertex.
type GeoPolygon struct {
	Vertices []*LatLngPoint
}

func (*GeoPolygon) isGeometry() {}

// LatLngPoint models a point on the earth's surface.
type LatLngPoint struct {
	Lat float64
	Lng float64
}
