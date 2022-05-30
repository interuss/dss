package apiv2

import (
	"time"

	ridpb "github.com/interuss/dss/pkg/api/v2/ridpbv2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// === RID -> Business ===

// FromTime converts proto to standard golang Time
func FromTime(t *ridpb.Time) (*time.Time, error) {
	if t == nil {
		return nil, nil
	}
	format := t.GetFormat()
	if format != "RFC3339" {
		return nil, stacktrace.NewError("Invalid time format '" + format + "'; expected 'RFC3339'")
	}
	value := t.GetValue()
	if value == nil {
		return nil, stacktrace.NewError("Time structure specified, but `value` was missing")
	}
	err := value.CheckValid()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error converting time from proto")
	}
	ts := value.AsTime()
	return &ts, nil
}

// FromAltitude converts proto to float
func FromAltitude(alt *ridpb.Altitude) (*float32, error) {
	if alt == nil {
		return nil, nil
	}
	ref := alt.GetReference()
	if ref != "WGS84" {
		return nil, stacktrace.NewError("Invalid altitude reference '" + ref + "'; expected 'WGS84'")
	}
	units := alt.GetUnits()
	if units != "M" {
		return nil, stacktrace.NewError("Invalid units '" + units + "'; expected 'M'")
	}
	value := float32(alt.GetValue())
	return &value, nil
}

// FromVolume4D converts proto to business object
func FromVolume4D(vol4 *ridpb.Volume4D) (*dssmodels.Volume4D, error) {
	vol3, err := FromVolume3D(vol4.GetVolume())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing spatial volume of Volume4D")
	}

	result := &dssmodels.Volume4D{
		SpatialVolume: vol3,
	}

	result.StartTime, err = FromTime(vol4.GetTimeStart())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing start time of Volume4D")
	}
	result.EndTime, err = FromTime(vol4.GetTimeEnd())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing end time of Volume4D")
	}

	return result, nil
}

// FromVolume3D converts proto to business object
func FromVolume3D(vol3 *ridpb.Volume3D) (*dssmodels.Volume3D, error) {
	altitudeLo, err := FromAltitude(vol3.GetAltitudeLower())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing lower altitude of Volume3D")
	}
	altitudeHi, err := FromAltitude(vol3.GetAltitudeUpper())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing upper altitude of Volume3D")
	}

	polygon := vol3.GetOutlinePolygon()
	if polygon != nil {
		footprint := FromPolygon(polygon)

		result := &dssmodels.Volume3D{
			Footprint:  footprint,
			AltitudeLo: altitudeLo,
			AltitudeHi: altitudeHi,
		}

		return result, nil
	}

	circle := vol3.GetOutlineCircle()
	if circle != nil {
		footprint, err := FromCircle(circle)
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

// FromPolygon converts proto to business object
func FromPolygon(polygon *ridpb.Polygon) *dssmodels.GeoPolygon {
	result := &dssmodels.GeoPolygon{}

	for _, ltlng := range polygon.Vertices {
		result.Vertices = append(result.Vertices, FromLatLngPoint(ltlng))
	}

	return result
}

// FromCircle converts proto to business object
func FromCircle(circle *ridpb.Circle) (*dssmodels.GeoCircle, error) {
	center := circle.GetCenter()
	if center == nil {
		return nil, stacktrace.NewError("Missing `center` from circle")
	}
	radius := circle.GetRadius()
	if radius == nil {
		return nil, stacktrace.NewError("Missing `radius` from circle")
	}
	if radius.GetUnits() != "M" {
		return nil, stacktrace.NewError("Only circle radius units of 'M' are acceptable for UTM")
	}
	result := &dssmodels.GeoCircle{
		Center:      *FromLatLngPoint(center),
		RadiusMeter: radius.GetValue(),
	}
	return result, nil
}

// FromLatLngPoint converts proto to business object
func FromLatLngPoint(pt *ridpb.LatLngPoint) *dssmodels.LatLngPoint {
	return &dssmodels.LatLngPoint{
		Lat: pt.Lat,
		Lng: pt.Lng,
	}
}

// === Business -> RID ===

// ToTime converts standard golang Time to proto
func ToTime(t *time.Time) *ridpb.Time {
	if t == nil {
		return nil
	}

	result := &ridpb.Time{
		Format: "RFC3339",
		Value:  tspb.New(*t),
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
// business object to v2 proto for API consumption.
func ToIdentificationServiceArea(i *ridmodels.IdentificationServiceArea) *ridpb.IdentificationServiceArea {
	result := &ridpb.IdentificationServiceArea{
		Id:         i.ID.String(),
		Owner:      i.Owner.String(),
		UssBaseUrl: i.URL,
		Version:    i.Version.String(),
		TimeStart:  ToTime(i.StartTime),
		TimeEnd:    ToTime(i.EndTime),
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
		UssBaseUrl:        s.URL,
		NotificationIndex: int32(s.NotificationIndex),
		Version:           s.Version.String(),
		TimeStart:         ToTime(s.StartTime),
		TimeEnd:           ToTime(s.EndTime),
	}

	return result
}
