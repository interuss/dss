package models

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/ridpb"
	"github.com/interuss/stacktrace"
)

func (sub *Subscription) ToProto() (*ridpb.Subscription, error) {
	if sub == nil {
		return nil, stacktrace.NewError("missing Subscription")
	}
	id, err := sub.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Subscription ID")
	}
	if sub.CellsVolume4D == nil {
		return nil, stacktrace.NewError("missing volume of Subscription %s", sub.ID)
	}
	return &ridpb.Subscription{
		Id:     id,
		Url:    sub.URL,
		Owner:  sub.Owner.String(),
		Writer: sub.Writer,
		Volume: sub.CellsVolume4D.ToProto(),
	}, nil
}

func (sub *Subscription) FromProto(pb *ridpb.Subscription) error {
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
	volume := new(dssmodels.CellsVolume4D)
	volume.FromProto(pb.GetVolume())

	*sub = Subscription{
		ID:            id,
		URL:           pb.GetUrl(),
		Owner:         dssmodels.Owner(pb.GetOwner()),
		Writer:        pb.GetWriter(),
		CellsVolume4D: volume,
	}
	return nil
}
