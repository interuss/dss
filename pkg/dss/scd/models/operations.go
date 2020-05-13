package models

import (
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
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
