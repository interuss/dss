package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/ridpb"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/operations"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetISA(ctx context.Context, id dssmodels.ID, _ bool) (*ridmodels.IdentificationServiceArea, error) {
	// We ignore the forUpdate parameter because raft requests are executed sequentially, so there is no need to lock the ISA for update.

	pbID, err := dssmodels.IDToProto(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ISA ID")
	}

	return r.consensus.HandleReadPayload[*ridmodels.IdentificationServiceArea](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_GetIsaRepoCall{
		GetIsaRepoCall: &ridpb.GetISARepoCall{Id: pbID},
	}}))
}

func (r *repo) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	pbID, err := dssmodels.IDToProto(isa.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ISA ID")
	}

	return r.consensus.HandleWritePayload[*ridmodels.IdentificationServiceArea](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_DeleteIsaRepoCall{
		DeleteIsaRepoCall: &ridpb.DeleteISARepoCall{Id: pbID, Version: dssmodels.VersionToProto(isa.Version)},
	}}))
}

func (r *repo) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	pbISA, err := ridmodels.ISAToProto(isa)
	if err != nil {
		return nil, err
	}

	return r.consensus.HandleWritePayload[*ridmodels.IdentificationServiceArea](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_InsertIsaRepoCall{
		InsertIsaRepoCall: &ridpb.InsertISARepoCall{Isa: pbISA},
	}}))
}

func (r *repo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	pbISA, err := ridmodels.ISAToProto(isa)
	if err != nil {
		return nil, err
	}

	return r.consensus.HandleWritePayload[*ridmodels.IdentificationServiceArea](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_UpdateIsaRepoCall{
		UpdateIsaRepoCall: &ridpb.UpdateISARepoCall{Isa: pbISA, Version: dssmodels.VersionToProto(isa.Version)},
	}}))
}

func (r *repo) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	return r.consensus.HandleReadPayload[[]*ridmodels.IdentificationServiceArea](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_SearchIsasRepoCall{
		SearchIsasRepoCall: &ridpb.SearchISAsRepoCall{
			Cells:    dssmodels.CellUnionToProto(cells),
			Earliest: dssmodels.TimeToProto(earliest),
			Latest:   dssmodels.TimeToProto(latest),
		},
	}}))
}

func (r *repo) ListExpiredISAs(_ context.Context, threshold time.Time) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "ListExpiredISAs not implemented for raftstore")
}

func (r *repo) DeleteExpiredISAs(_ context.Context, threshold time.Time, limit int) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpiredISAs not implemented for raftstore")
}

func (r *repo) CountISAs(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadPayload[int64](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_CountIsasRepoCall{
		CountIsasRepoCall: &ridpb.CountISAsRepoCall{},
	}}))
}
