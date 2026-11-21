package models

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/ridpb"
	"github.com/interuss/stacktrace"
)

func (isa *IdentificationServiceArea) ToProto() (*ridpb.IdentificationServiceArea, error) {
	if isa == nil {
		return nil, stacktrace.NewError("missing ISA")
	}
	id, err := isa.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ISA ID")
	}
	return &ridpb.IdentificationServiceArea{
		Id:     id,
		Url:    isa.URL,
		Owner:  isa.Owner.String(),
		Writer: isa.Writer,
		Volume: isa.CellsVolume4D.ToProto(),
	}, nil
}

func (isa *IdentificationServiceArea) FromProto(pb *ridpb.IdentificationServiceArea) error {
	if pb == nil {
		return stacktrace.NewError("missing ISA")
	}
	var id dssmodels.ID
	if err := id.FromProto(pb.GetId()); err != nil {
		return stacktrace.Propagate(err, "invalid ISA ID")
	}
	if pb.GetVolume() == nil {
		return stacktrace.NewError("missing volume of ISA %s", id)
	}
	*isa = IdentificationServiceArea{
		ID:     id,
		URL:    pb.GetUrl(),
		Owner:  dssmodels.Owner(pb.GetOwner()),
		Writer: pb.GetWriter(),
	}
	isa.CellsVolume4D.FromProto(pb.GetVolume())
	return nil
}
