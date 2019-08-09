package models

import (
	"time"

	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
)

type Subscription struct {
	// Embed the proto
	// Unfortunately some types don't implement scanner/valuer, so we add placeholders below.
	ID                ID
	Url               string
	NotificationIndex int
	Owner             Owner
	Cells             s2.CellUnion
	// TODO(steeling): abstract NullTime away from models.
	StartTime  *time.Time
	EndTime    *time.Time
	UpdatedAt  *time.Time
	AltitudeHi *float32
	AltitudeLo *float32
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
