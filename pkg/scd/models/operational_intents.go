package models

import (
	"time"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
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

// IsValidInDSS indicates whether an OperationalIntent may be transitioned to the specified
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
	ID              dssmodels.ID
	Manager         dssmodels.Manager
	UssAvailability UssAvailabilityState
	Version         VersionNumber
	State           OperationalIntentState
	OVN             OVN
	StartTime       *time.Time
	EndTime         *time.Time
	USSBaseURL      string
	SubscriptionID  dssmodels.ID
	AltitudeLower   *float32
	AltitudeUpper   *float32
	Cells           s2.CellUnion
}

func (s OperationalIntentState) String() string {
	return string(s)
}

func (s OperationalIntentState) ToRest() restapi.OperationalIntentState {
	return restapi.OperationalIntentState(s)
}

// ToRest converts the OperationalIntent to its SCD v1 REST model API format
func (o *OperationalIntent) ToRest() *restapi.OperationalIntentReference {
	ovn := restapi.EntityOVN(o.OVN.String())
	result := &restapi.OperationalIntentReference{
		Id:              restapi.EntityID(o.ID.String()),
		Ovn:             &ovn,
		Manager:         o.Manager.String(),
		Version:         int32(o.Version),
		UssBaseUrl:      restapi.OperationalIntentUssBaseURL(o.USSBaseURL),
		SubscriptionId:  restapi.SubscriptionID(o.SubscriptionID.String()),
		State:           o.State.ToRest(),
		UssAvailability: o.UssAvailability.ToRest(),
	}

	if o.StartTime != nil {
		result.TimeStart = restapi.Time{
			Value:  o.StartTime.Format(time.RFC3339Nano),
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	if o.EndTime != nil {
		result.TimeEnd = restapi.Time{
			Value:  o.EndTime.Format(time.RFC3339Nano),
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	return result
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

// SetCells is a convenience function that accepts an int64 array and converts
// to s2.CellUnion.
// TODO: wrap s2.CellUnion in a custom type that embeds the struct such that
// we can still call its function directly, but also implements scan for sql
// driver.
func (o *OperationalIntent) SetCells(cids []int64) {
	cells := s2.CellUnion{}
	for _, id := range cids {
		cells = append(cells, s2.CellID(id))
	}
	o.Cells = cells
}

// RequiresKey indicates whether this OperationalIntent requires its OVN to be included in the provided keys when
// another intersecting OperationalIntent is being created or updated.
func (o *OperationalIntent) RequiresKey() bool {
	return !(o.UssAvailability == UssAvailabilityStateDown && o.State == OperationalIntentStateAccepted)
}
