package models

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
)

// Constraint models a constraint, sans planar geographic information.
type Constraint struct {
	ID            ID
	Version       Version
	OVN           OVN
	Owner         dssmodels.Owner
	StartTime     *time.Time
	EndTime       *time.Time
	AltitudeLower *float32
	AltitudeUpper *float32
	USSBaseURL    string
}

// ToProto converts the Constraint to its proto API format
func (c *Constraint) ToProto() (*scdpb.ConstraintReference, error) {
	result := &scdpb.ConstraintReference{
		Id:         c.ID.String(),
		Ovn:        c.OVN.String(),
		Owner:      c.Owner.String(),
		Version:    int32(c.Version),
		UssBaseUrl: c.USSBaseURL,
	}

	if c.StartTime != nil {
		ts, err := ptypes.TimestampProto(*c.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = &scdpb.Time{
			Value:  ts,
			Format: dssmodels.TimeFormatRFC3339,
		}
	}

	if c.EndTime != nil {
		ts, err := ptypes.TimestampProto(*c.EndTime)
		if err != nil {
			return nil, err
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
		return dsserr.BadRequest("Constraint must have an time_start")
	}

	// EndTime cannot be omitted for new Constraints.
	if c.EndTime == nil {
		return dsserr.BadRequest("Constraint must have an time_end")
	}

	// EndTime cannot be before StartTime.
	if c.EndTime.Sub(*c.StartTime) < 0 {
		return dsserr.BadRequest("Constraint time_end must be after time_start")
	}

	return nil
}
