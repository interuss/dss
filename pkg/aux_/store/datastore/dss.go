package datastore

import (
	"context"
	"time"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/datastore"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

func (r *repo) SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error {

	if locality == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Locality not set")
	}

	if publicEndpoint == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Public endpoint not set")
	}

	var exists bool

	if err := r.QueryRow(ctx, "SELECT EXISTS (SELECT * FROM pool_participants WHERE locality = $1)", locality).Scan(&exists); err != nil {
		return stacktrace.Propagate(err, "Error checking metadata existence")
	}

	if exists {
		q, err := r.Query(ctx, "UPDATE pool_participants SET public_endpoint = $2, updated_at = transaction_timestamp() WHERE locality = $1", locality, publicEndpoint)
		q.Close()

		if err != nil {
			return stacktrace.Propagate(err, "Error updating metadata")
		}

	} else {

		q, err := r.Query(ctx, "INSERT INTO pool_participants (locality, public_endpoint, updated_at) VALUES ($1, $2, transaction_timestamp())", locality, publicEndpoint)
		q.Close()

		if err != nil {
			return stacktrace.Propagate(err, "Error updating metadata")
		}
	}

	return nil

}

func (r *repo) GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error) {

	rows, err := r.Query(ctx, `
        SELECT
            pp.locality, pp.public_endpoint, pp.updated_at, lts.source, lts.timestamp, lts.next_expected_timestamp, lts.reporter
        FROM
            pool_participants pp
        LEFT JOIN (
            SELECT h1.locality, source, timestamp, next_expected_timestamp, reporter
            FROM heartbeats h1
            JOIN (
                SELECT locality, MAX(timestamp) AS max_timestamp
                FROM heartbeats
                GROUP BY locality
            ) h2 ON h1.locality = h2.locality AND h1.timestamp = h2.max_timestamp
        ) lts ON pp.locality = lts.locality;
    `)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting pool participants")
	}
	defer rows.Close()

	var metadata []*auxmodels.DSSMetadata
	for rows.Next() {
		var (
			m = new(auxmodels.DSSMetadata)
		)
		err = rows.Scan(
			&m.Locality,
			&m.PublicEndpoint,
			&m.UpdatedAt,
			&m.LatestTimestamp.Source,
			&m.LatestTimestamp.Timestamp,
			&m.LatestTimestamp.NextHeartbeatExpectedBefore,
			&m.LatestTimestamp.Reporter,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning pool participants row")
		}
		metadata = append(metadata, m)
	}
	if err = rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}

	return metadata, nil

}

func (r *repo) RecordHeartbeat(ctx context.Context, heartbeat auxmodels.Heartbeat) error {

	if heartbeat.Locality == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Locality not set")
	}

	if heartbeat.Source == "" {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Source not set")
	}

	if heartbeat.Timestamp == nil {
		now := time.Now()
		heartbeat.Timestamp = &now
	}

	if heartbeat.NextHeartbeatExpectedBefore != nil && heartbeat.NextHeartbeatExpectedBefore.Before(*heartbeat.Timestamp) {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Cannot expect the timestamp of the next heartbeat before the timestamp of the new heartbeat")
	}

	var exists bool

	if err := r.QueryRow(ctx, "SELECT EXISTS (SELECT * FROM heartbeats WHERE locality = $1 AND source = $2)", heartbeat.Locality, heartbeat.Source).Scan(&exists); err != nil {
		return stacktrace.Propagate(err, "Error checking heartbeat existence")
	}

	if exists {
		q, err := r.Query(ctx, "UPDATE heartbeats SET timestamp = $3, next_expected_timestamp = $4, reporter = $5 WHERE locality = $1 AND source = $2", heartbeat.Locality, heartbeat.Source, heartbeat.Timestamp, heartbeat.NextHeartbeatExpectedBefore, heartbeat.Reporter)
		q.Close()

		if err != nil {
			return stacktrace.Propagate(err, "Error updating heartbeats")
		}

	} else {
		q, err := r.Query(ctx, "INSERT INTO heartbeats (locality, source, timestamp, next_expected_timestamp, reporter) VALUES ($1, $2, $3, $4, $5)", heartbeat.Locality, heartbeat.Source, heartbeat.Timestamp, heartbeat.NextHeartbeatExpectedBefore, heartbeat.Reporter)

		q.Close()

		if err != nil {
			return stacktrace.Propagate(err, "Error updating heartbeats")
		}
	}

	return nil

}

// GetDSSAirspaceRepresentationID gets the ID of the common DSS Airspace Representation the Datastore represents.
func (r *repo) GetDSSAirspaceRepresentationID(ctx context.Context) (string, error) {
	switch r.version.Type {
	case datastore.CockroachDB:
		var darID string
		if err := r.QueryRow(ctx, "SELECT crdb_internal.cluster_id()").Scan(&darID); err != nil {
			return darID, stacktrace.Propagate(err, "Error getting CockroachDB cluster ID")
		}
		return darID, nil
	case datastore.Yugabyte:

		var darID string

		var count string
		if err := r.QueryRow(ctx, "SELECT COUNT(DISTINCT universe_uuid) from yb_servers();").Scan(&count); err != nil {
			return darID, stacktrace.Propagate(err, "Error getting universe_uuid from Yugabyte. Are you using a version >= 2.25.2.0?")
		}

		if count != "1" {
			return darID, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "Found multiple universe_uuid in yugabyte reporting, this configuration is not supported.")
		}

		if err := r.QueryRow(ctx, "SELECT DISTINCT universe_uuid FROM yb_servers() LIMIT 1").Scan(&darID); err != nil {
			return darID, stacktrace.Propagate(err, "Error getting universe_uuid from Yugabyte. Are you using a version >= 2.25.2.0?")
		}
		return darID, nil

	default:
		return "", stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSAirspaceRepresentationID is not yet supported in current Datastore type '%s'", r.version.Type)
	}
}
