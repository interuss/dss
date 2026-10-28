package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
)

func (r *repo) GetUssAvailability(ctx context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	return r.consensus.HandleReadPayload[*scdmodels.UssAvailabilityStatus](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetUssAvailabilityRepoCall{
		GetUssAvailabilityRepoCall: &scdpb.GetUssAvailabilityRepoCall{Uss: id.String()},
	}}))
}

func (r *repo) UpsertUssAvailability(ctx context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	availability, err := ussa.Availability.ToProto()
	if err != nil {
		return nil, err
	}

	return r.consensus.HandleWritePayload[*scdmodels.UssAvailabilityStatus](ctx, operations.ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_UpsertUssAvailabilityRepoCall{
		UpsertUssAvailabilityRepoCall: &scdpb.UpsertUssAvailabilityRepoCall{
			Uss:          ussa.Uss.String(),
			Availability: availability,
			Version:      ussa.Version.ToProto(),
		},
	}}))
}
