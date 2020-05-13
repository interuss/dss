package models

import (
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// Aggregates constants for operations.
const (
	OperationStateUnknown       OperationState = ""
	OperationStateAccepted      OperationState = "Accepted"
	OperationStateActivated     OperationState = "Activated"
	OperationStateNonConforming OperationState = "NonConforming"
	OperationStateContingent    OperationState = "Contingent"
	OperationStateEnded         OperationState = "Ended"
)

// OperationState models the state of an operation.
type OperationState string

// Operation models an operation.
type Operation struct {
	ID             ID
	Version        Version
	OVN            OVN
	Owner          dssmodels.Owner
	StartTime      *time.Time
	EndTime        *time.Time
	AltitudeLower  *float32
	AltitudeUpper  *float32
	USSBaseURL     string
	State          OperationState
	Cells          s2.CellUnion
	SubscriptionID ID
}

// AdjustTimeRange adjusts the time range to the max allowed ranges on an
// Operation.
func (o *Operation) AdjustTimeRange(now time.Time, old *Operation) error {
	if o.StartTime == nil {
		// If StartTime was omitted, default to Now() for new ISAs or re-
		// use the existing time of existing ISAs.
		if old == nil {
			o.StartTime = &now
		} else {
			o.StartTime = old.StartTime
		}
	} else {
		// If setting the StartTime explicitly ensure it is not too far in the past.
		if now.Sub(*o.StartTime) > maxClockSkew {
			return dsserr.BadRequest("Operation time_start must not be in the past")
		}
	}

	// If EndTime was omitted default to the existing ISA's EndTime.
	if o.EndTime == nil && old != nil {
		o.EndTime = old.EndTime
	}

	// EndTime cannot be omitted for new ISAs.
	if o.EndTime == nil {
		return dsserr.BadRequest("Operation must have an time_end")
	}

	// EndTime cannot be before StartTime.
	if o.EndTime.Sub(*o.StartTime) < 0 {
		return dsserr.BadRequest("Operation time_end must be after time_start")
	}

	return nil
}
