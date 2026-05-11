package raftstore

import (
	"context"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

// SaveOwnMetadata returns nil instead of dsserr.NotImplemented because it is needed to allow the server to startup.
func (r *repo) SaveOwnMetadata(_ context.Context, locality string, publicEndpoint string) error {
	return nil
}

func (r *repo) GetDSSMetadata(_ context.Context) ([]*auxmodels.DSSMetadata, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSMetadata not implemented for raftstore")
}

func (r *repo) RecordHeartbeat(_ context.Context, heartbeat auxmodels.Heartbeat) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "RecordHeartbeat not implemented for raftstore")
}

func (r *repo) GetDSSAirspaceRepresentationID(_ context.Context) (string, error) {
	return "", stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDSSAirspaceRepresentationID not implemented for raftstore")
}
