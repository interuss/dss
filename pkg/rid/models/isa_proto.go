package models

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/ridpb"
	"github.com/interuss/stacktrace"
)

func ISAToProto(isa *IdentificationServiceArea) (*ridpb.IdentificationServiceArea, error) {
	if isa == nil {
		return nil, nil
	}
	id, err := dssmodels.IDToProto(isa.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ISA ID")
	}
	return &ridpb.IdentificationServiceArea{
		Id:     id,
		Url:    isa.URL,
		Owner:  isa.Owner.String(),
		Writer: isa.Writer,
		Volume: dssmodels.CellsVolume4DToProto(isa.CellsVolume4D),
	}, nil
}

func ISAFromProto(pb *ridpb.IdentificationServiceArea, pbVersion string) (*IdentificationServiceArea, error) {
	if pb == nil {
		return nil, nil
	}
	id, err := dssmodels.IDFromProto(pb.GetId())
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ISA ID")
	}
	version, err := dssmodels.VersionFromProto(pbVersion)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ISA version")
	}
	return &IdentificationServiceArea{
		ID:            id,
		URL:           pb.GetUrl(),
		Owner:         dssmodels.Owner(pb.GetOwner()),
		Version:       version,
		Writer:        pb.GetWriter(),
		CellsVolume4D: dssmodels.CellsVolume4DFromProto(pb.GetVolume()),
	}, nil
}
