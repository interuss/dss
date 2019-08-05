package geo

import (
	"bufio"
	"bytes"
	"errors"
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

	// WindingOrderCCW describes the counter-clockwise winding order.
	WindingOrderCCW WindingOrder = 0
	// WindingOrderCW describes the clockwise winding order.
	WindingOrderCW WindingOrder = 1
)

var (
	// DefaultRegionCoverer is the default s2.RegionCoverer for mapping areas
	// and extents to s2.CellUnion instances.
	DefaultRegionCoverer = &s2.RegionCoverer{
		MinLevel: DefaultMinimumCellLevel,
		MaxLevel: DefaultMaximumCellLevel,
	}

	errOddNumberOfCoordinatesInAreaString = errors.New("odd number of coordinates in area string")
	errNotEnoughPointsInPolygon           = errors.New("not enough points in polygon")
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

func Volume4DToCellIDs(v4 *dspb.Volume4D, rc *s2.RegionCoverer) s2.CellUnion {
	return Volume3DToCellIDs(v4.SpatialVolume, rc)
}

func Volume3DToCellIDs(v3 *dspb.Volume3D, rc *s2.RegionCoverer) s2.CellUnion {
	return GeoPolygonToCellIDs(v3.Footprint, rc)
}

func GeoPolygonToCellIDs(geopolygon *dspb.GeoPolygon, rc *s2.RegionCoverer) s2.CellUnion {
	var points []s2.Point
	for _, ltlng := range geopolygon.Vertices {
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(ltlng.Lat, ltlng.Lng)))
	}
	loop := s2.LoopFromPoints(points)

	return rc.Covering(loop)
}

// AreaToCellIDs parses "area" in the format 'lat0,lon0,lat1,lon1,...'
// and returns the resulting s2.CellUnion.
//
// TODO(tvoss):
//   * Agree and implement a maximum number of points in area
func AreaToCellIDs(area string, winding WindingOrder, rc *s2.RegionCoverer) (s2.CellUnion, error) {
	var (
		lat, lng = float64(0), float64(0)
		points   = []s2.Point{}
		counter  = 0
		scanner  = bufio.NewScanner(strings.NewReader(area))
	)
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

			switch winding {
			case WindingOrderCCW:
				points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)))
			case WindingOrderCW:
				points = append([]s2.Point{s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))}, points...)
			}
		}

		counter++
	}

	switch {
	case counter%2 != 0:
		return nil, errOddNumberOfCoordinatesInAreaString
	case len(points) < 3:
		return nil, errNotEnoughPointsInPolygon
	}

	return rc.Covering(s2.LoopFromPoints(points)), nil
}
