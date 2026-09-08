package memstore

import (
	"context"
	"database/sql"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

func (r *repo) SaveOwnMetadata(ctx context.Context, loc string, publicEndpoint string) error {
	now := timestamp.MustFromContext(ctx)

	r.state.Participants[locality(loc)] = &participant{
		PublicEndpoint: publicEndpoint,
		UpdatedAt:      now.UTC(),
	}
	return nil
}

func (r *repo) GetDSSMetadata(_ context.Context) ([]*auxmodels.DSSMetadata, error) {
	metadata := make([]*auxmodels.DSSMetadata, 0, len(r.state.Participants))
	for loc, p := range r.state.Participants {
		updatedAt := p.UpdatedAt
		m := &auxmodels.DSSMetadata{
			Locality:       string(loc),
			PublicEndpoint: p.PublicEndpoint,
			UpdatedAt:      &updatedAt,
		}

		// Find the latest heartbeat across all sources for this locality.
		var latest *heartbeat
		var latestSource string
		for key, hb := range r.state.Heartbeats {
			if key.Locality != loc {
				continue
			}
			if latest == nil || hb.Timestamp.After(*latest.Timestamp) {
				latest = hb
				latestSource = key.Source
			}
		}

		if latest != nil {
			m.LatestTimestamp.Source = sql.NullString{String: latestSource, Valid: true}
			m.LatestTimestamp.Timestamp = latest.Timestamp
			m.LatestTimestamp.NextHeartbeatExpectedBefore = latest.NextHeartbeatExpectedBefore
			m.LatestTimestamp.Reporter = sql.NullString{String: latest.Reporter, Valid: true}
		}

		metadata = append(metadata, m)
	}
	return metadata, nil
}

func (r *repo) RecordHeartbeat(_ context.Context, hb auxmodels.Heartbeat) error {
	r.state.Heartbeats[heartbeatKey{Locality: locality(hb.Locality), Source: hb.Source}] = &heartbeat{
		Timestamp:                   hb.Timestamp,
		NextHeartbeatExpectedBefore: hb.NextHeartbeatExpectedBefore,
		Reporter:                    hb.Reporter,
	}
	return nil
}

func (r *repo) GetDSSAirspaceRepresentationID(_ context.Context) (string, error) {
	return "", stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSAirspaceRepresentationID not implementable for memstore")
}
