package models

import (
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
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
	UpdatedAt         *time.Time
	AltitudeHi        *float32
	AltitudeLo        *float32
}

// Apply fields from s2 onto s, preferring any fields set in s2 except for ID
// and Owner.
func (s *Subscription) Apply(s2 *Subscription) *Subscription {
	new := *s
	if s2.Url != "" {
		new.Url = s2.Url
	}
	if s2.Cells != nil {
		new.Cells = s2.Cells
	}
	if s2.StartTime != nil {
		new.StartTime = s2.StartTime
	}
	if s2.EndTime != nil {
		new.EndTime = s2.EndTime
	}
	if s2.UpdatedAt != nil {
		new.UpdatedAt = s2.UpdatedAt
	}
	if s2.AltitudeHi != nil {
		new.AltitudeHi = s2.AltitudeHi
	}
	if s2.AltitudeLo != nil {
		new.AltitudeLo = s2.AltitudeLo
	}
	return &new
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

func (s *Subscription) Version() Version {
	return VersionFromTimestamp(s.UpdatedAt)
}

func (s *Subscription) ToProto() (*dspb.Subscription, error) {
	result := &dspb.Subscription{
		Id:                s.ID.String(),
		Owner:             s.Owner.String(),
		Url:               s.Url,
		NotificationIndex: int32(s.NotificationIndex),
		Version:           s.Version().String(),
	}

	if s.StartTime != nil {
		ts, err := ptypes.TimestampProto(*s.StartTime)
		if err != nil {
			return nil, err
		}
		result.StartTime = ts
	}

	if s.EndTime != nil {
		ts, err := ptypes.TimestampProto(*s.EndTime)
		if err != nil {
			return nil, err
		}
		result.EndTime = ts
	}
	return result, nil
}

func (s *Subscription) SetExtents(extents *dspb.Volume4D) error {
	var err error
	if extents == nil {
		return nil
	}
	if startTime := extents.GetStartTime(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return err
		}
		s.StartTime = &ts
	}

	if endTime := extents.GetEndTime(); endTime != nil {
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
