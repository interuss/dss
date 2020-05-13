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

// ValidateTimeRange validates the time range of o.
func (o *Operation) ValidateTimeRange() error {
	if o.StartTime == nil {
		return dsserr.BadRequest("Operation must have an time_start")
	}

	// EndTime cannot be omitted for new Operations.
	if o.EndTime == nil {
		return dsserr.BadRequest("Operation must have an time_end")
	}

	// EndTime cannot be before StartTime.
	if o.EndTime.Sub(*o.StartTime) < 0 {
		return dsserr.BadRequest("Operation time_end must be after time_start")
	}

	return nil
}
