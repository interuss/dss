package models

import (
	"time"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dssmodels "github.com/interuss/dss/pkg/models"
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
		Version:         restapi.EntityVersion(c.Version),
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
