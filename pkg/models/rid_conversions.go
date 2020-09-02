package models

import (
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/proto"

	"github.com/interuss/dss/pkg/api/v1/ridpb"
	"github.com/interuss/stacktrace"
)

// === RID -> Business ===

// Volume4DFromRIDProto converts proto to model object
func Volume4DFromRIDProto(vol4 *ridpb.Volume4D) (*Volume4D, error) {
	vol3, err := Volume3DFromRIDProto(vol4.GetSpatialVolume())
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	result := &Volume4D{
		SpatialVolume: vol3,
	}

	if startTime := vol4.GetTimeStart(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time from proto")
		}
		result.StartTime = &ts
	}

	if endTime := vol4.GetTimeEnd(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time from proto")
		}
		result.EndTime = &ts
	}

	return result, nil
}

// Volume3DFromRIDProto convert proto to model object
func Volume3DFromRIDProto(vol3 *ridpb.Volume3D) (*Volume3D, error) {
	footprint := vol3.GetFootprint()
	if footprint == nil {
		return nil, stacktrace.NewError("spatial_volume missing required footprint")
	}
	polygonFootprint := GeoPolygonFromRIDProto(footprint)

	result := &Volume3D{
		Footprint:  polygonFootprint,
		AltitudeLo: proto.Float32(vol3.GetAltitudeLo()),
		AltitudeHi: proto.Float32(vol3.GetAltitudeHi()),
	}

	return result, nil
}

// GeoPolygonFromRIDProto convert proto to model object
func GeoPolygonFromRIDProto(footprint *ridpb.GeoPolygon) *GeoPolygon {
	result := &GeoPolygon{}

	for _, ltlng := range footprint.Vertices {
		result.Vertices = append(result.Vertices, PointFromRIDProto(ltlng))
	}

	return result
}

// PointFromRIDProto convert proto to model object
func PointFromRIDProto(pt *ridpb.LatLngPoint) *LatLngPoint {
	return &LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}
}

// === Business -> RID ===

// ToRIDProto converts Volume4D model obj to proto
func (vol4 *Volume4D) ToRIDProto() (*ridpb.Volume4D, error) {
	vol3, err := vol4.SpatialVolume.ToRIDProto()
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	result := &ridpb.Volume4D{
		SpatialVolume: vol3,
	}

	if vol4.StartTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.StartTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time from proto")
		}
		result.TimeStart = ts
	}

	if vol4.EndTime != nil {
		ts, err := ptypes.TimestampProto(*vol4.EndTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time from proto")
		}
		result.TimeEnd = ts
	}

	return result, nil
}

// ToRIDProto converts Volume3D model obj to proto
func (vol3 *Volume3D) ToRIDProto() (*ridpb.Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	result := &ridpb.Volume3D{}

	if vol3.AltitudeLo != nil {
		result.AltitudeLo = *vol3.AltitudeLo
	}

	if vol3.AltitudeHi != nil {
		result.AltitudeHi = *vol3.AltitudeHi
	}

	switch t := vol3.Footprint.(type) {
	case nil:
		// Empty on purpose
	case *GeoPolygon:
		result.Footprint = t.ToRIDProto()
	default:
		return nil, stacktrace.NewError("Unsupported geometry type: %T", vol3.Footprint)
	}

	return result, nil
}

// ToRIDProto converts GeoPolygon model obj to proto
func (gp *GeoPolygon) ToRIDProto() *ridpb.GeoPolygon {
	if gp == nil {
		return nil
	}

	result := &ridpb.GeoPolygon{}

	for _, pt := range gp.Vertices {
		result.Vertices = append(result.Vertices, pt.ToRIDProto())
	}

	return result
}

// ToRIDProto converts latlngpoint model obj to proto
func (pt *LatLngPoint) ToRIDProto() *ridpb.LatLngPoint {
	result := &ridpb.LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}

	return result
}
