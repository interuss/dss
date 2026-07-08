package actions

import (
	"context"
	"time"

	restapi "github.com/interuss/dss/pkg/api/auxv1"
	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/aux_/repos"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/locality"
	"github.com/interuss/stacktrace"
)

func GetDSSInstances(ctx context.Context, r repos.Repository, _ *restapi.GetDSSInstancesRequest) (any, error) {
	metadata, err := r.GetDSSMetadata(ctx)
	if err != nil {
		return nil, err
	}

	instances := make([]restapi.DSSInstance, len(metadata))
	for index, instanceMetadata := range metadata {
		instances[index] = restapi.DSSInstance{
			Id:             instanceMetadata.Locality,
			PublicEndpoint: &instanceMetadata.PublicEndpoint,
		}

		if instanceMetadata.LatestTimestamp.Source.Valid {
			instances[index].MostRecentHeartbeat = &restapi.Heartbeat{
				Timestamp: instanceMetadata.LatestTimestamp.Timestamp.Format(time.RFC3339Nano),
				Reporter:  &instanceMetadata.LatestTimestamp.Reporter.String,
				Source:    instanceMetadata.LatestTimestamp.Source.String,
			}

			if instanceMetadata.LatestTimestamp.NextHeartbeatExpectedBefore != nil {
				nextExpectedTimestamp := instanceMetadata.LatestTimestamp.NextHeartbeatExpectedBefore.Format(time.RFC3339Nano)
				instances[index].MostRecentHeartbeat.NextHeartbeatExpectedBefore = &nextExpectedTimestamp
			}
		}
	}

	return &restapi.DSSInstancesResponse{DssInstances: &instances}, nil
}

func PutDSSInstancesHeartbeat(ctx context.Context, r repos.Repository, a *restapi.PutDSSInstancesHeartbeatRequest) (any, error) {
	if a.Source == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Source not set")
	}

	locality, err := locality.LocalityFromContext(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get request locality")
	}

	heartbeat := auxmodels.Heartbeat{
		Source:   *a.Source,
		Reporter: *a.Auth.ClientID,
		Locality: locality,
	}

	if a.Timestamp != nil {
		ts, err := time.Parse(time.RFC3339Nano, *a.Timestamp)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Unable to parse timestamp as RFC3339 time")
		}
		heartbeat.Timestamp = &ts
	}

	if a.NextHeartbeatExpectedBefore != nil {
		ts, err := time.Parse(time.RFC3339Nano, *a.NextHeartbeatExpectedBefore)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Unable to parse next heartbeat expected before as RFC3339 time")
		}
		heartbeat.NextHeartbeatExpectedBefore = &ts
	}

	return nil, r.RecordHeartbeat(ctx, heartbeat)
}
