package models

import (
	"errors"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"
)

type IdentificationServiceArea struct {
	ID         ID
	Url        string
	Owner      Owner
	Cells      s2.CellUnion
	StartTime  *time.Time
	EndTime    *time.Time
	Version    *Version
	AltitudeHi *float32
	AltitudeLo *float32
}

func (i *IdentificationServiceArea) ToProto() (*dspb.IdentificationServiceArea, error) {
	result := &dspb.IdentificationServiceArea{
		Id:         i.ID.String(),
		Owner:      i.Owner.String(),
		FlightsUrl: i.Url,
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

func (i *IdentificationServiceArea) SetExtents(extents *dspb.Volume4D) error {
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
	i.Cells, err = geo.GeoPolygonToCellIDs(footprint)
	return err
}

func (s *IdentificationServiceArea) AdjustTimeRange(now time.Time, old *IdentificationServiceArea) error {
	if s.StartTime == nil {
		// If StartTime was omitted, default to Now() for new ISAs or re-
		// use the existing time of existing ISAs.
		if old == nil {
			s.StartTime = &now
		} else {
			s.StartTime = old.StartTime
		}
	} else {
		// If setting the StartTime explicitly ensure it is not too far in the past.
		if now.Sub(*s.StartTime) > maxClockSkew {
			return dsserr.BadRequest("IdentificationServiceArea time_start must not be in the past")
		}
	}

	// If EndTime was omitted default to the existing ISA's EndTime.
	if s.EndTime == nil && old != nil {
		s.EndTime = old.EndTime
	}

	// EndTime cannot be omitted for new ISAs.
	if s.EndTime == nil {
		return dsserr.BadRequest("IdentificationServiceArea must have an time_end")
	}

	// EndTime cannot be before StartTime.
	if s.EndTime.Sub(*s.StartTime) < 0 {
		return dsserr.BadRequest("IdentificationServiceArea time_end must be after time_start")
	}

	return nil
}
