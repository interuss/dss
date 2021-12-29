package models

import (
	dsserr "github.com/interuss/dss/pkg/errors"
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
	if s == "" {
		// Set availability default to Unknown
		s = "Unknown"
	} else {
		s = strings.Title(strings.ToLower(s))
	}
	switch UssAvailabilityState(s) {
	case UssAvailabilityStateUnknown, UssAvailabilityStateNormal, UssAvailabilityStateDown:
		return UssAvailabilityState(s), nil
	}
	return "", stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid USS availability state")
}
