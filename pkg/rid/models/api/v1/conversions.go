package apiv1

import (
	"time"

	restapi "github.com/interuss/dss/pkg/api/ridv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

// === RID -> Business ===

// FromVolume4D converts RID v1 REST model to business object
func FromVolume4D(vol4 *restapi.Volume4D) (*dssmodels.Volume4D, error) {
	vol3, err := FromVolume3D(&vol4.SpatialVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error converting spatial volume")
	}
	result := &dssmodels.Volume4D{SpatialVolume: vol3}

	if vol4.TimeStart != nil {
		ts, err := time.Parse(time.RFC3339Nano, *vol4.TimeStart)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time")
		}
		result.StartTime = &ts
	}

	if vol4.TimeEnd != nil {
		ts, err := time.Parse(time.RFC3339Nano, *vol4.TimeEnd)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time")
		}
		result.EndTime = &ts
	}

	return result, nil
}

// FromVolume3D converts RID v1 REST model to business object
func FromVolume3D(vol3 *restapi.Volume3D) (*dssmodels.Volume3D, error) {
	p, err := FromGeoPolygon(&vol3.Footprint)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error converting polygon")
	}
	return &dssmodels.Volume3D{
		Footprint:  p,
		AltitudeLo: (*float32)(vol3.AltitudeLo),
		AltitudeHi: (*float32)(vol3.AltitudeHi),
	}, nil
}

// FromGeoPolygon converts RID v1 REST model to business object
func FromGeoPolygon(footprint *restapi.GeoPolygon) (*dssmodels.GeoPolygon, error) {
	result := &dssmodels.GeoPolygon{}

	for _, ltlng := range footprint.Vertices {
		v, err := FromLatLngPoint(&ltlng)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting polygon vertex")
		}
		result.Vertices = append(result.Vertices, v)
	}

	return result, nil
}

// FromLatLngPoint converts RID v1 REST model to business object
func FromLatLngPoint(pt *restapi.LatLngPoint) (*dssmodels.LatLngPoint, error) {
	return dssmodels.NewLatLngPoint(float64(pt.Lat), float64(pt.Lng))
}

// === Business -> RID ===

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
		result.TimeStart = i.StartTime.Format(time.RFC3339Nano)
	}

	if i.EndTime != nil {
		result.TimeEnd = i.EndTime.Format(time.RFC3339Nano)
	}
	return result
}

// MakeSubscribersToNotify groups the passed subscriptions by their callback URL,
// returning a collection of subscribers to notify that contains one entry per distinct callback URL.
func MakeSubscribersToNotify(subscriptions []*ridmodels.Subscription) []restapi.SubscriberToNotify {
	subscriptionsByURL := map[string][]restapi.SubscriptionState{}
	for _, sub := range subscriptions {
		notifIdx := restapi.SubscriptionNotificationIndex(sub.NotificationIndex)
		subID := restapi.SubscriptionUUID(sub.ID)
		subState := restapi.SubscriptionState{
			SubscriptionId:    &subID,
			NotificationIndex: &notifIdx,
		}
		subscriptionsByURL[sub.URL] = append(subscriptionsByURL[sub.URL], subState)
	}

	result := []restapi.SubscriberToNotify{}
	for url, states := range subscriptionsByURL {
		result = append(result, restapi.SubscriberToNotify{
			Url:           restapi.URL(url),
			Subscriptions: states,
		})
	}

	return result
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
		ts := s.StartTime.Format(time.RFC3339Nano)
		result.TimeStart = &ts
	}

	if s.EndTime != nil {
		ts := s.EndTime.Format(time.RFC3339Nano)
		result.TimeEnd = &ts
	}
	return result
}
