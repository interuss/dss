package raftstore

import (
	"context"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

func (r *repo) SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SaveOwnMetadata is not yet supported in raftstore")
}

func (r *repo) GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error) {

	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSMetadata is not yet supported in raftstore")

}

func (r *repo) RecordHeartbeat(ctx context.Context, heartbeat auxmodels.Heartbeat) error {

	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "RecordHeartbeat is not yet supported in raftstore")

}

// GetDSSAirspaceRepresentationID gets the ID of the common DSS Airspace Representation the raftstore represents.
func (r *repo) GetDSSAirspaceRepresentationID(ctx context.Context) (string, error) {
	return "", stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSAirspaceRepresentationID is not yet supported in raftstore")
}
