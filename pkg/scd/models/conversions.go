package models

import (
	"time"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

const (
	// TimeFormatRFC3339 is the string used for RFC3339
	TimeFormatRFC3339 = "RFC3339"
	UnitsM            = "M"
	ReferenceW84      = "W84"
)

// TODO: remove along with UnionVolumes4DFromSCDRest; CellsVolume4DValidator is its cells-native
// replacement.
type Volume4DValidator func(*dssmodels.Volume4D) error

// TODO: remove along with UnionVolumes4DFromSCDRest; WithRequireCellsTimeBounds is its cells-native
// replacement.
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

// TODO: remove along with UnionVolumes4DFromSCDRest; WithRequireCellsEndTimeAfter is its cells-native
// replacement.
func WithRequireEndTimeAfter(now time.Time) Volume4DValidator {
	return func(v *dssmodels.Volume4D) error {
		if v.EndTime != nil && v.EndTime.Before(now) {
			return stacktrace.NewError("End time may not be in the past")
		}
		return nil
	}
}

// TODO: remove along with UnionVolumes4DFromSCDRest, once operational intents (its only remaining
// caller) migrates to the cells-native path.
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
// TODO: remove once callers (constraints, operational intents) use CellsVolume4DFromSCDRest
func UnionVolumes4DFromSCDRest(vol4s []restapi.Volume4D, validators ...Volume4DValidator) (*dssmodels.Volume4D, error) {
	volumes := make([]*dssmodels.Volume4D, len(vol4s))
	for idx, vol4 := range vol4s {
		volume, err := volume4DFromSCDRest(&vol4)
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

func CellsVolume4DFromSCDRest(vol4 *restapi.Volume4D) (*dssmodels.CellsVolume4D, error) {
	filter := &dssmodels.CellsVolume4D{}

	if vol4.TimeStart != nil {
		ts, err := time.Parse(time.RFC3339Nano, vol4.TimeStart.Value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time")
		}
		filter.StartTime = &ts
	}

	if vol4.TimeEnd != nil {
		ts, err := time.Parse(time.RFC3339Nano, vol4.TimeEnd.Value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time")
		}
		filter.EndTime = &ts
	}

	if filter.StartTime != nil && filter.EndTime != nil && filter.StartTime.After(*filter.EndTime) {
		return nil, stacktrace.NewError("Start time cannot be after end time")
	}

	altLo, altHi, err := altitudeBoundsFromSCDRest(&vol4.Volume)
	if err != nil {
		return nil, err
	}
	filter.AltitudeLo = altLo
	filter.AltitudeHi = altHi

	cells, err := cellsFromOutlineSCDRest(&vol4.Volume)
	if err != nil {
		return filter, err
	}
	filter.Cells = cells

	return filter, nil
}

func altitudeBoundsFromSCDRest(vol3 *restapi.Volume3D) (*float32, *float32, error) {
	var altLo *float32
	if vol3.AltitudeLower != nil {
		if vol3.AltitudeLower.Units != UnitsM {
			return nil, nil, stacktrace.NewError("Invalid lower altitude unit")
		}
		if vol3.AltitudeLower.Reference != ReferenceW84 {
			return nil, nil, stacktrace.NewError("Invalid lower altitude reference")
		}
		altLo = new(float32(vol3.AltitudeLower.Value))
	}

	var altHi *float32
	if vol3.AltitudeUpper != nil {
		if vol3.AltitudeUpper.Units != UnitsM {
			return nil, nil, stacktrace.NewError("Invalid upper altitude unit")
		}
		if vol3.AltitudeUpper.Reference != ReferenceW84 {
			return nil, nil, stacktrace.NewError("Invalid upper altitude reference")
		}
		altHi = new(float32(vol3.AltitudeUpper.Value))
	}

	if altLo != nil && altHi != nil && *altLo > *altHi {
		return nil, nil, stacktrace.NewError("Lower altitude cannot be greater than upper altitude")
	}

	return altLo, altHi, nil
}

func cellsFromOutlineSCDRest(vol3 *restapi.Volume3D) (s2.CellUnion, error) {
	switch {
	case vol3.OutlineCircle != nil && vol3.OutlinePolygon != nil:
		return nil, stacktrace.NewError("Both circle and polygon specified in outline geometry")
	case vol3.OutlinePolygon != nil:
		return polygonCellsFromSCDRest(vol3.OutlinePolygon)
	case vol3.OutlineCircle != nil:
		return circleCellsFromSCDRest(vol3.OutlineCircle)
	default:
		return nil, geo.ErrMissingFootprint
	}
}

func circleCellsFromSCDRest(c *restapi.Circle) (s2.CellUnion, error) {
	lat := float64(c.Center.Lat)
	lng := float64(c.Center.Lng)
	if err := geo.ValidateLatLng(lat, lng); err != nil {
		return nil, err
	}
	if !(c.Radius.Value > 0) {
		return nil, geo.ErrRadiusMustBeLargerThan0
	}

	return geo.RegionCoverer.Covering(s2.RegularLoop(
		s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)),
		geo.DistanceMetersToAngle(float64(c.Radius.Value)),
		20,
	)), nil
}

func polygonCellsFromSCDRest(p *restapi.Polygon) (s2.CellUnion, error) {
	var points []s2.Point
	for _, v := range p.Vertices {
		lat := float64(v.Lat)
		lng := float64(v.Lng)
		if err := geo.ValidateLatLng(lat, lng); err != nil {
			return nil, err
		}
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)))
	}
	if len(points) < 3 {
		return nil, geo.ErrNotEnoughPointsInPolygon
	}
	return geo.Covering(points)
}

type CellsVolume4DValidator func(*dssmodels.CellsVolume4D) error

func WithRequireCellsTimeBounds() CellsVolume4DValidator {
	return func(v *dssmodels.CellsVolume4D) error {
		if v.StartTime == nil {
			return stacktrace.NewError("Missing start time")
		}
		if v.EndTime == nil {
			return stacktrace.NewError("Missing end time")
		}
		return nil
	}
}

func WithRequireCellsEndTimeAfter(now time.Time) CellsVolume4DValidator {
	return func(v *dssmodels.CellsVolume4D) error {
		if v.EndTime != nil && v.EndTime.Before(now) {
			return stacktrace.NewError("End time may not be in the past")
		}
		return nil
	}
}

func UnionCellsVolume4DFromSCDRest(vol4s []restapi.Volume4D, validators ...CellsVolume4DValidator) (*dssmodels.CellsVolume4D, error) {
	cellsVolumes := make([]*dssmodels.CellsVolume4D, len(vol4s))
	for idx, vol4 := range vol4s {
		cellsVolume, err := CellsVolume4DFromSCDRest(&vol4)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to parse volume %d", idx)
		}
		cellsVolumes[idx] = cellsVolume
	}
	union := dssmodels.UnionCellsVolumes4D(cellsVolumes...)

	for _, validator := range validators {
		if err := validator(union); err != nil {
			return nil, stacktrace.Propagate(err, "Invalid volume union")
		}
	}

	return union, nil
}

// TODO: remove along with UnionVolumes4DFromSCDRest.
func volume4DFromSCDRest(vol4 *restapi.Volume4D) (*dssmodels.Volume4D, error) {
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
// TODO: remove along with UnionVolumes4DFromSCDRest.
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

	switch {
	case vol3.OutlineCircle != nil && vol3.OutlinePolygon != nil:
		return nil, stacktrace.NewError("Both circle and polygon specified in outline geometry")
	case vol3.OutlinePolygon != nil:
		return &dssmodels.Volume3D{
			Footprint:  GeoPolygonFromSCDRest(vol3.OutlinePolygon),
			AltitudeLo: altLo,
			AltitudeHi: altHi,
		}, nil
	case vol3.OutlineCircle != nil:
		return &dssmodels.Volume3D{
			Footprint:  GeoCircleFromSCDRest(vol3.OutlineCircle),
			AltitudeLo: altLo,
			AltitudeHi: altHi,
		}, nil
	}

	return &dssmodels.Volume3D{
		AltitudeLo: altLo,
		AltitudeHi: altHi,
	}, nil
}

// GeoCircleFromSCDRest converts a circle SCD v1 REST model to a GeoCircle
// TODO: remove along with UnionVolumes4DFromSCDRest; circleCellsFromSCDRest is its cells-native
// replacement.
func GeoCircleFromSCDRest(c *restapi.Circle) *dssmodels.GeoCircle {
	return &dssmodels.GeoCircle{
		Center:      *LatLngPointFromSCDRest(c.Center),
		RadiusMeter: c.Radius.Value,
	}
}

// GeoPolygonFromSCDRest converts a polygon SCD v1 REST model to a GeoPolygon
// TODO: remove along with UnionVolumes4DFromSCDRest; polygonCellsFromSCDRest is its cells-native
// replacement.
func GeoPolygonFromSCDRest(p *restapi.Polygon) *dssmodels.GeoPolygon {
	result := &dssmodels.GeoPolygon{}
	for _, ltlng := range p.Vertices {
		result.Vertices = append(result.Vertices, LatLngPointFromSCDRest(&ltlng))
	}

	return result
}

// LatLngPointFromSCDRest converts a point SCD v1 REST model to a latlngpoint
// TODO: remove along with UnionVolumes4DFromSCDRest; only GeoCircleFromSCDRest/GeoPolygonFromSCDRest
// call it.
func LatLngPointFromSCDRest(p *restapi.LatLngPoint) *dssmodels.LatLngPoint {
	return &dssmodels.LatLngPoint{
		Lat: float64(p.Lat),
		Lng: float64(p.Lng),
	}
}
