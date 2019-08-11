package geo

import (
	"bufio"
	"bytes"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/golang/geo/s2"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
)

const (
	// DefaultMinimumCellLevel is the default minimum cell level, chosen such
	// that the minimum cell size is ~1km^2.
	DefaultMinimumCellLevel int = 13
	// DefaultMaximumCellLevel is the default minimum cell level, chosen such
	// that the maximum cell size is ~1km^2.
	DefaultMaximumCellLevel int = 13
	maxAllowedSqMi              = 1000
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
	errBadCoordSet                        = errors.New("area is too large")
	maxArea                               = maxLoopArea()
)

// WindingOrder describes the winding order for enumerating
// vertices of an area.
type WindingOrder int

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

func Volume4DToCellIDs(v4 *dspb.Volume4D) s2.CellUnion {
	return Volume3DToCellIDs(v4.SpatialVolume)
}

func Volume3DToCellIDs(v3 *dspb.Volume3D) s2.CellUnion {
	return GeoPolygonToCellIDs(v3.Footprint)
}

func GeoPolygonToCellIDs(geopolygon *dspb.GeoPolygon) s2.CellUnion {
	var points []s2.Point
	for _, ltlng := range geopolygon.Vertices {
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(ltlng.Lat, ltlng.Lng)))
	}
	loop := s2.LoopFromPoints(points)

	return RegionCoverer.Covering(loop)
}

func maxLoopArea() float {
	var (
		sqMiEarth     = 197000000 // rought square miles of earth.
		scalingFactor = sqMiEarth / 4 * math.Pi
	)
	return maxAllowedSqMi / scalingFactor
}

// AreaToCellIDs parses "area" in the format 'lat0,lon0,lat1,lon1,...'
// and returns the resulting s2.CellUnion.
//
// TODO(tvoss):
//   * Agree and implement a maximum number of points in area
func AreaToCellIDs(area string) (s2.CellUnion, error) {
	var (
		lat, lng = float64(0), float64(0)
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
				return nil, err
			}
			lat = f
		case 1:
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, err
			}
			lng = f
			points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)))
		}

		counter++
	}
	loop := s2.LoopFromPoints(points)
	// TODO(steeling): consider setting max number of vertices.
	area := loop.Area()
	if area == 0 {
		return nil, errBadCoordSet
	}
	if area > maxArea() {
		return nil, errTooLarge
	}
	return RegionCoverer.Covering(loop), nil
}
