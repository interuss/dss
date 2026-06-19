package common

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

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

// AreaFromRest parses "area" in the format 'lat0,lon0,lat1,lon1,...'
// and returns the resulting GeoPolygon, or else:
// * ErrOddNumberOfCoordinatesInAreaString
// * ErrNotEnoughPointsInPolygon
// * ErrBadCoordSet
//
// TODO(tvoss):
// * Agree and implement a maximum number of points in area
func FromGeoPolygonSring(area string) (*dssmodels.GeoPolygon, error) {
	var (
		lat, lng float64
		err      error
		counter  = 0
		scanner  = bufio.NewScanner(strings.NewReader(area))
	)
	numCoords := strings.Count(area, ",") + 1
	if numCoords%2 == 1 {
		return nil, geo.ErrOddNumberOfCoordinatesInAreaString
	}
	if numCoords/2 < 3 {
		return nil, geo.ErrNotEnoughPointsInPolygon
	}
	scanner.Split(splitAtComma)

	vertices := make([]*dssmodels.LatLngPoint, 0, numCoords/2)
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		switch counter % 2 {
		case 0:
			lat, err = strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, stacktrace.Propagate(geo.ErrBadCoordSet, "Unable to parse lat: %s", err.Error())
			}
		case 1:
			lng, err = strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, stacktrace.Propagate(geo.ErrBadCoordSet, "Unable to parse lng: %s", err.Error())
			}
			vertices = append(vertices, &dssmodels.LatLngPoint{Lat: lat, Lng: lng})
		}
		counter++
	}
	return &dssmodels.GeoPolygon{Vertices: vertices}, nil
}
