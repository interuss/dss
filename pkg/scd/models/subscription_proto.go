package models

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/stacktrace"
)

func (s *Subscription) ToProto() (*scdpb.Subscription, error) {
	if s == nil {
		return nil, stacktrace.NewError("missing Subscription")
	}
	id, err := s.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Subscription ID")
	}
	return &scdpb.Subscription{
		Id:                          id,
		Manager:                     s.Manager.String(),
		UssBaseUrl:                  s.USSBaseURL,
		NotifyForOperationalIntents: s.NotifyForOperationalIntents,
		NotifyForConstraints:        s.NotifyForConstraints,
		ImplicitSubscription:        s.ImplicitSubscription,
		Volume:                      s.CellsVolume4D.ToProto(),
	}, nil
}

func (s *Subscription) FromProto(pb *scdpb.Subscription) error {
	if pb == nil {
		return stacktrace.NewError("missing Subscription")
	}
	var id dssmodels.ID
	if err := id.FromProto(pb.GetId()); err != nil {
		return stacktrace.Propagate(err, "invalid Subscription ID")
	}
	if pb.GetVolume() == nil {
		return stacktrace.NewError("missing volume of Subscription %s", id)
	}
	*s = Subscription{
		ID:                          id,
		Manager:                     dssmodels.Manager(pb.GetManager()),
		USSBaseURL:                  pb.GetUssBaseUrl(),
		NotifyForOperationalIntents: pb.GetNotifyForOperationalIntents(),
		NotifyForConstraints:        pb.GetNotifyForConstraints(),
		ImplicitSubscription:        pb.GetImplicitSubscription(),
	}
	s.CellsVolume4D.FromProto(pb.GetVolume())
	return nil
}
