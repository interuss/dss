package models

import (
	"errors"
	"time"

	"github.com/interuss/dss/pkg/api/v1/ridpb"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"google.golang.org/protobuf/proto"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
)

var (
	// maxSubscriptionDuration is the largest allowed interval between StartTime
	// and EndTime.
	maxSubscriptionDuration = time.Hour * 24

	// maxClockSkew is the largest allowed interval between the StartTime of a new
	// subscription and the server's idea of the current time.
	maxClockSkew = time.Minute * 5
)

// Subscription represents a USS subscription over a given 4D volume.
type Subscription struct {
	ID                dssmodels.ID
	URL               string
	NotificationIndex int
	Owner             dssmodels.Owner
	Cells             s2.CellUnion
	StartTime         *time.Time
	EndTime           *time.Time
	Version           *dssmodels.Version
	AltitudeHi        *float32
	AltitudeLo        *float32
}

// SetCells is a convenience function that accepts an int64 array and converts
// to s2.CellUnion.
// TODO: wrap s2.CellUnion in a custom type that embeds the struct such that
// we can still call its function directly, but also implements scan for sql
// driver.
func (s *Subscription) SetCells(cids []int64) {
	cells := s2.CellUnion{}
	for _, id := range cids {
		cells = append(cells, s2.CellID(id))
	}
	s.Cells = cells
}

// ToNotifyProto converts a subscription to a SubscriberToNotify proto for
// API consumption.
func (s *Subscription) ToNotifyProto() *ridpb.SubscriberToNotify {
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

// ToProto converts a subscription struct to a Subscription proto for
// API consumption.
func (s *Subscription) ToProto() (*ridpb.Subscription, error) {
	result := &ridpb.Subscription{
		Id:                s.ID.String(),
		Owner:             s.Owner.String(),
		Callbacks:         &ridpb.SubscriptionCallbacks{IdentificationServiceAreaUrl: s.URL},
		NotificationIndex: int32(s.NotificationIndex),
		Version:           s.Version.String(),
	}

	if s.StartTime != nil {
		ts, err := ptypes.TimestampProto(*s.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = ts
	}

	if s.EndTime != nil {
		ts, err := ptypes.TimestampProto(*s.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd = ts
	}
	return result, nil
}

// SetExtents performs some data validation and sets the 4D volume on the
// Subscription.
func (s *Subscription) SetExtents(extents *ridpb.Volume4D) error {
	var err error
	if extents == nil {
		return nil
	}
	if startTime := extents.GetTimeStart(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return err
		}
		s.StartTime = &ts
	}

	if endTime := extents.GetTimeEnd(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return err
		}
		s.EndTime = &ts
	}

	space := extents.GetSpatialVolume()
	if space == nil {
		return errors.New("missing required spatial_volume")
	}
	s.AltitudeHi = proto.Float32(space.GetAltitudeHi())
	s.AltitudeLo = proto.Float32(space.GetAltitudeLo())
	footprint := space.GetFootprint()
	if footprint == nil {
		return errors.New("spatial_volume missing required footprint")
	}
	s.Cells, err = dssmodels.GeoPolygonFromRIDProto(footprint).CalculateCovering()
	return err
}

// AdjustTimeRange adjusts the time range to the max allowed ranges on a
// subscription.
func (s *Subscription) AdjustTimeRange(now time.Time, old *Subscription) error {
	if s.StartTime == nil {
		// If StartTime was omitted, default to Now() for new subscriptions or re-
		// use the existing time of existing subscriptions.
		if old == nil {
			s.StartTime = &now
		} else {
			s.StartTime = old.StartTime
		}
	} else {
		// If setting the StartTime explicitly ensure it is not too far in the past.
		if now.Sub(*s.StartTime) > maxClockSkew {
			return dsserr.BadRequest("subscription time_start must not be in the past")
		}
	}

	// If EndTime was omitted default to the existing subscription's EndTime.
	if s.EndTime == nil && old != nil {
		s.EndTime = old.EndTime
	}

	// Or if this is a new subscription default to StartTime + 1 day.
	if s.EndTime == nil {
		truncatedEndTime := s.StartTime.Add(maxSubscriptionDuration)
		s.EndTime = &truncatedEndTime
	}

	// EndTime cannot be before StartTime.
	if s.EndTime.Sub(*s.StartTime) < 0 {
		return dsserr.BadRequest("subscription time_end must be after time_start")
	}

	// EndTime cannot be 24 hrs after StartTime
	if s.EndTime.Sub(*s.StartTime) > maxSubscriptionDuration {
		return dsserr.BadRequest("subscription window exceeds 24 hours")
	}

	return nil
}
