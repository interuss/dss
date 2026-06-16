package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

type getISAPayload struct {
	ID        dssmodels.ID
	ForUpdate bool
}

func (r *repo) GetISA(ctx context.Context, id dssmodels.ID, forUpdate bool) (*ridmodels.IdentificationServiceArea, error) {
	result, err := r.consensus.ProposeValue(ctx, getISA, &getISAPayload{ID: id, ForUpdate: forUpdate}, true)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	isa, ok := result.(*ridmodels.IdentificationServiceArea)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return isa, nil
}

func (r *repo) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	result, err := r.consensus.ProposeValue(ctx, deleteISA, isa, false)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	isa, ok := result.(*ridmodels.IdentificationServiceArea)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return isa, nil
}

func (r *repo) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	result, err := r.consensus.ProposeValue(ctx, insertISA, isa, false)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	isa, ok := result.(*ridmodels.IdentificationServiceArea)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return isa, nil
}

func (r *repo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	result, err := r.consensus.ProposeValue(ctx, updateISA, isa, false)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	isa, ok := result.(*ridmodels.IdentificationServiceArea)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return isa, nil
}

type searchISAsPayload struct {
	Cells    s2.CellUnion
	Earliest *time.Time
	Latest   *time.Time
}

func (r *repo) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	result, err := r.consensus.ProposeValue(ctx, searchISAs, &searchISAsPayload{Cells: cells, Earliest: earliest, Latest: latest}, true)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	isa, ok := result.([]*ridmodels.IdentificationServiceArea)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return isa, nil
}

type listExpiredISAsPayload struct {
	Writer    string
	Threshold time.Time
}

func (r *repo) ListExpiredISAs(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	result, err := r.consensus.ProposeValue(ctx, listExpiredISAs, &listExpiredISAsPayload{Writer: writer, Threshold: threshold}, true)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	isa, ok := result.([]*ridmodels.IdentificationServiceArea)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return isa, nil
}

func (r *repo) CountISAs(ctx context.Context) (int64, error) {
	result, err := r.consensus.ProposeValue(ctx, countISAs, nil, true)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, nil
	}

	count, ok := result.(int64)
	if !ok {
		return 0, stacktrace.NewError("invalid result type: %T", result)
	}

	return count, nil
}
