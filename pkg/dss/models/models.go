package models

import (
	"errors"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	"github.com/steeling/InterUSS-Platform/pkg/dssv1"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"
)

const (
	// Convert updatedAt to a string, why not make it smaller
	// WARNING: Changing this will cause RMW errors
	// 32 is the highest value allowed by strconv
	versionBase = 32
)

type (
	ID      string
	Owner   string
	Version struct {
		t time.Time
		s string
	}
)

type S2Model struct {
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

func (s *S2Model) SetExtents(extents *dssv1.Volume4D) error {
	var err error
	if extents == nil {
		return nil
	}
	if startTime := extents.GetTimeStart(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return err
		}
		s.StartTime = &ts
	}

	if endTime := extents.GetTimeEnd(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return err
		}
		s.EndTime = &ts
	}

	space := extents.GetSpatialVolume()
	if space == nil {
		return errors.New("missing required spatial_volume")
	}
	s.AltitudeHi = proto.Float32(space.GetAltitudeHi())
	s.AltitudeLo = proto.Float32(space.GetAltitudeLo())
	footprint := space.GetFootprint()
	if footprint == nil {
		return errors.New("spatial_volume missing required footprint")
	}
	s.Cells, err = geo.GeoPolygonToCellIDs(footprint)
	return err
}

func (s *S2Model) AdjustTimeRange(now time.Time, old *Subscription) error {
	if s.StartTime == nil {
		// If StartTime was omitted, default to Now() for new subscriptions or re-
		// use the existing time of existing subscriptions.
		if old == nil {
			s.StartTime = &now
		} else {
			s.StartTime = old.StartTime
		}
	} else {
		// If setting the StartTime explicitly ensure it is not too far in the past.
		if now.Sub(*s.StartTime) > maxClockSkew {
			return dsserr.BadRequest("subscription time_start must not be in the past")
		}
	}

	// If EndTime was omitted default to the existing subscription's EndTime.
	if s.EndTime == nil && old != nil {
		s.EndTime = old.EndTime
	}

	// Or if this is a new subscription default to StartTime + 1 day.  Also
	// truncate long existing subscriptions to 1 day.
	if s.EndTime == nil || s.EndTime.Sub(*s.StartTime) > maxSubscriptionDuration {
		truncatedEndTime := s.StartTime.Add(maxSubscriptionDuration)
		s.EndTime = &truncatedEndTime
	}

	// EndTime cannot be before StartTime.
	if s.EndTime.Sub(*s.StartTime) < 0 {
		return dsserr.BadRequest("subscription time_end must be after time_start")
	}

	return nil
}

func (id ID) String() string {
	return string(id)
}

func (owner Owner) String() string {
	return string(owner)
}

func VersionFromString(s string) (*Version, error) {
	v := &Version{s: s}
	if s == "" {
		return nil, nil
	}
	nanos, err := strconv.ParseUint(string(s), versionBase, 64)
	if err != nil {
		return nil, err
	}
	v.t = time.Unix(0, int64(nanos))
	return v, nil
}

func VersionFromTime(t time.Time) *Version {
	return &Version{
		t: t,
		s: strconv.FormatUint(uint64(t.UnixNano()), versionBase),
	}
}

func (v *Version) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	temp := VersionFromTime(src.(time.Time))
	*v = *temp
	return nil
}

func (v *Version) Empty() bool {
	return v == nil
}

func (v *Version) Matches(v2 *Version) bool {
	if v == nil || v2 == nil {
		return false
	}
	return v.s == v2.s
}

func (v *Version) String() string {
	if v == nil {
		return ""
	}
	return v.s
}

func (v *Version) ToTimestamp() time.Time {
	return v.t
}

func ptrToFloat32(f float32) *float32 {
	return &f
}
