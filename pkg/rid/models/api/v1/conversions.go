package apiv1

import (
	"time"

	restapi "github.com/interuss/dss/pkg/api/rid_v1"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

// === RID -> Business ===

// FromVolume4D converts RID v1 REST model to business object
func FromVolume4D(vol4 *restapi.Volume4D) (*dssmodels.Volume4D, error) {
	result := &dssmodels.Volume4D{
		SpatialVolume: FromVolume3D(&vol4.SpatialVolume),
	}

	if vol4.TimeStart != nil {
		ts, err := time.Parse(time.RFC3339, *vol4.TimeStart)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time from proto")
		}
		result.StartTime = &ts
	}

	if vol4.TimeEnd != nil {
		ts, err := time.Parse(time.RFC3339, *vol4.TimeEnd)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time from proto")
		}
		result.EndTime = &ts
	}

	return result, nil
}

// FromVolume3D converts RID v1 REST model to business object
func FromVolume3D(vol3 *restapi.Volume3D) *dssmodels.Volume3D {
	return &dssmodels.Volume3D{
		Footprint:  FromGeoPolygon(&vol3.Footprint),
		AltitudeLo: (*float32)(vol3.AltitudeLo),
		AltitudeHi: (*float32)(vol3.AltitudeHi),
	}
}

// FromGeoPolygon converts RID v1 REST model to business object
func FromGeoPolygon(footprint *restapi.GeoPolygon) *dssmodels.GeoPolygon {
	result := &dssmodels.GeoPolygon{}

	for _, ltlng := range footprint.Vertices {
		result.Vertices = append(result.Vertices, FromLatLngPoint(&ltlng))
	}

	return result
}

// FromLatLngPoint converts RID v1 REST model to business object
func FromLatLngPoint(pt *restapi.LatLngPoint) *dssmodels.LatLngPoint {
	return &dssmodels.LatLngPoint{
		Lat: float64(pt.Lat),
		Lng: float64(pt.Lng),
	}
}

// === Business -> RID ===

// ToVolume4D converts Volume4D business object to a RID v1 REST model
func ToVolume4D(vol4 *dssmodels.Volume4D) (*restapi.Volume4D, error) {
	vol3, err := ToVolume3D(vol4.SpatialVolume)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}

	result := &restapi.Volume4D{
		SpatialVolume: *vol3,
	}

	if vol4.StartTime != nil {
		ts := vol4.StartTime.Format(time.RFC3339)
		result.TimeStart = &ts
	}

	if vol4.EndTime != nil {
		ts := vol4.EndTime.Format(time.RFC3339)
		result.TimeEnd = &ts
	}

	return result, nil
}

// ToVolume3D converts Volume3D business object to a RID v1 REST model
func ToVolume3D(vol3 *dssmodels.Volume3D) (*restapi.Volume3D, error) {
	if vol3 == nil {
		return nil, nil
	}

	result := &restapi.Volume3D{}

	if vol3.AltitudeLo != nil {
		result.AltitudeLo = (*restapi.Altitude)(vol3.AltitudeLo)
	}

	if vol3.AltitudeHi != nil {
		result.AltitudeHi = (*restapi.Altitude)(vol3.AltitudeHi)
	}

	switch t := vol3.Footprint.(type) {
	case nil:
		// Empty on purpose
	case *dssmodels.GeoPolygon:
		result.Footprint = *ToGeoPolygon(t)
	default:
		return nil, stacktrace.NewError("Unsupported geometry type: %T", vol3.Footprint)
	}

	return result, nil
}

// ToGeoPolygon converts GeoPolygon business object to a RID v1 REST model
func ToGeoPolygon(gp *dssmodels.GeoPolygon) *restapi.GeoPolygon {
	if gp == nil {
		return nil
	}

	result := &restapi.GeoPolygon{}

	for _, pt := range gp.Vertices {
		result.Vertices = append(result.Vertices, *ToLatLngPoint(pt))
	}

	return result
}

// ToLatLngPoint converts latlngpoint business object to a RID v1 REST model
func ToLatLngPoint(pt *dssmodels.LatLngPoint) *restapi.LatLngPoint {
	result := &restapi.LatLngPoint{
		Lat: restapi.Latitude(pt.Lat),
		Lng: restapi.Longitude(pt.Lng),
	}

	return result
}

// ToIdentificationServiceArea converts an IdentificationServiceArea
// business object to a RID v1 REST model for API consumption.
func ToIdentificationServiceArea(i *ridmodels.IdentificationServiceArea) *restapi.IdentificationServiceArea {
	result := &restapi.IdentificationServiceArea{
		Id:         restapi.EntityUUID(i.ID.String()),
		Owner:      i.Owner.String(),
		FlightsUrl: restapi.RIDFlightsURL(i.URL),
		Version:    restapi.Version(i.Version.String()),
	}

	if i.StartTime != nil {
		result.TimeStart = i.StartTime.Format(time.RFC3339)
	}

	if i.EndTime != nil {
		result.TimeEnd = i.EndTime.Format(time.RFC3339)
	}
	return result
}

// ToSubscriberToNotify converts a Subscription to a SubscriberToNotify RID v1
// REST model for API consumption.
func ToSubscriberToNotify(s *ridmodels.Subscription) *restapi.SubscriberToNotify {
	notifIndex := restapi.SubscriptionNotificationIndex(s.NotificationIndex)
	subID := restapi.SubscriptionUUID(s.ID.String())
	return &restapi.SubscriberToNotify{
		Url: restapi.URL(s.URL),
		Subscriptions: []restapi.SubscriptionState{
			{
				NotificationIndex: &notifIndex,
				SubscriptionId:    &subID,
			},
		},
	}
}

// ToSubscription converts a subscription business object to a Subscription RID
// v1 REST model for API consumption.
func ToSubscription(s *ridmodels.Subscription) *restapi.Subscription {
	result := &restapi.Subscription{
		Id:    restapi.SubscriptionUUID(s.ID.String()),
		Owner: s.Owner.String(),
		Callbacks: restapi.SubscriptionCallbacks{
			IdentificationServiceAreaUrl: (*restapi.IdentificationServiceAreaURL)(&s.URL),
		},
		NotificationIndex: restapi.SubscriptionNotificationIndex(s.NotificationIndex),
		Version:           restapi.Version(s.Version.String()),
	}

	if s.StartTime != nil {
		ts := s.StartTime.Format(time.RFC3339)
		result.TimeStart = &ts
	}

	if s.EndTime != nil {
		ts := s.EndTime.Format(time.RFC3339)
		result.TimeEnd = &ts
	}
	return result
}
