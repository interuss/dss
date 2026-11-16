package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/evictionpb"
	"github.com/interuss/dss/pkg/pb/ridpb"
	"github.com/interuss/dss/pkg/rid/operations"
)

func (r *repo) DeleteExpired(ctx context.Context, params dssmodels.EvictionParameters) (map[string]int, error) {
	return r.consensus.HandleWritePayload[map[string]int](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_DeleteExpiredRepoCall{
		DeleteExpiredRepoCall: &evictionpb.DeleteExpiredRepoCall{Parameters: params.ToProto()},
	}}))
}

func (r *repo) GetEvictionConfig(ctx context.Context) (dssmodels.EvictionConfig, error) {
	return r.consensus.HandleReadPayload[dssmodels.EvictionConfig](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_GetEvictionConfigRepoCall{
		GetEvictionConfigRepoCall: &evictionpb.GetConfigRepoCall{},
	}}))
}

func (r *repo) SetEvictionConfig(ctx context.Context, config dssmodels.EvictionConfig) (dssmodels.EvictionConfig, error) {
	return r.consensus.HandleWritePayload[dssmodels.EvictionConfig](ctx, operations.ProposalPayload(&ridpb.Payload{Payload: &ridpb.Payload_SetEvictionConfigRepoCall{
		SetEvictionConfigRepoCall: &evictionpb.SetConfigRepoCall{Config: config.ToProto()},
	}}))
}
