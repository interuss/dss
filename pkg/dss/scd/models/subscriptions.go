package models

import (
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	commonmodels "github.com/interuss/dss/pkg/dss/models"

	"github.com/golang/protobuf/ptypes"
)

type Subscription struct {
	ID                ID
	Version           int
	NotificationIndex int
	Owner             commonmodels.Owner
	StartTime         *time.Time
	EndTime           *time.Time
	AltitudeHi        *float32
	AltitudeLo        *float32

	BaseURL              string
	NotifyForOperations  bool
	NotifyForConstraints bool
	ImplicitSubscription bool
	DependentOperations  []ID
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
		result.TimeStart.Value = ts
		result.TimeStart.Format = TimeFormatRfc3339
	}

	if s.EndTime != nil {
		ts, err := ptypes.TimestampProto(*s.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd.Value = ts
		result.TimeStart.Format = TimeFormatRfc3339
	}
	return result, nil
}
