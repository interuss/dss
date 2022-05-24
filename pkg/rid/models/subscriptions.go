package models

import (
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"

	"github.com/golang/geo/s2"
	"github.com/interuss/stacktrace"
)

var (
	// maxSubscriptionDuration is the largest allowed interval between StartTime
	// and EndTime.
	maxSubscriptionDuration = time.Hour * 24

	// maxClockSkew is the largest allowed interval between the StartTime of a new
	// subscription and the server's idea of the current time.
	maxClockSkew = time.Minute * 5
)

// Subscription represents a USS subscription over a given 4D volume.
type Subscription struct {
	ID                dssmodels.ID
	URL               string
	NotificationIndex int
	Owner             dssmodels.Owner
	Cells             s2.CellUnion
	StartTime         *time.Time
	EndTime           *time.Time
	Version           *dssmodels.Version
	AltitudeHi        *float32
	AltitudeLo        *float32
	Writer            string
}

// SetCells is a convenience function that accepts an int64 array and converts
// to s2.CellUnion.
// TODO: wrap s2.CellUnion in a custom type that embeds the struct such that
// we can still call its function directly, but also implements scan for sql
// driver.
func (s *Subscription) SetCells(cids []int64) {
	cells := s2.CellUnion{}
	for _, id := range cids {
		cells = append(cells, s2.CellID(id))
	}
	s.Cells = cells
}

// SetExtents performs some data validation and sets the 4D volume on the
// Subscription.
func (s *Subscription) SetExtents(extents *dssmodels.Volume4D) error {
	var err error
	if extents == nil {
		return nil
	}
	s.StartTime = extents.StartTime
	s.EndTime = extents.EndTime
	s.AltitudeHi = extents.SpatialVolume.AltitudeHi
	s.AltitudeLo = extents.SpatialVolume.AltitudeLo
	s.Cells, err = extents.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return stacktrace.Propagate(err, "Error calculating covering for Subscription")
	}
	return nil
}

// AdjustTimeRange adjusts the time range to the max allowed ranges on a
// subscription.
func (s *Subscription) AdjustTimeRange(now time.Time, old *Subscription) error {
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
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription time_start must not be in the past")
		}
	}

	// If EndTime was omitted default to the existing subscription's EndTime.
	if s.EndTime == nil && old != nil {
		s.EndTime = old.EndTime
	}

	// Or if this is a new subscription default to StartTime + 1 day.
	if s.EndTime == nil {
		truncatedEndTime := s.StartTime.Add(maxSubscriptionDuration)
		s.EndTime = &truncatedEndTime
	}

	// EndTime cannot be before StartTime.
	if s.EndTime.Sub(*s.StartTime) < 0 {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription time_end must be after time_start")
	}

	// EndTime cannot be 24 hrs after StartTime
	if s.EndTime.Sub(*s.StartTime) > maxSubscriptionDuration {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription window exceeds 24 hours")
	}

	return nil
}
