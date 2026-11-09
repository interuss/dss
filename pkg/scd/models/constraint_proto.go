package models

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/stacktrace"
)

func (c *Constraint) ToProto() (*scdpb.Constraint, error) {
	if c == nil {
		return nil, stacktrace.NewError("missing Constraint")
	}
	id, err := c.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Constraint ID")
	}
	if c.CellsVolume4D == nil {
		return nil, stacktrace.NewError("missing volume of Constraint %s", c.ID)
	}
	return &scdpb.Constraint{
		Id:         id,
		Manager:    c.Manager.String(),
		UssBaseUrl: c.USSBaseURL,
		Volume:     c.CellsVolume4D.ToProto(),
	}, nil
}

func (c *Constraint) FromProto(pb *scdpb.Constraint) error {
	if pb == nil {
		return stacktrace.NewError("missing Constraint")
	}
	var id dssmodels.ID
	if err := id.FromProto(pb.GetId()); err != nil {
		return stacktrace.Propagate(err, "invalid Constraint ID")
	}
	if pb.GetVolume() == nil {
		return stacktrace.NewError("missing volume of Constraint %s", id)
	}
	volume := new(dssmodels.CellsVolume4D)
	volume.FromProto(pb.GetVolume())

	*c = Constraint{
		ID:            id,
		Manager:       dssmodels.Manager(pb.GetManager()),
		USSBaseURL:    pb.GetUssBaseUrl(),
		CellsVolume4D: volume,
	}
	return nil
}
