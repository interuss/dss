package models

import (
	"time"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
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
		return nil
	}
	if wrapper := space.GetAltitudeHi(); wrapper != nil {
		i.AltitudeHi = ptrToFloat32(wrapper.GetValue())
	}
	if wrapper := space.GetAltitudeLo(); wrapper != nil {
		i.AltitudeLo = ptrToFloat32(wrapper.GetValue())
	}
	footprint := space.GetFootprint()
	if footprint == nil {
		return nil
	}
	i.Cells, err = geo.GeoPolygonToCellIDs(footprint)
	return err
}
