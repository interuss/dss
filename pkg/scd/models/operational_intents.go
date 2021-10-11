package models

import (
	"time"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

// Aggregates constants for operational intents.
const (
	OperationalIntentStateUnknown       OperationalIntentState = ""
	OperationalIntentStateAccepted      OperationalIntentState = "Accepted"
	OperationalIntentStateActivated     OperationalIntentState = "Activated"
	OperationalIntentStateNonconforming OperationalIntentState = "Nonconforming"
	OperationalIntentStateContingent    OperationalIntentState = "Contingent"
)

// OperationState models the state of an operation.
type OperationalIntentState string

// RequiresKey indicates whether transitioning an OperationalIntent to this
// OperationalIntentState requires a valid key.
func (s OperationalIntentState) RequiresKey() bool {
	switch s {
	case OperationalIntentStateNonconforming:
		fallthrough
	case OperationalIntentStateContingent:
		return false
	}
	return true
}

// IsValid indicates whether an OperationalIntent may be transitioned to the specified
// state via a DSS PUT.
func (s OperationalIntentState) IsValidInDSS() bool {
	switch s {
	case OperationalIntentStateAccepted:
		fallthrough
	case OperationalIntentStateActivated:
		fallthrough
	case OperationalIntentStateNonconforming:
		fallthrough
	case OperationalIntentStateContingent:
		return true
	}
	return false
}

// OperationalIntent models an operational intent.
type OperationalIntent struct {
	// Reference
	ID             dssmodels.ID
	Manager        dssmodels.Manager
	Version        VersionNumber
	State          OperationalIntentState
	OVN            OVN
	StartTime      *time.Time
	EndTime        *time.Time
	USSBaseURL     string
	SubscriptionID dssmodels.ID
	AltitudeLower  *float32
	AltitudeUpper  *float32
	Cells          s2.CellUnion
}

func (s OperationalIntentState) String() string {
	return string(s)
}

// ToProto converts the OperationalIntent to its proto API format
func (o *OperationalIntent) ToProto() (*scdpb.OperationalIntentReference, error) {
	result := &scdpb.OperationalIntentReference{
		Id:              o.ID.String(),
		Ovn:             o.OVN.String(),
		Manager:         o.Manager.String(),
		Version:         int32(o.Version),
		UssBaseUrl:      o.USSBaseURL,
		SubscriptionId:  o.SubscriptionID.String(),
		State:           o.State.String(),
		UssAvailability: UssAvailabilityStateUnknown.String(),
	}

	if o.StartTime != nil {
		ts, err := ptypes.TimestampProto(*o.StartTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time to proto")
		}
		result.TimeStart = &scdpb.Time{
			Value:  ts,
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	if o.EndTime != nil {
		ts, err := ptypes.TimestampProto(*o.EndTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting end time to proto")
		}
		result.TimeEnd = &scdpb.Time{
			Value:  ts,
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	return result, nil
}

// ValidateTimeRange validates the time range of o.
func (o *OperationalIntent) ValidateTimeRange() error {
	if o.StartTime == nil {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Operation must have an time_start")
	}

	// EndTime cannot be omitted for new Operational Intents.
	if o.EndTime == nil {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Operation must have an time_end")
	}

	// EndTime cannot be before StartTime.
	if o.EndTime.Sub(*o.StartTime) < 0 {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Operation time_end must be after time_start")
	}

	return nil
}
