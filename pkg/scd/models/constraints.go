package models

import (
	"time"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

// Constraint models a constraint, as known by the DSS
type Constraint struct {
	ID              dssmodels.ID
	Manager         dssmodels.Manager
	UssAvailability UssAvailabilityState
	Version         VersionNumber
	OVN             OVN
	StartTime       *time.Time
	EndTime         *time.Time
	USSBaseURL      string
	AltitudeLower   *float32
	AltitudeUpper   *float32
	Cells           s2.CellUnion
}

// ToRest converts the Constraint to its SCD v1 REST model API format
func (c *Constraint) ToRest() *restapi.ConstraintReference {
	ovn := restapi.EntityOVN(c.OVN.String())
	result := &restapi.ConstraintReference{
		Id:              restapi.EntityID(c.ID.String()),
		Ovn:             &ovn,
		Manager:         c.Manager.String(),
		Version:         int32(c.Version),
		UssBaseUrl:      restapi.ConstraintUssBaseURL(c.USSBaseURL),
		UssAvailability: UssAvailabilityStateUnknown.ToRest(),
	}

	if c.StartTime != nil {
		result.TimeStart = restapi.Time{
			Value:  c.StartTime.Format(time.RFC3339Nano),
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	if c.EndTime != nil {
		result.TimeEnd = restapi.Time{
			Value:  c.EndTime.Format(time.RFC3339Nano),
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	return result
}

// ValidateTimeRange validates the time range of c.
func (c *Constraint) ValidateTimeRange() error {
	if c.StartTime == nil {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Constraint must have an time_start")
	}

	// EndTime cannot be omitted for new Constraints.
	if c.EndTime == nil {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Constraint must have an time_end")
	}

	// EndTime cannot be before StartTime.
	if c.EndTime.Sub(*c.StartTime) < 0 {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Constraint time_end must be after time_start")
	}

	return nil
}
