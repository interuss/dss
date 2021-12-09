package models

import dssmodels "github.com/interuss/dss/pkg/models"

// Aggregates constants for uss availability.
const (
	UssAvailabilityStateUnknown UssAvailabilityState = "Unknown"
	UssAvailabilityStateNormal  UssAvailabilityState = "Normal"
	UssAvailabilityStateDown    UssAvailabilityState = "Down"
)

// UssAvailabilityState models the state of an uss availability.
type UssAvailabilityState string

// UssAvailabilityStatus models an uss availability status.
type UssAvailabilityStatus struct {
	Uss          dssmodels.Manager
	Availability UssAvailabilityState
	Version      OVN
}

func (u UssAvailabilityState) String() string {
	return string(u)
}
