package datastore

import (
	"context"

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

		if _, err := r.Query(ctx, "UPDATE pool_participants SET public_endpoint = $2, updated_at = transaction_timestamp() WHERE locality = $1", locality, publicEndpoint); err != nil {
			return stacktrace.Propagate(err, "Error updating metadata")
		}

	} else {

		if _, err := r.Query(ctx, "INSERT INTO pool_participants (locality, public_endpoint, updated_at) VALUES ($1, $2, transaction_timestamp())", locality, publicEndpoint); err != nil {
			return stacktrace.Propagate(err, "Error updating metadata")
		}
	}

	return nil

}

func (r *repo) GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error) {

	rows, err := r.Query(ctx, "SELECT locality, public_endpoint, updated_at FROM pool_participants")
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

// GetDSSAirspaceRepresentationID gets the ID of the common DSS Airspace Representation the Datastore represents.
func (r *repo) GetDSSAirspaceRepresentationID(ctx context.Context) (string, error) {
	if r.version.Type == datastore.CockroachDB {
		var darID string
		if err := r.QueryRow(ctx, "SELECT crdb_internal.cluster_id()").Scan(&darID); err != nil {
			return darID, stacktrace.Propagate(err, "Error getting CockroachDB cluster ID")
		}
		return darID, nil
	} else {
		return "", stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSAirspaceRepresentationID is not yet supported in current Datastore type '%s'", r.version.Type)
	}
}
