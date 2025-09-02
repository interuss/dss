package models

import (
	"strings"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
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

func (u UssAvailabilityState) ToRest() restapi.UssAvailabilityState {
	return restapi.UssAvailabilityState(u)
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

func UssAvailabilityStateFromRest(s restapi.UssAvailabilityState) (UssAvailabilityState, error) {
	return UssAvailabilityStateFromString(string(s))
}

func (u UssAvailabilityStatus) ToRest() *restapi.UssAvailabilityStatus {
	return &restapi.UssAvailabilityStatus{
		Uss:          u.Uss.String(),
		Availability: u.Availability.ToRest(),
	}
}
