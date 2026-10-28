package models

import (
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/stacktrace"
)

func (s UssAvailabilityState) ToProto() (scdpb.UssAvailabilityState, error) {
	switch s {
	case UssAvailabilityStateUnknown:
		return scdpb.UssAvailabilityState_USS_AVAILABILITY_STATE_UNKNOWN, nil
	case UssAvailabilityStateNormal:
		return scdpb.UssAvailabilityState_USS_AVAILABILITY_STATE_NORMAL, nil
	case UssAvailabilityStateDown:
		return scdpb.UssAvailabilityState_USS_AVAILABILITY_STATE_DOWN, nil
	default:
		return 0, stacktrace.NewError("invalid USS availability state %q", s)
	}
}

func (s *UssAvailabilityState) FromProto(pb scdpb.UssAvailabilityState) error {
	switch pb {
	case scdpb.UssAvailabilityState_USS_AVAILABILITY_STATE_UNKNOWN:
		*s = UssAvailabilityStateUnknown
	case scdpb.UssAvailabilityState_USS_AVAILABILITY_STATE_NORMAL:
		*s = UssAvailabilityStateNormal
	case scdpb.UssAvailabilityState_USS_AVAILABILITY_STATE_DOWN:
		*s = UssAvailabilityStateDown
	default:
		return stacktrace.NewError("invalid USS availability state %d", pb)
	}
	return nil
}
