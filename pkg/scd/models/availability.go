package models

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
	"strings"
)

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

func UssAvailabilityStateFromString(s string) (UssAvailabilityState, error) {
	switch strings.ToLower(s) {
	case "", "unknown":
		return UssAvailabilityStateUnknown, nil
	case "normal":
		return UssAvailabilityStateNormal, nil
	case "down":
		return UssAvailabilityStateDown, nil
	default:
		return UssAvailabilityStateUnknown, stacktrace.NewError("Invalid USS availability state")
	}
}
