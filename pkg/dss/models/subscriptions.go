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
	ID                string
	Url               string
	NotificationIndex int
	Owner             string
	Cells             s2.CellUnion
	// TODO(steeling): abstract NullTime away from models.
	StartTime  NullTime
	EndTime    NullTime
	UpdatedAt  time.Time
	AltitudeHi float32
	AltitudeLo float32
}

// Apply fields from s2 onto s, preferring any fields set in s2.
func (s *Subscription) Apply(s2 *Subscription) *Subscription {
	new := *s
	if s2.Url != "" {
		new.Url = s2.Url
	}
	if s2.Cells != nil {
		new.Cells = s2.Cells
	}
	if s2.StartTime.Valid {
		new.StartTime = s2.StartTime
	}
	if s2.EndTime.Valid {
		new.EndTime = s2.EndTime
	}
	if !s2.UpdatedAt.IsZero() {
		new.UpdatedAt = s2.UpdatedAt
	}
	if s2.AltitudeHi != 0 {
		new.AltitudeHi = s2.AltitudeHi
	}
	// TODO(steeling) what if the update is to make it 0, we need an omitempty, pointer, or some other type.
	if s2.AltitudeLo != 0 {
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
				Subscription:      s.ID,
			},
		},
	}
}

func (s *Subscription) Version() string {
	return timestampToVersionString(s.UpdatedAt)
}

func (s *Subscription) ToProto() (*dspb.Subscription, error) {
	result := &dspb.Subscription{
		Id:                s.ID,
		Owner:             s.Owner,
		Url:               s.Url,
		NotificationIndex: int32(s.NotificationIndex),
		Version:           s.Version(),
	}

	if s.StartTime.Valid {
		ts, err := ptypes.TimestampProto(s.StartTime.Time)
		if err != nil {
			return nil, err
		}
		result.StartTime = ts
	}

	if s.EndTime.Valid {
		ts, err := ptypes.TimestampProto(s.EndTime.Time)
		if err != nil {
			return nil, err
		}
		result.EndTime = ts
	}
	return result, nil
}
