package models

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/stacktrace"
)

func (s OperationalIntentState) ToProto() (scdpb.OperationalIntentState, error) {
	switch s {
	case OperationalIntentStateUnknown:
		return scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_UNKNOWN, nil
	case OperationalIntentStateAccepted:
		return scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_ACCEPTED, nil
	case OperationalIntentStateActivated:
		return scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_ACTIVATED, nil
	case OperationalIntentStateNonconforming:
		return scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_NONCONFORMING, nil
	case OperationalIntentStateContingent:
		return scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_CONTINGENT, nil
	default:
		return 0, stacktrace.NewError("invalid OperationalIntent state %q", s)
	}
}

func (s *OperationalIntentState) FromProto(pb scdpb.OperationalIntentState) error {
	switch pb {
	case scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_UNKNOWN:
		*s = OperationalIntentStateUnknown
	case scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_ACCEPTED:
		*s = OperationalIntentStateAccepted
	case scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_ACTIVATED:
		*s = OperationalIntentStateActivated
	case scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_NONCONFORMING:
		*s = OperationalIntentStateNonconforming
	case scdpb.OperationalIntentState_OPERATIONAL_INTENT_STATE_CONTINGENT:
		*s = OperationalIntentStateContingent
	default:
		return stacktrace.NewError("invalid OperationalIntent state %d", pb)
	}
	return nil
}

func (o *OperationalIntent) ToProto() (*scdpb.UpsertOperationalIntentRepoCall_OperationalIntent, error) {
	if o == nil {
		return nil, stacktrace.NewError("missing OperationalIntent")
	}
	id, err := o.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid OperationalIntent ID")
	}
	if o.CellsVolume4D == nil {
		return nil, stacktrace.NewError("missing volume of OperationalIntent %s", o.ID)
	}
	var subscriptionID []byte
	if o.SubscriptionID != nil {
		subscriptionID, err = o.SubscriptionID.ToProto()
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription ID of OperationalIntent %s", o.ID)
		}
	}
	state, err := o.State.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid OperationalIntent state of OperationalIntent %s", o.ID)
	}
	return &scdpb.UpsertOperationalIntentRepoCall_OperationalIntent{
		Id:             id,
		Manager:        o.Manager.String(),
		Version:        int32(o.Version),
		State:          state,
		NewOvn:         o.OVN.ToProto(),
		PastOvns:       OVNsToProto(o.PastOVNs),
		UssBaseUrl:     o.USSBaseURL,
		SubscriptionId: subscriptionID,
		Volume:         o.CellsVolume4D.ToProto(),
	}, nil
}

func (o *OperationalIntent) FromProto(pb *scdpb.UpsertOperationalIntentRepoCall_OperationalIntent) error {
	if pb == nil {
		return stacktrace.NewError("missing OperationalIntent")
	}
	var id dssmodels.ID
	if err := id.FromProto(pb.GetId()); err != nil {
		return stacktrace.Propagate(err, "invalid OperationalIntent ID")
	}
	if pb.GetVolume() == nil {
		return stacktrace.NewError("missing volume of OperationalIntent %s", id)
	}
	var subscriptionID *dssmodels.ID
	if pb.GetSubscriptionId() != nil {
		subscriptionID = new(dssmodels.ID)
		if err := subscriptionID.FromProto(pb.GetSubscriptionId()); err != nil {
			return stacktrace.Propagate(err, "invalid Subscription ID of OperationalIntent %s", id)
		}
	}
	decoded := OperationalIntent{
		ID:             id,
		Manager:        dssmodels.Manager(pb.GetManager()),
		Version:        VersionNumber(pb.GetVersion()),
		USSBaseURL:     pb.GetUssBaseUrl(),
		SubscriptionID: subscriptionID,
		CellsVolume4D:  new(dssmodels.CellsVolume4D),
	}
	if err := decoded.State.FromProto(pb.GetState()); err != nil {
		return err
	}
	if err := decoded.OVN.FromProto(pb.GetNewOvn()); err != nil {
		return err
	}
	pastOVNs, err := OVNsFromProto(pb.GetPastOvns())
	if err != nil {
		return err
	}
	decoded.PastOVNs = pastOVNs
	decoded.CellsVolume4D.FromProto(pb.GetVolume())

	*o = decoded
	return nil
}
