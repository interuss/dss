package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/evictionpb"
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/dss/pkg/scd/operations"
)

func (r *repo) DeleteExpired(ctx context.Context, params dssmodels.EvictionParameters) (map[string]int, error) {
	return r.consensus.HandleWritePayload[map[string]int](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_DeleteExpiredRepoCall{
		DeleteExpiredRepoCall: &evictionpb.DeleteExpiredRepoCall{Parameters: params.ToProto()},
	}}))
}

func (r *repo) GetEvictionState(ctx context.Context) (dssmodels.EvictionState, error) {
	return r.consensus.HandleReadPayload[dssmodels.EvictionState](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetEvictionStateRepoCall{
		GetEvictionStateRepoCall: &evictionpb.GetStateRepoCall{},
	}}))
}

func (r *repo) SetEvictionConfig(ctx context.Context, config dssmodels.EvictionConfig) (dssmodels.EvictionConfig, error) {
	return r.consensus.HandleWritePayload[dssmodels.EvictionConfig](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_SetEvictionConfigRepoCall{
		SetEvictionConfigRepoCall: &evictionpb.SetConfigRepoCall{Config: config.ToProto()},
	}}))
}

func (r *repo) SetEvictionDisabled(ctx context.Context, disabled bool) (dssmodels.EvictionState, error) {
	return r.consensus.HandleWritePayload[dssmodels.EvictionState](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_SetEvictionDisabledRepoCall{
		SetEvictionDisabledRepoCall: &evictionpb.SetDisabledRepoCall{Disabled: disabled},
	}}))
}
