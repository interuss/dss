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

// ToProto converts the Constraint to its proto API format
func (c *Constraint) ToProto() (*scdpb.ConstraintReference, error) {
	result := &scdpb.ConstraintReference{
		Id:              c.ID.String(),
		Ovn:             c.OVN.String(),
		Manager:         c.Manager.String(),
		Version:         int32(c.Version),
		UssBaseUrl:      c.USSBaseURL,
		UssAvailability: UssAvailabilityStateUnknown.String(),
	}

	if c.StartTime != nil {
		ts, err := ptypes.TimestampProto(*c.StartTime)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting start time to proto")
		}
		result.TimeStart = &scdpb.Time{
			Value:  ts,
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	if c.EndTime != nil {
		ts, err := ptypes.TimestampProto(*c.EndTime)
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
