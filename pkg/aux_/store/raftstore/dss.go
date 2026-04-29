package raftstore

import (
	"context"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
)

func (r *repo) SaveOwnMetadata(_ context.Context, locality string, publicEndpoint string) error {
	// TODO: implement
	return nil
}

func (r *repo) GetDSSMetadata(_ context.Context) ([]*auxmodels.DSSMetadata, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) RecordHeartbeat(_ context.Context, heartbeat auxmodels.Heartbeat) error {
	// TODO: implement
	return nil
}

func (r *repo) GetDSSAirspaceRepresentationID(_ context.Context) (string, error) {
	// TODO: implement
	return "", nil
}
