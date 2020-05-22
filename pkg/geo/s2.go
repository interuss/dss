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

// Levelify takes a cell union that might have been normalized and returns to
// the appropriate level
func Levelify(cells *s2.CellUnion) {
	// thirty is the number of s2 cells, we make it negative to get the number
	// of cells we want
	cells.Denormalize(DefaultMinimumCellLevel, 1)
}

func ValidateCell(cell s2.CellID) error {
	if cell.Level() < DefaultMinimumCellLevel || cell.Level() > DefaultMaximumCellLevel {
		return errors.New("cells must be at level 13 at current implementation")
	}
	return nil
}

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
		// Area may be too large because vertices were wound in the opposite direction; check below
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
