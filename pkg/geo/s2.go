package geo

import (
	"bufio"
	"bytes"
	"math"
	"strconv"
	"strings"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/palantir/stacktrace"
)

const (
	// DefaultMinimumCellLevel is the default minimum cell level, chosen such
	// that the minimum cell size is ~1km^2.
	DefaultMinimumCellLevel = 13
	// DefaultMaximumCellLevel is the default minimum cell level, chosen such
	// that the maximum cell size is ~1km^2.
	DefaultMaximumCellLevel = 13
	maxAllowedAreaKm2       = 2500.0
	radiusEarthMeter        = 6371010.0

	earthAreaKm2 = 510072000.0 // rough area of the earth in KM².
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
)

// Levelify takes a cell union that might have been normalized and returns to
// the appropriate level
func Levelify(cells *s2.CellUnion) {
	// thirty is the number of s2 cells, we make it negative to get the number
	// of cells we want
	cells.Denormalize(DefaultMinimumCellLevel, 1)
}

func ValidateCell(cell s2.CellID) error {
	if cell.Level() < DefaultMinimumCellLevel || cell.Level() > DefaultMaximumCellLevel {
		return stacktrace.NewError("Cells must be at level 13 at current implementation")
	}
	return nil
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

// SetCells is a convenience function that accepts an int64 array and converts
// to s2.CellUnion.
func CellUnionFromInt64(cellIds []int64) s2.CellUnion {
	cells := s2.CellUnion{}
	for _, id := range cellIds {
		cells = append(cells, s2.CellID(id))
	}
	return cells
}

// DistanceMetersToAngle converts distance in [m] to an s1.Angle in radians.
func DistanceMetersToAngle(distance float64) s1.Angle {
	return s1.Angle(distance / radiusEarthMeter)
}

func loopAreaKm2(loop *s2.Loop) float64 {
	if loop.IsEmpty() {
		return 0
	}
	return (loop.Area() * earthAreaKm2) / (4.0 * math.Pi)
}

// Covering calculates the S2 covering of a set of S2 points. Will try the loop
// in both clockwise and counter clockwise.
func Covering(points []s2.Point) (s2.CellUnion, error) {
	loop := s2.LoopFromPoints(points)
	area1 := loopAreaKm2(loop)
	if area1 > maxAllowedAreaKm2 {
		// This may have happened because the vertices were not ordered counter-clockwise.
		// We can try reversing to see if that's the case.
		for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
			points[i], points[j] = points[j], points[i]
		}
		loop = s2.LoopFromPoints(points)
	}
	area2 := loopAreaKm2(loop)
	if area2 > maxAllowedAreaKm2 {
		return nil, stacktrace.Propagate(
			ErrAreaTooLarge, "Area is too large (%fkm² > %fkm²)",
			math.Min(area1, area2), maxAllowedAreaKm2)
	}
	if area2 <= 0 {
		// Since the loop has no area, try a PolyLine
		pl := s2.Polyline(loop.Vertices())
		return RegionCoverer.Covering(&pl), nil
	}
	return RegionCoverer.Covering(loop), nil
}

// AreaToCellIDs parses "area" in the format 'lat0,lon0,lat1,lon1,...'
// and returns the resulting s2.CellUnion, or else:
// * ErrOddNumberOfCoordinatesInAreaString
// * ErrNotEnoughPointsInPolygon
// * ErrBadCoordSet
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
		return nil, ErrOddNumberOfCoordinatesInAreaString
	}
	if numCoords/2 < 3 {
		return nil, ErrNotEnoughPointsInPolygon
	}
	scanner.Split(splitAtComma)

	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		switch counter % 2 {
		case 0:
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, stacktrace.Propagate(ErrBadCoordSet, "Unable to parse lat: %s", err.Error())
			}
			lat = f
		case 1:
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, stacktrace.Propagate(ErrBadCoordSet, "Unable to parse lng: %s", err.Error())
			}
			lng = f
			points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)))
		}

		counter++
	}
	return Covering(points)
}
