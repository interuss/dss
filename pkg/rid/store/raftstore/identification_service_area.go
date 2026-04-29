package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

func (r *repo) GetISA(_ context.Context, id dssmodels.ID, forUpdate bool) (*ridmodels.IdentificationServiceArea, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) DeleteISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) InsertISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) UpdateISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) SearchISAs(_ context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) ListExpiredISAs(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	// TODO: implement
	return nil, nil
}
