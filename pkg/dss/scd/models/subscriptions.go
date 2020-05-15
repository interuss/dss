package models

import (
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	dsserr "github.com/interuss/dss/pkg/errors"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
)

const (
	// maxSubscriptionDuration is the largest allowed interval between StartTime
	// and EndTime.
	maxSubscriptionDuration = time.Hour * 24

	// maxClockSkew is the largest allowed interval between the StartTime of a new
	// subscription and the server's idea of the current time.
	maxClockSkew = time.Minute * 5
)

type Subscription struct {
	ID                ID
	Version           Version
	NotificationIndex int
	Owner             dssmodels.Owner
	StartTime         *time.Time
	EndTime           *time.Time
	AltitudeHi        *float32
	AltitudeLo        *float32

	BaseURL              string
	NotifyForOperations  bool
	NotifyForConstraints bool
	ImplicitSubscription bool
	DependentOperations  []ID
	Cells                s2.CellUnion
}

func (s *Subscription) ToProto() (*scdpb.Subscription, error) {
	result := &scdpb.Subscription{
		Id:                   s.ID.String(),
		Version:              int32(s.Version),
		NotificationIndex:    int32(s.NotificationIndex),
		UssBaseUrl:           s.BaseURL,
		NotifyForOperations:  s.NotifyForOperations,
		NotifyForConstraints: s.NotifyForConstraints,
		ImplicitSubscription: s.ImplicitSubscription,
	}

	for i := 0; i < len(s.DependentOperations); i++ {
		result.DependentOperations = append(result.DependentOperations, s.DependentOperations[i].String())
	}

	if s.StartTime != nil {
		ts, err := ptypes.TimestampProto(*s.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = &scdpb.Time{
			Value:  ts,
			Format: TimeFormatRfc3339,
		}
	}

	if s.EndTime != nil {
		ts, err := ptypes.TimestampProto(*s.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd = &scdpb.Time{
			Value:  ts,
			Format: TimeFormatRfc3339,
		}
	}
	return result, nil
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
