package geo

import (
	"github.com/golang/geo/s2"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
)

var staticCellLevel = 13

func Volume4DToCellIDs(v4 *dspb.Volume4D) []uint64 {
	return Volume3DToCellIDs(v4.SpatialVolume)
}

func Volume3DToCellIDs(v3 *dspb.Volume3D) []uint64 {
	return GeoPolygonToCellIDs(v3.Footprint)
}

func GeoPolygonToCellIDs(geopolygon *dspb.GeoPolygon) []uint64 {
	var points []s2.Point
	for _, ltlng := range geopolygon.Vertices {
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(ltlng.Lat, ltlng.Lng)))
	}
	loop := s2.LoopFromPoints(points)

	rc := RegionCoverer()
	cellUnion := rc.Covering(loop)
	ids := make([]uint64, len(cellUnion))
	for i, cellId := range cellUnion {
		ids[i] = uint64(cellId)
	}
	return ids
}

func RegionCoverer() *s2.RegionCoverer {
	return &s2.RegionCoverer{MaxLevel: staticCellLevel, MinLevel: staticCellLevel}
}
