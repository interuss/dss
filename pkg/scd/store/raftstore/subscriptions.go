package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
	"github.com/interuss/stacktrace"
)

const (
	listExpiredSubscriptions consensus.RequestType[[]dssmodels.ID] = "listExpiredSubscriptions"
)

func (r *repo) SearchSubscriptions(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	if cellsVolume == nil {
		return nil, stacktrace.NewError("missing volume")
	}

	return r.consensus.HandleReadPayload[[]*scdmodels.Subscription](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_SearchSubscriptionsRepoCall{
		SearchSubscriptionsRepoCall: &scdpb.SearchSubscriptionsRepoCall{Volume: cellsVolume.ToProto()},
	}}))
}

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	pbID, err := id.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Subscription ID")
	}

	return r.consensus.HandleReadPayload[*scdmodels.Subscription](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetSubscriptionRepoCall{
		GetSubscriptionRepoCall: &scdpb.GetSubscriptionRepoCall{Id: pbID},
	}}))
}

func (r *repo) UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	pbSub, err := sub.ToProto()
	if err != nil {
		return nil, err
	}

	return r.consensus.HandleWritePayload[*scdmodels.Subscription](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_UpsertSubscriptionRepoCall{
		UpsertSubscriptionRepoCall: &scdpb.UpsertSubscriptionRepoCall{Subscription: pbSub, NotificationIndex: int32(sub.NotificationIndex)},
	}}))
}

func (r *repo) DeleteSubscription(ctx context.Context, id dssmodels.ID) error {
	pbID, err := id.ToProto()
	if err != nil {
		return stacktrace.Propagate(err, "invalid Subscription ID")
	}

	_, err = r.consensus.HandleWritePayload[any](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_DeleteSubscriptionRepoCall{
		DeleteSubscriptionRepoCall: &scdpb.DeleteSubscriptionRepoCall{Id: pbID},
	}}))
	return err
}

func (r *repo) IncrementNotificationIndicesForOperationalIntents(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	if cellsVolume == nil {
		return nil, stacktrace.NewError("missing volume")
	}

	return r.consensus.HandleWritePayload[[]*scdmodels.Subscription](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_IncrementNotificationIndicesForOperationalIntentsRepoCall{
		IncrementNotificationIndicesForOperationalIntentsRepoCall: &scdpb.IncrementNotificationIndicesForOperationalIntentsRepoCall{Volume: cellsVolume.ToProto()},
	}}))
}

func (r *repo) IncrementNotificationIndicesForConstraints(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	if cellsVolume == nil {
		return nil, stacktrace.NewError("missing volume")
	}

	return r.consensus.HandleWritePayload[[]*scdmodels.Subscription](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_IncrementNotificationIndicesForConstraintsRepoCall{
		IncrementNotificationIndicesForConstraintsRepoCall: &scdpb.IncrementNotificationIndicesForConstraintsRepoCall{Volume: cellsVolume.ToProto()},
	}}))
}

// LockSubscriptionsOnCells is a no-op in the raftstore implementation since raft requests are executed sequentially
func (r *repo) LockSubscriptionsOnCells(_ context.Context, _ s2.CellUnion, _ []dssmodels.ID, _ *time.Time, _ *time.Time) error {
	return nil
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, threshold time.Time) ([]dssmodels.ID, error) {
	buf, err := json.Marshal(threshold)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, listExpiredSubscriptions, buf)
}

func (r *repo) DeleteExpiredSubscriptions(_ context.Context, threshold time.Time, limit int) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpiredSubscriptions not implemented for raftstore")
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadPayload[int64](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_CountSubscriptionsRepoCall{
		CountSubscriptionsRepoCall: &scdpb.CountSubscriptionsRepoCall{},
	}}))
}
