package models

import (
	"errors"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/interuss/dss/pkg/api/v1/ridpb"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// IdentificationServiceArea represents a USS ISA over a given 4D volume.
type IdentificationServiceArea struct {
	ID         dssmodels.ID
	URL        string
	Owner      dssmodels.Owner
	Cells      s2.CellUnion
	StartTime  *time.Time
	EndTime    *time.Time
	Version    *dssmodels.Version
	AltitudeHi *float32
	AltitudeLo *float32
}

// ToProto converts an IdentificationServiceArea struct to an
// IdentificationServiceArea proto for API consumption.
func (i *IdentificationServiceArea) ToProto() (*ridpb.IdentificationServiceArea, error) {
	result := &ridpb.IdentificationServiceArea{
		Id:         i.ID.String(),
		Owner:      i.Owner.String(),
		FlightsUrl: i.URL,
		Version:    i.Version.String(),
	}

	if i.StartTime != nil {
		ts, err := ptypes.TimestampProto(*i.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = ts
	}

	if i.EndTime != nil {
		ts, err := ptypes.TimestampProto(*i.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd = ts
	}
	return result, nil
}

// SetExtents performs some data validation and sets the 4D volume on the
// IdentificationServiceArea.
func (i *IdentificationServiceArea) SetExtents(extents *ridpb.Volume4D) error {
	var err error
	if extents == nil {
		return nil
	}
	if startTime := extents.GetTimeStart(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return err
		}
		i.StartTime = &ts
	}

	if endTime := extents.GetTimeEnd(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return err
		}
		i.EndTime = &ts
	}

	space := extents.GetSpatialVolume()
	if space == nil {
		return errors.New("missing required spatial_volume")
	}
	i.AltitudeHi = proto.Float32(space.GetAltitudeHi())
	i.AltitudeLo = proto.Float32(space.GetAltitudeLo())
	footprint := space.GetFootprint()
	if footprint == nil {
		return errors.New("spatial_volume missing required footprint")
	}
	i.Cells, err = footprint.ToCommon().CalculateCovering()
	return err
}

// AdjustTimeRange adjusts the time range to the max allowed ranges on a
// IdentificationServiceArea.
func (i *IdentificationServiceArea) AdjustTimeRange(now time.Time, old *IdentificationServiceArea) error {
	if i.StartTime == nil {
		// If StartTime was omitted, default to Now() for new ISAs or re-
		// use the existing time of existing ISAs.
		if old == nil {
			i.StartTime = &now
		} else {
			i.StartTime = old.StartTime
		}
	} else {
		// If setting the StartTime explicitly ensure it is not too far in the past.
		if now.Sub(*i.StartTime) > maxClockSkew {
			return dsserr.BadRequest("IdentificationServiceArea time_start must not be in the past")
		}
	}

	// If EndTime was omitted default to the existing ISA's EndTime.
	if i.EndTime == nil && old != nil {
		i.EndTime = old.EndTime
	}

	// EndTime cannot be omitted for new ISAs.
	if i.EndTime == nil {
		return dsserr.BadRequest("IdentificationServiceArea must have an time_end")
	}

	// EndTime cannot be before StartTime.
	if i.EndTime.Sub(*i.StartTime) < 0 {
		return dsserr.BadRequest("IdentificationServiceArea time_end must be after time_start")
	}

	return nil
}
