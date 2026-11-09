package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
	"github.com/interuss/stacktrace"
)

func (r *repo) SearchConstraints(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Constraint, error) {
	if cellsVolume == nil {
		return nil, stacktrace.NewError("missing volume")
	}

	return r.consensus.HandleReadPayload[[]*scdmodels.Constraint](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_SearchConstraintsRepoCall{
		SearchConstraintsRepoCall: &scdpb.SearchConstraintsRepoCall{Volume: cellsVolume.ToProto()},
	}}))
}

func (r *repo) GetConstraint(ctx context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	pbID, err := id.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Constraint ID")
	}

	return r.consensus.HandleReadPayload[*scdmodels.Constraint](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetConstraintRepoCall{
		GetConstraintRepoCall: &scdpb.GetConstraintRepoCall{Id: pbID},
	}}))
}

func (r *repo) UpsertConstraint(ctx context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	pbConstraint, err := constraint.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode Constraint")
	}

	return r.consensus.HandleWritePayload[*scdmodels.Constraint](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_UpsertConstraintRepoCall{
		UpsertConstraintRepoCall: &scdpb.UpsertConstraintRepoCall{Constraint: pbConstraint, Version: int32(constraint.Version)},
	}}))
}

func (r *repo) DeleteConstraint(ctx context.Context, id dssmodels.ID) error {
	pbID, err := id.ToProto()
	if err != nil {
		return stacktrace.Propagate(err, "invalid Constraint ID")
	}

	_, err = r.consensus.HandleWritePayload[any](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_DeleteConstraintRepoCall{
		DeleteConstraintRepoCall: &scdpb.DeleteConstraintRepoCall{Id: pbID},
	}}))
	return err
}

func (r *repo) CountConstraints(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadPayload[int64](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_CountConstraintsRepoCall{
		CountConstraintsRepoCall: &scdpb.CountConstraintsRepoCall{},
	}}))
}
