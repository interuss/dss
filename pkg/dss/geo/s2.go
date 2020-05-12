package geo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/dss/models"
)

const (
	// DefaultMinimumCellLevel is the default minimum cell level, chosen such
	// that the minimum cell size is ~1km^2.
	DefaultMinimumCellLevel = 13
	// DefaultMaximumCellLevel is the default minimum cell level, chosen such
	// that the maximum cell size is ~1km^2.
	DefaultMaximumCellLevel = 13
	maxAllowedAreaKm2       = 2500.0
	minLat                  = -90.0
	maxLat                  = 90.0
	minLng                  = -180.0
	maxLng                  = 180.0
	radiusEarthMeter        = 6371010.0
)

var (
	// defaultRegionCoverer is the default s2.RegionCoverer for mapping areas
	// and extents to s2.CellUnion instances.
	defaultRegionCoverer = &s2.RegionCoverer{
		MinLevel: DefaultMinimumCellLevel,
		MaxLevel: DefaultMaximumCellLevel,
	}
	// RegionCoverer provides an overridable interface to defaultRegionCoverer
	RegionCoverer = defaultRegionCoverer

	errOddNumberOfCoordinatesInAreaString = errors.New("odd number of coordinates in area string")
	errNotEnoughPointsInPolygon           = errors.New("not enough points in polygon")
	errBadCoordSet                        = errors.New("coordinates did not create a well formed area")
)

// ErrAreaTooLarge is the error passed back when the requested Area is larger
// than maxAllowedAreaKm2
type ErrAreaTooLarge struct {
	msg string
}

// Error returns the error message for ErrAreaTooLarge.
func (e *ErrAreaTooLarge) Error() string {
	return e.msg
}

func splitAtComma(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, ','); i >= 0 {
		return i + 1, data[:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

// DistanceMetersToAngle converts distance in [m] to an s1.Angle in radians.
func DistanceMetersToAngle(distance float64) s1.Angle {
	return s1.Angle(distance / radiusEarthMeter)
}

// Volume4DToCellIDs converts a 4d volume to S2 cells, ignoring the time and
// altitude bounds.
func Volume4DToCellIDs(v4 *models.Volume4D) (s2.CellUnion, error) {
	if v4 == nil {
		return nil, errBadCoordSet
	}
	return Volume3DToCellIDs(v4.SpatialVolume)
}

// Volume3DToCellIDs converts a 4d volume to S2 cells, ignoring the  altitude
// bounds.
func Volume3DToCellIDs(v3 *models.Volume3D) (s2.CellUnion, error) {
	if v3 == nil {
		return nil, errBadCoordSet
	}

	switch t := v3.Footprint.(type) {
	case *models.GeoPolygon:
		return PolygonToCellIDs(t)
	case *models.GeoCircle:
		return CircleToCellIDs(t)
	default:
		return nil, errBadCoordSet
	}
}

// GeometryToCellIDs returns an s2 cell covering for geometry.
func GeometryToCellIDs(geometry models.Geometry) (s2.CellUnion, error) {
	switch t := geometry.(type) {
	case *models.GeoCircle:
		return CircleToCellIDs(t)
	case *models.GeoPolygon:
		return PolygonToCellIDs(t)
	default:
		return nil, fmt.Errorf("unsupported geometry type: %T", t)
	}
}

// CircleToCellIDs converts a geocircle to S2 cells.
func CircleToCellIDs(geocircle *models.GeoCircle) (s2.CellUnion, error) {
	if (geocircle.Center.Lat > maxLat) || (geocircle.Center.Lat < minLat) || (geocircle.Center.Lng > maxLng) || (geocircle.Center.Lng < minLng) {
		return nil, errBadCoordSet
	}

	return CoveringForLoop(s2.RegularLoop(
		s2.PointFromLatLng(s2.LatLngFromDegrees(geocircle.Center.Lat, geocircle.Center.Lng)),
		DistanceMetersToAngle(float64(geocircle.RadiusMeter)),
		20,
	))
}

// PolygonToCellIDs converts a geopolygon to S2 cells.
func PolygonToCellIDs(geopolygon *models.GeoPolygon) (s2.CellUnion, error) {
	var points []s2.Point
	if geopolygon == nil {
		return nil, errBadCoordSet
	}
	for _, ltlng := range geopolygon.Vertices {
		// ensure that coordinates passed are actually on earth
		if (ltlng.Lat > maxLat) || (ltlng.Lat < minLat) || (ltlng.Lng > maxLng) || (ltlng.Lng < minLng) {
			return nil, errBadCoordSet
		}
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(ltlng.Lat, ltlng.Lng)))
	}
	if len(points) < 3 {
		return nil, errNotEnoughPointsInPolygon
	}
	return Covering(points)
}

func loopAreaKm2(loop *s2.Loop) float64 {
	if loop.IsEmpty() {
		return 0
	}
	const earthAreaKm2 = 510072000.0 // rough area of the earth in KM².
	return (loop.Area() * earthAreaKm2) / 4.0 * math.Pi
}

// Covering calculates the S2 covering of a set of S2 points. Will try the loop
// in both clockwise and counter clockwise.
func Covering(points []s2.Point) (s2.CellUnion, error) {
	cu, err := CoveringForLoop(s2.LoopFromPoints(points))
	switch err.(type) {
	case nil:
		return cu, nil
	case *ErrAreaTooLarge:
		// Empty on purpose.
	default:
		return nil, err
	}

	// This probably happened because the vertices were not ordered counter-clockwise.
	// We can try reversing to see if that's the case.
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return CoveringForLoop(s2.LoopFromPoints(points))
}

// CoveringForLoop calculates an s2 cell covering for loop or returns an area if
// the area covered by loop is too large.
func CoveringForLoop(loop *s2.Loop) (s2.CellUnion, error) {
	if loopAreaKm2(loop) > maxAllowedAreaKm2 {
		return nil, &ErrAreaTooLarge{
			msg: fmt.Sprintf("area is too large (%fkm² > %fkm²)", loopAreaKm2(loop), maxAllowedAreaKm2),
		}
	}
	return RegionCoverer.Covering(loop), nil
}

// AreaToCellIDs parses "area" in the format 'lat0,lon0,lat1,lon1,...'
// and returns the resulting s2.CellUnion.
//
// TODO(tvoss):
//   * Agree and implement a maximum number of points in area
func AreaToCellIDs(area string) (s2.CellUnion, error) {
	var (
		lat, lng float64
		points   = []s2.Point{}
		counter  = 0
		scanner  = bufio.NewScanner(strings.NewReader(area))
	)
	numCoords := strings.Count(area, ",") + 1
	if numCoords%2 == 1 {
		return nil, errOddNumberOfCoordinatesInAreaString
	}
	if numCoords/2 < 3 {
		return nil, errNotEnoughPointsInPolygon
	}
	scanner.Split(splitAtComma)

	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		switch counter % 2 {
		case 0:
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, errBadCoordSet
			}
			lat = f
		case 1:
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, errBadCoordSet
			}
			lng = f
			points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)))
		}

		counter++
	}
	return Covering(points)
}
