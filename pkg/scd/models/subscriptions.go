package models

import (
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/interuss/stacktrace"
)

const (
	// maxSubscriptionDuration is the largest allowed interval between StartTime
	// and EndTime.
	maxSubscriptionDuration = time.Hour * 24

	// maxClockSkew is the largest allowed interval between the StartTime of a new
	// subscription and the server's idea of the current time.
	maxClockSkew = time.Minute * 5
)

// Subscription represents an SCD subscription
type Subscription struct {
	ID dssmodels.ID

	// Version is an OVN-like string constructed from the Subscription's
	// updated_at field in the database; it may be unspecified when creating a new
	// Subscription in the database.
	Version                     OVN
	NotificationIndex           int
	Manager                     dssmodels.Manager
	StartTime                   *time.Time
	EndTime                     *time.Time
	AltitudeHi                  *float32
	AltitudeLo                  *float32
	USSBaseURL                  string
	NotifyForOperationalIntents bool
	NotifyForConstraints        bool
	ImplicitSubscription        bool
	Cells                       s2.CellUnion
}

// ToProto converts the Subscription to its proto API format
func (s *Subscription) ToProto(dependentOperationalIntents []dssmodels.ID) (*scdpb.Subscription, error) {
	result := &scdpb.Subscription{
		Id:                          s.ID.String(),
		Version:                     s.Version.String(),
		NotificationIndex:           int32(s.NotificationIndex),
		UssBaseUrl:                  s.USSBaseURL,
		NotifyForOperationalIntents: s.NotifyForOperationalIntents,
		NotifyForConstraints:        s.NotifyForConstraints,
		ImplicitSubscription:        s.ImplicitSubscription,
	}

	if s.StartTime != nil {
		ts, err := ptypes.TimestampProto(*s.StartTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time to proto")
		}
		result.TimeStart = &scdpb.Time{
			Value:  ts,
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	if s.EndTime != nil {
		ts, err := ptypes.TimestampProto(*s.EndTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time to proto")
		}
		result.TimeEnd = &scdpb.Time{
			Value:  ts,
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	for _, op := range dependentOperationalIntents {
		result.DependentOperationalIntents = append(result.DependentOperationalIntents, op.String())
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
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription time_start must not be in the past")
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
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription time_end must be after time_start")
	}

	// EndTime cannot be 24 hrs after StartTime
	if s.EndTime.Sub(*s.StartTime) > maxSubscriptionDuration {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription window exceeds 24 hours")
	}

	return nil
}

// ValidateDependentOps validates subscription against given operations in all 4 dimensions
func (s *Subscription) ValidateDependentOps(operationalIntents []*OperationalIntent) error {
	for _, op := range operationalIntents {
		if err := s.ValidateDependentOp(op); err != nil {
			return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Subscription does not cover dependent operations")
		}
	}
	return nil
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

// ValidateDependentOp validates subscription against single operation in all 4 dimensions
func (s *Subscription) ValidateDependentOp(operationalIntent *OperationalIntent) error {
	// validate 2d area
	if !s.Cells.Contains(operationalIntent.Cells) {
		return stacktrace.NewError("Subscription does not cover dependent operation 2d area, %s", operationalIntent.ID)
	}
	// validate altitudes
	if *operationalIntent.AltitudeLower < *s.AltitudeLo || *operationalIntent.AltitudeUpper > *s.AltitudeHi {
		return stacktrace.NewError("Subscription does not cover dependent operation altitude, %s", operationalIntent.ID)
	}
	// validate time range
	// Check if subscription start time is no later than the maximum latency (5 minutes) gap with dependent operation start time
	if (*s.StartTime).Sub(*operationalIntent.StartTime).Minutes() > 5 {
		return stacktrace.NewError("Subscription start time does not cover dependent operation's start time, %s", operationalIntent.ID)
	}

	if (*operationalIntent.EndTime).After(*s.EndTime) {
		return stacktrace.NewError("Subscription does not cover dependent operation's end time, %s", operationalIntent.ID)
	}
	return nil
}
