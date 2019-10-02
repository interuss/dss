package models

import (
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dssv1"

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
	S2Model
	NotificationIndex int
}

func (s *Subscription) ToNotifyProto() *dssv1.SubscriberToNotify {
	return &dssv1.SubscriberToNotify{
		Url: s.Url,
		Subscriptions: []*dssv1.SubscriptionState{
			&dssv1.SubscriptionState{
				NotificationIndex: int32(s.NotificationIndex),
				SubscriptionId:    s.ID.String(),
			},
		},
	}
}

func (s *Subscription) ToProto() (*dssv1.Subscription, error) {
	result := &dssv1.Subscription{
		Id:                s.ID.String(),
		Owner:             s.Owner.String(),
		Callbacks:         &dssv1.SubscriptionCallbacks{IdentificationServiceAreaUrl: s.Url},
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
