package raftstore

import (
	"context"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetOperationalIntent(ctx context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	pbID, err := id.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid OperationalIntent ID")
	}

	return r.consensus.HandleReadPayload[*scdmodels.OperationalIntent](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetOperationalIntentRepoCall{
		GetOperationalIntentRepoCall: &scdpb.GetOperationalIntentRepoCall{Id: pbID},
	}}))
}

func (r *repo) DeleteOperationalIntent(ctx context.Context, id dssmodels.ID) error {
	pbID, err := id.ToProto()
	if err != nil {
		return stacktrace.Propagate(err, "invalid OperationalIntent ID")
	}

	_, err = r.consensus.HandleWritePayload[any](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_DeleteOperationalIntentRepoCall{
		DeleteOperationalIntentRepoCall: &scdpb.DeleteOperationalIntentRepoCall{Id: pbID},
	}}))
	return err
}

func (r *repo) UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	pbOperation, err := operation.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode OperationalIntent")
	}

	return r.consensus.HandleWritePayload[*scdmodels.OperationalIntent](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_UpsertOperationalIntentRepoCall{
		UpsertOperationalIntentRepoCall: &scdpb.UpsertOperationalIntentRepoCall{OperationalIntent: pbOperation},
	}}))
}

func (r *repo) SearchOperationalIntents(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.OperationalIntent, error) {
	if cellsVolume == nil {
		return nil, stacktrace.NewError("missing volume")
	}

	return r.consensus.HandleReadPayload[[]*scdmodels.OperationalIntent](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_SearchOperationalIntentsRepoCall{
		SearchOperationalIntentsRepoCall: &scdpb.SearchOperationalIntentsRepoCall{Volume: cellsVolume.ToProto()},
	}}))
}

func (r *repo) GetDependentOperationalIntents(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	pbID, err := subscriptionID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Subscription ID")
	}

	return r.consensus.HandleReadPayload[[]dssmodels.ID](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetDependentOperationalIntentsRepoCall{
		GetDependentOperationalIntentsRepoCall: &scdpb.GetDependentOperationalIntentsRepoCall{SubscriptionId: pbID},
	}}))
}

func (r *repo) ListExpiredOperationalIntents(_ context.Context, threshold time.Time) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "ListExpiredOperationalIntents not implemented for raftstore")
}

func (r *repo) DeleteExpiredOperationalIntents(_ context.Context, threshold time.Time, limit int) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpiredOperationalIntents not implemented for raftstore")
}

func (r *repo) CountOperationalIntents(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadPayload[int64](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_CountOperationalIntentsRepoCall{
		CountOperationalIntentsRepoCall: &scdpb.CountOperationalIntentsRepoCall{},
	}}))
}
