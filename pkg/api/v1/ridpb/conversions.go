// Conversions between remote ID geo objects and common model geo objects

package ridpb

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/interuss/dss/pkg/dss/models"
)

// RID -> business

func (vol4 *Volume4D) ToCommon() (*models.Volume4D, error) {
	vol3, err := vol4.GetSpatialVolume().ToCommon()
	if err != nil {
		return nil, err
	}

	result := &models.Volume4D{
		SpatialVolume: vol3,
	}

	if startTime := vol4.GetTimeStart(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return nil, err
		}
		result.StartTime = &ts
	}

	if endTime := vol4.GetTimeEnd(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return nil, err
		}
		result.EndTime = &ts
	}

	return result, nil
}

func (vol3 *Volume3D) ToCommon() (*models.Volume3D, error) {
	footprint := vol3.GetFootprint()
	if footprint == nil {
		return nil, errors.New("spatial_volume missing required footprint")
	}
	polygonFootprint := footprint.ToCommon()

	result := &models.Volume3D{
		Footprint:  polygonFootprint,
		AltitudeLo: proto.Float32(vol3.GetAltitudeLo()),
		AltitudeHi: proto.Float32(vol3.GetAltitudeHi()),
	}

	return result, nil
}

func (footprint *GeoPolygon) ToCommon() *models.GeoPolygon {
	result := &models.GeoPolygon{}

	for _, ltlng := range footprint.Vertices {
		result.Vertices = append(result.Vertices, ltlng.ToCommon())
	}

	return result
}

func (pt *LatLngPoint) ToCommon() *models.LatLngPoint {
	return &models.LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}
}

// Business -> RID

func MakeRidVolume4D(vol4 *models.Volume4D) (*Volume4D, error) {
	vol3, err := MakeRidVolume3D(vol4.SpatialVolume)
	if err != nil {
		return nil, err
	}

	result := &Volume4D{
		SpatialVolume: vol3,
	}

	if vol4.StartTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = ts
	}

	if vol4.EndTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd = ts
	}

	return result, nil
}

func MakeRidVolume3D(vol3 *models.Volume3D) (*Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	result := &Volume3D{}

	if vol3.AltitudeLo != nil {
		result.AltitudeLo = *vol3.AltitudeLo
	}

	if vol3.AltitudeHi != nil {
		result.AltitudeHi = *vol3.AltitudeHi
	}

	switch t := vol3.Footprint.(type) {
	case nil:
		// Empty on purpose
	case *models.GeoPolygon:
		result.Footprint = MakeRidGeoPolygon(t)
	default:
		return nil, fmt.Errorf("unsupported geometry type: %T", vol3.Footprint)
	}

	return result, nil
}

func MakeRidGeoPolygon(polygon *models.GeoPolygon) *GeoPolygon {
	if polygon == nil {
		return nil
	}

	result := &GeoPolygon{}

	for _, pt := range polygon.Vertices {
		result.Vertices = append(result.Vertices, MakeRidLatLngPoint(pt))
	}

	return result
}

func MakeRidLatLngPoint(pt *models.LatLngPoint) *LatLngPoint {
	result := &LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}

	return result
}
