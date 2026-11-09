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

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	pbID, err := id.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Subscription ID")
	}

	return r.consensus.HandleReadPayload[*ridmodels.Subscription](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_GetSubscriptionRepoCall{
		GetSubscriptionRepoCall: &ridpb.GetSubscriptionRepoCall{Id: pbID},
	}}))
}

func (r *repo) DeleteSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	pbID, err := sub.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Subscription ID")
	}

	return r.consensus.HandleWritePayload[*ridmodels.Subscription](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_DeleteSubscriptionRepoCall{
		DeleteSubscriptionRepoCall: &ridpb.DeleteSubscriptionRepoCall{Id: pbID, Version: sub.Version.ToProto()},
	}}))
}

func (r *repo) InsertSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	pbSub, err := sub.ToProto()
	if err != nil {
		return nil, err
	}

	return r.consensus.HandleWritePayload[*ridmodels.Subscription](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_InsertSubscriptionRepoCall{
		InsertSubscriptionRepoCall: &ridpb.InsertSubscriptionRepoCall{Subscription: pbSub},
	}}))
}

func (r *repo) UpdateSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	pbSub, err := sub.ToProto()
	if err != nil {
		return nil, err
	}

	return r.consensus.HandleWritePayload[*ridmodels.Subscription](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_UpdateSubscriptionRepoCall{
		UpdateSubscriptionRepoCall: &ridpb.UpdateSubscriptionRepoCall{Subscription: pbSub, Version: sub.Version.ToProto()},
	}}))
}

func (r *repo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	return r.consensus.HandleReadPayload[[]*ridmodels.Subscription](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_SearchSubscriptionsRepoCall{
		SearchSubscriptionsRepoCall: &ridpb.SearchSubscriptionsRepoCall{Cells: dssmodels.CellUnionToProto(cells)},
	}}))
}

func (r *repo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	return r.consensus.HandleReadPayload[[]*ridmodels.Subscription](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_SearchSubscriptionsByOwnerRepoCall{
		SearchSubscriptionsByOwnerRepoCall: &ridpb.SearchSubscriptionsByOwnerRepoCall{Cells: dssmodels.CellUnionToProto(cells), Owner: owner.String()},
	}}))
}

func (r *repo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	return r.consensus.HandleWritePayload[[]*ridmodels.Subscription](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_UpdateNotificationIdxsInCellsRepoCall{
		UpdateNotificationIdxsInCellsRepoCall: &ridpb.UpdateNotificationIdxsInCellsRepoCall{Cells: dssmodels.CellUnionToProto(cells)},
	}}))
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	return r.consensus.HandleReadPayload[int](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_MaxSubscriptionCountInCellsByOwnerRepoCall{
		MaxSubscriptionCountInCellsByOwnerRepoCall: &ridpb.MaxSubscriptionCountInCellsByOwnerRepoCall{Cells: dssmodels.CellUnionToProto(cells), Owner: owner.String()},
	}}))
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, threshold time.Time) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "ListExpiredSubscriptions not implemented for raftstore")
}

func (r *repo) DeleteExpiredSubscriptions(_ context.Context, threshold time.Time, limit int) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpiredSubscriptions not implemented for raftstore")
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadPayload[int64](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_CountSubscriptionsRepoCall{
		CountSubscriptionsRepoCall: &ridpb.CountSubscriptionsRepoCall{},
	}}))
}
