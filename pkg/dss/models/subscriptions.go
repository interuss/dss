package models

import (
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	"github.com/steeling/InterUSS-Platform/pkg/dssproto"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
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
			&dspb.SubscriptionState{
				NotificationIndex: int32(s.NotificationIndex),
				Subscription:      s.ID.String(),
			},
		},
	}
}

func (s *Subscription) ToProto() (*dspb.Subscription, error) {
	result := &dspb.Subscription{
		Id:                s.ID.String(),
		Owner:             s.Owner.String(),
		Callbacks:         &dssproto.SubscriptionCallbacks{IdentificationServiceAreaUrl: s.Url},
		NotificationIndex: int32(s.NotificationIndex),
		Version:           s.Version.String(),
	}

	if s.StartTime != nil {
		ts, err := ptypes.TimestampProto(*s.StartTime)
		if err != nil {
			return nil, err
		}
		result.Begins = ts
	}

	if s.EndTime != nil {
		ts, err := ptypes.TimestampProto(*s.EndTime)
		if err != nil {
			return nil, err
		}
		result.Expires = ts
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
		return nil
	}
	if wrapper := space.GetAltitudeHi(); wrapper != nil {
		s.AltitudeHi = ptrToFloat32(wrapper.GetValue())
	}
	if wrapper := space.GetAltitudeLo(); wrapper != nil {
		s.AltitudeLo = ptrToFloat32(wrapper.GetValue())
	}
	footprint := space.GetFootprint()
	if footprint == nil {
		return nil
	}
	s.Cells, err = geo.GeoPolygonToCellIDs(footprint)

	return err
}
