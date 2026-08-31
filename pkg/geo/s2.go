package geo

import (
	"math"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/interuss/stacktrace"
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

// validateLoop returns an error if any of the edges formed by the specified
// points intersect each other.  There is an edge between the last and first
// vertices.
func validateLoop(points []s2.Point) error {
	n := len(points)
	for i := 0; i < n-2; i++ {
		upperBound := n
		if i == 0 {
			upperBound = n - 1
		}
		for j := i + 2; j < upperBound; j++ {
			if s2.CrossingSign(points[i], points[i+1], points[j], points[(j+1)%n]) != s2.DoNotCross {
				return stacktrace.NewError("Intersection found between polygon edge %d and %d", i, j)
			}
		}
	}
	return nil
}

// Covering calculates the S2 covering of a set of S2 points representing a
// polygon. Will try the loop in both clockwise and counter clockwise.
func Covering(points []s2.Point) (s2.CellUnion, error) {
	err := validateLoop(points)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error validating polygon")
	}
	loop := s2.LoopFromPoints(points)
	loop.Normalize()
	err = loop.Validate()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error validating loop")
	}
	area := loopAreaKm2(loop)
	if area > maxAllowedAreaKm2 {
		return nil, stacktrace.Propagate(
			ErrAreaTooLarge, "Area is too large (%fkm² > %fkm²)",
			area, maxAllowedAreaKm2)
	}
	if area <= 0 {
		// Since the loop has no area, try a PolyLine
		pl := s2.Polyline(loop.Vertices())
		return RegionCoverer.Covering(&pl), nil
	}
	return RegionCoverer.Covering(loop), nil
}
