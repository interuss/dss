package raftstore

import (
	"context"
	"strconv"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

type saveOwnMetadataPayload struct {
	Locality       string
	PublicEndpoint string
}

func (r *repo) SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error {
	if locality == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Locality not set")
	}
	if publicEndpoint == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Public endpoint not set")
	}

	_, err := r.consensus.HandleClientRequest(ctx, saveOwnMetadata, saveOwnMetadataPayload{
		Locality:       locality,
		PublicEndpoint: publicEndpoint,
	}, false)
	return err
}

func (r *repo) GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error) {
	result, err := r.consensus.HandleClientRequest(ctx, getDSSMetadata, nil, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose %s", getDSSMetadata)
	}
	if result == nil {
		return nil, nil
	}
	return result.([]*auxmodels.DSSMetadata), nil
}

func (r *repo) RecordHeartbeat(ctx context.Context, heartbeat auxmodels.Heartbeat) error {
	if heartbeat.Locality == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Locality not set")
	}
	if heartbeat.Source == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Source not set")
	}

	if heartbeat.Timestamp == nil {
		now, err := timestamp.RequestTimestampFromContext(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "failed to get request timestamp")
		}

		heartbeat.Timestamp = &now
	}

	if heartbeat.NextHeartbeatExpectedBefore != nil && heartbeat.NextHeartbeatExpectedBefore.Before(*heartbeat.Timestamp) {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Cannot expect the timestamp of the next heartbeat before the timestamp of the new heartbeat")
	}

	_, err := r.consensus.HandleClientRequest(ctx, recordHeartbeat, heartbeat, false)
	return err
}

func (r *repo) GetDSSAirspaceRepresentationID(_ context.Context) (string, error) {
	return strconv.Itoa(int(raftparams.GetClusterID())), nil
}
