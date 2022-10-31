package apiv2

import (
	"time"

	restapi "github.com/interuss/dss/pkg/api/ridv2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

// === RID -> Business ===

// FromTime converts RID v2 REST model to standard golang Time
func FromTime(t *restapi.Time) (*time.Time, error) {
	if t == nil {
		return nil, nil
	}
	if t.Format != "RFC3339" {
		return nil, stacktrace.NewError("Invalid time format '%v'; expected 'RFC3339'", t.Format)
	}
	if t.Value == "" {
		return nil, stacktrace.NewError("Time structure specified, but `value` was missing")
	}
	ts, err := time.Parse(time.RFC3339Nano, t.Value)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error converting time")
	}
	return &ts, nil
}

// FromAltitude converts RID v2 REST model to float
func FromAltitude(alt *restapi.Altitude) (*float32, error) {
	if alt == nil {
		return nil, nil
	}
	if alt.Reference != "WGS84" {
		return nil, stacktrace.NewError("Invalid altitude reference '%v'; expected 'WGS84'", alt.Reference)
	}
	if alt.Units != "M" {
		return nil, stacktrace.NewError("Invalid units '%v'; expected 'M'", alt.Units)
	}
	value := float32(alt.Value)
	return &value, nil
}

// FromVolume4D converts RID v2 REST model to business object
func FromVolume4D(vol4 *restapi.Volume4D) (*dssmodels.Volume4D, error) {
	vol3, err := FromVolume3D(&vol4.Volume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing spatial volume of Volume4D")
	}

	result := &dssmodels.Volume4D{
		SpatialVolume: vol3,
	}

	result.StartTime, err = FromTime(vol4.TimeStart)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing start time of Volume4D")
	}
	result.EndTime, err = FromTime(vol4.TimeEnd)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing end time of Volume4D")
	}

	return result, nil
}

// FromVolume3D converts RID v2 REST model to business object
func FromVolume3D(vol3 *restapi.Volume3D) (*dssmodels.Volume3D, error) {
	altitudeLo, err := FromAltitude(vol3.AltitudeLower)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing lower altitude of Volume3D")
	}
	altitudeHi, err := FromAltitude(vol3.AltitudeUpper)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing upper altitude of Volume3D")
	}

	if vol3.OutlinePolygon != nil {
		if vol3.OutlineCircle != nil {
			return nil, stacktrace.NewError("Only one of outline_circle or outline_polygon may be specified")
		}
		footprint := FromPolygon(vol3.OutlinePolygon)

		result := &dssmodels.Volume3D{
			Footprint:  footprint,
			AltitudeLo: altitudeLo,
			AltitudeHi: altitudeHi,
		}

		return result, nil
	}

	if vol3.OutlineCircle != nil {
		footprint, err := FromCircle(vol3.OutlineCircle)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error parsing outline_circle for Volume3D")
		}

		result := &dssmodels.Volume3D{
			Footprint:  footprint,
			AltitudeLo: altitudeLo,
			AltitudeHi: altitudeHi,
		}

		return result, nil
	}

	return nil, stacktrace.NewError("Neither outline_polygon nor outline_circle were specified in volume")
}

// FromPolygon converts RID v2 REST model to business object
func FromPolygon(polygon *restapi.Polygon) *dssmodels.GeoPolygon {
	result := &dssmodels.GeoPolygon{}

	for _, ltlng := range polygon.Vertices {
		result.Vertices = append(result.Vertices, FromLatLngPoint(&ltlng))
	}

	return result
}

// FromCircle converts RID v2 REST model to business object
func FromCircle(circle *restapi.Circle) (*dssmodels.GeoCircle, error) {
	if circle.Center == nil {
		return nil, stacktrace.NewError("Missing `center` from circle")
	}
	if circle.Radius == nil {
		return nil, stacktrace.NewError("Missing `radius` from circle")
	}
	if circle.Radius.Units != "M" {
		return nil, stacktrace.NewError("Only circle radius units of 'M' are acceptable for UTM")
	}
	result := &dssmodels.GeoCircle{
		Center:      *FromLatLngPoint(circle.Center),
		RadiusMeter: circle.Radius.Value,
	}
	return result, nil
}

// FromLatLngPoint converts RID v2 REST model to business object
func FromLatLngPoint(pt *restapi.LatLngPoint) *dssmodels.LatLngPoint {
	return &dssmodels.LatLngPoint{
		Lat: float64(pt.Lat),
		Lng: float64(pt.Lng),
	}
}

// === Business -> RID ===

// ToTime converts standard golang Time to RID v2 REST model
func ToTime(t *time.Time) *restapi.Time {
	if t == nil {
		return nil
	}

	result := &restapi.Time{
		Format: "RFC3339",
		Value:  t.Format(time.RFC3339Nano),
	}

	return result
}

// ToLatLngPoint converts latlngpoint business object to RID v2 REST model
func ToLatLngPoint(pt *dssmodels.LatLngPoint) *restapi.LatLngPoint {
	result := &restapi.LatLngPoint{
		Lat: restapi.Latitude(pt.Lat),
		Lng: restapi.Longitude(pt.Lng),
	}

	return result
}

// ToIdentificationServiceArea converts an IdentificationServiceArea
// business object to RID v2 REST model for API consumption.
func ToIdentificationServiceArea(i *ridmodels.IdentificationServiceArea) *restapi.IdentificationServiceArea {
	result := &restapi.IdentificationServiceArea{
		Id:         restapi.EntityUUID(i.ID.String()),
		Owner:      i.Owner.String(),
		UssBaseUrl: restapi.FlightsUSSBaseURL(i.URL),
		Version:    restapi.Version(i.Version.String()),
	}
	if i.StartTime != nil {
		result.TimeStart = *ToTime(i.StartTime)
	}
	if i.EndTime != nil {
		result.TimeEnd = *ToTime(i.EndTime)
	}

	return result
}

// ToSubscriberToNotify converts a subscription to a SubscriberToNotify RID v2 REST model
// for API consumption.
func ToSubscriberToNotify(s *ridmodels.Subscription) *restapi.SubscriberToNotify {
	notifIdx := restapi.SubscriptionNotificationIndex(s.NotificationIndex)
	return &restapi.SubscriberToNotify{
		Url: restapi.URL(s.URL),
		Subscriptions: []restapi.SubscriptionState{
			{
				NotificationIndex: &notifIdx,
				SubscriptionId:    restapi.SubscriptionUUID(s.ID.String()),
			},
		},
	}
}

// ToSubscription converts a subscription business object to a Subscription
// RID v2 REST model for API consumption.
func ToSubscription(s *ridmodels.Subscription) *restapi.Subscription {
	notifIdx := restapi.SubscriptionNotificationIndex(s.NotificationIndex)
	result := &restapi.Subscription{
		Id:                restapi.SubscriptionUUID(s.ID.String()),
		Owner:             s.Owner.String(),
		UssBaseUrl:        restapi.SubscriptionUSSBaseURL(s.URL),
		NotificationIndex: &notifIdx,
		Version:           restapi.Version(s.Version.String()),
		TimeStart:         ToTime(s.StartTime),
		TimeEnd:           ToTime(s.EndTime),
	}

	return result
}
