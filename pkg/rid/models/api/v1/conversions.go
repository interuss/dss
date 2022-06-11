package apiv1

import (
	ridpb "github.com/interuss/dss/pkg/api/v1/ridpbv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/proto"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// === RID -> Business ===

// FromVolume4D converts proto to business object
func FromVolume4D(vol4 *ridpb.Volume4D) (*dssmodels.Volume4D, error) {
	vol3, err := FromVolume3D(vol4.GetSpatialVolume())
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	result := &dssmodels.Volume4D{
		SpatialVolume: vol3,
	}

	if startTime := vol4.GetTimeStart(); startTime != nil {
		err := startTime.CheckValid()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time from proto")
		}
		ts := startTime.AsTime()
		result.StartTime = &ts
	}

	if endTime := vol4.GetTimeEnd(); endTime != nil {
		err := endTime.CheckValid()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time from proto")
		}
		ts := endTime.AsTime()
		result.EndTime = &ts
	}

	return result, nil
}

// FromVolume3D converts proto to business object
func FromVolume3D(vol3 *ridpb.Volume3D) (*dssmodels.Volume3D, error) {
	footprint := vol3.GetFootprint()
	if footprint == nil {
		return nil, stacktrace.NewError("spatial_volume missing required footprint")
	}
	polygonFootprint := FromGeoPolygon(footprint)

	result := &dssmodels.Volume3D{
		Footprint:  polygonFootprint,
		AltitudeLo: proto.Float32(vol3.GetAltitudeLo()),
		AltitudeHi: proto.Float32(vol3.GetAltitudeHi()),
	}

	return result, nil
}

// FromGeoPolygon converts proto to business object
func FromGeoPolygon(footprint *ridpb.GeoPolygon) *dssmodels.GeoPolygon {
	result := &dssmodels.GeoPolygon{}

	for _, ltlng := range footprint.Vertices {
		result.Vertices = append(result.Vertices, FromLatLngPoint(ltlng))
	}

	return result
}

// FromLatLngPoint converts proto to business object
func FromLatLngPoint(pt *ridpb.LatLngPoint) *dssmodels.LatLngPoint {
	return &dssmodels.LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}
}

// === Business -> RID ===

// ToVolume4D converts Volume4D business object to proto
func ToVolume4D(vol4 *dssmodels.Volume4D) (*ridpb.Volume4D, error) {
	vol3, err := ToVolume3D(vol4.SpatialVolume)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	result := &ridpb.Volume4D{
		SpatialVolume: vol3,
	}

	if vol4.StartTime != nil {
		ts := tspb.New(*vol4.StartTime)
		result.TimeStart = ts
	}

	if vol4.EndTime != nil {
		ts := tspb.New(*vol4.EndTime)
		result.TimeEnd = ts
	}

	return result, nil
}

// ToVolume3D converts Volume3D business object to proto
func ToVolume3D(vol3 *dssmodels.Volume3D) (*ridpb.Volume3D, error) {
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
	case *dssmodels.GeoPolygon:
		result.Footprint = ToGeoPolygon(t)
	default:
		return nil, stacktrace.NewError("Unsupported geometry type: %T", vol3.Footprint)
	}

	return result, nil
}

// ToGeoPolygon converts GeoPolygon business object to proto
func ToGeoPolygon(gp *dssmodels.GeoPolygon) *ridpb.GeoPolygon {
	if gp == nil {
		return nil
	}

	result := &ridpb.GeoPolygon{}

	for _, pt := range gp.Vertices {
		result.Vertices = append(result.Vertices, ToLatLngPoint(pt))
	}

	return result
}

// ToLatLngPoint converts latlngpoint business object to proto
func ToLatLngPoint(pt *dssmodels.LatLngPoint) *ridpb.LatLngPoint {
	result := &ridpb.LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}

	return result
}

// IdentificationServiceAreaToProto converts an IdentificationServiceArea
// business object to v1 proto for API consumption.
func ToIdentificationServiceArea(i *ridmodels.IdentificationServiceArea) *ridpb.IdentificationServiceArea {
	result := &ridpb.IdentificationServiceArea{
		Id:         i.ID.String(),
		Owner:      i.Owner.String(),
		FlightsUrl: i.URL,
		Version:    i.Version.String(),
	}

	if i.StartTime != nil {
		ts := tspb.New(*i.StartTime)
		result.TimeStart = ts
	}

	if i.EndTime != nil {
		ts := tspb.New(*i.EndTime)
		result.TimeEnd = ts
	}
	return result
}

// ToSubscriberToNotify converts a subscription to a SubscriberToNotify proto
// for API consumption.
func ToSubscriberToNotify(s *ridmodels.Subscription) *ridpb.SubscriberToNotify {
	return &ridpb.SubscriberToNotify{
		Url: s.URL,
		Subscriptions: []*ridpb.SubscriptionState{
			{
				NotificationIndex: int32(s.NotificationIndex),
				SubscriptionId:    s.ID.String(),
			},
		},
	}
}

// ToSubscription converts a subscription business object to a Subscription
// proto for API consumption.
func ToSubscription(s *ridmodels.Subscription) *ridpb.Subscription {
	result := &ridpb.Subscription{
		Id:                s.ID.String(),
		Owner:             s.Owner.String(),
		Callbacks:         &ridpb.SubscriptionCallbacks{IdentificationServiceAreaUrl: s.URL},
		NotificationIndex: int32(s.NotificationIndex),
		Version:           s.Version.String(),
	}

	if s.StartTime != nil {
		ts := tspb.New(*s.StartTime)
		result.TimeStart = ts
	}

	if s.EndTime != nil {
		ts := tspb.New(*s.EndTime)
		result.TimeEnd = ts
	}
	return result
}
