package models

import (
	"errors"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/interuss/dss/pkg/dss/geo"
	dspb "github.com/interuss/dss/pkg/dssproto"
	dsserr "github.com/interuss/dss/pkg/errors"

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

type Subscription struct {
	ID                ID
	Url               string
	NotificationIndex int
	Owner             Owner
	Cells             s2.CellUnion
	StartTime         *time.Time
	EndTime           *time.Time
	Version           *Version
	AltitudeHi        *float32
	AltitudeLo        *float32
}

func (s *Subscription) ToNotifyProto() *dspb.SubscriberToNotify {
	return &dspb.SubscriberToNotify{
		Url: s.Url,
		Subscriptions: []*dspb.SubscriptionState{
			{
				NotificationIndex: int32(s.NotificationIndex),
				SubscriptionId:    s.ID.String(),
			},
		},
	}
}

func (s *Subscription) ToProto() (*dspb.Subscription, error) {
	result := &dspb.Subscription{
		Id:                s.ID.String(),
		Owner:             s.Owner.String(),
		Callbacks:         &dspb.SubscriptionCallbacks{IdentificationServiceAreaUrl: s.Url},
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

func (s *Subscription) SetExtents(extents *dspb.Volume4D) error {
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
	s.Cells, err = geo.GeoPolygonToCellIDs(footprint)
	return err
}

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
