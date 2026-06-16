package memstore

import (
	"context"
	"database/sql"
	"time"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

func (r *repo) SaveOwnMetadata(_ context.Context, locality string, publicEndpoint string) error {
	if locality == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Locality not set")
	}
	if publicEndpoint == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Public endpoint not set")
	}

	r.state.Participants[locality] = &participant{
		PublicEndpoint: publicEndpoint,
		UpdatedAt:      time.Now().UTC(),
	}
	return nil
}

func (r *repo) GetDSSMetadata(_ context.Context) ([]*auxmodels.DSSMetadata, error) {
	metadata := make([]*auxmodels.DSSMetadata, 0, len(r.state.Participants))
	for locality, p := range r.state.Participants {
		updatedAt := p.UpdatedAt
		m := &auxmodels.DSSMetadata{
			Locality:       locality,
			PublicEndpoint: p.PublicEndpoint,
			UpdatedAt:      &updatedAt,
		}

		// Find the latest heartbeat across all sources for this locality.
		var latest auxmodels.Heartbeat
		found := false
		for key, hb := range r.state.Heartbeats {
			if key.Locality != locality {
				continue
			}
			if !found || hb.Timestamp.After(*latest.Timestamp) {
				latest = hb
				found = true
			}
		}

		if found {
			m.LatestTimestamp.Source = sql.NullString{String: latest.Source, Valid: true}
			m.LatestTimestamp.Timestamp = latest.Timestamp
			m.LatestTimestamp.NextHeartbeatExpectedBefore = latest.NextHeartbeatExpectedBefore
			m.LatestTimestamp.Reporter = sql.NullString{String: latest.Reporter, Valid: true}
		}

		metadata = append(metadata, m)
	}
	return metadata, nil
}

func (r *repo) RecordHeartbeat(_ context.Context, heartbeat auxmodels.Heartbeat) error {
	if heartbeat.Locality == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Locality not set")
	}
	if heartbeat.Source == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Source not set")
	}

	if heartbeat.Timestamp == nil {
		now := time.Now().UTC()
		heartbeat.Timestamp = &now
	}

	if heartbeat.NextHeartbeatExpectedBefore != nil && heartbeat.NextHeartbeatExpectedBefore.Before(*heartbeat.Timestamp) {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Cannot expect the timestamp of the next heartbeat before the timestamp of the new heartbeat")
	}

	r.state.Heartbeats[heartbeatKey{Locality: heartbeat.Locality, Source: heartbeat.Source}] = heartbeat
	return nil
}

func (r *repo) GetDSSAirspaceRepresentationID(_ context.Context) (string, error) {
	return "", stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSAirspaceRepresentationID not implementable for memstore")
}
