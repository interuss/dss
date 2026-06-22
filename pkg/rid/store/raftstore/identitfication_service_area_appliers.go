package raftstore

import (
	"context"
	"encoding/json"

	"github.com/golang/geo/s2"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/stacktrace"
)

type ISATransactionResult struct {
	Ret  *ridmodels.IdentificationServiceArea
	Subs []*ridmodels.Subscription
}

type DeleteISATransactionPayload struct {
	ID      dssmodels.ID
	Owner   dssmodels.Owner
	Version *dssmodels.Version
}

func (r *repo) deleteISATransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*ISATransactionResult, error) {
	var payload DeleteISATransactionPayload
	err := json.Unmarshal(proposal.Value, &payload)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal payload")
	}

	old, err := mem.GetISA(ctx, payload.ID, true)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", payload.ID.String())
	case !payload.Version.Matches(old.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"ISA currently at version %s but client specified %s", old.Version, payload.Version)
	case old.Owner != payload.Owner:
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"ISA owned by %s, but %s attempted to delete", old.Owner, payload.Owner)
	}

	checkpoint := r.memStore.Checkpoint()
	ret, err := mem.DeleteISA(ctx, old)
	if ret == nil || err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error deleting ISA")
	}

	subs, err := mem.UpdateNotificationIdxsInCells(ctx, old.Cells)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error updating notification indices")
	}

	return &ISATransactionResult{Ret: ret, Subs: subs}, nil
}

func (r *repo) insertISATransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*ISATransactionResult, error) {
	var isa *ridmodels.IdentificationServiceArea
	err := json.Unmarshal(proposal.Value, &isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal payload")
	}

	old, err := mem.GetISA(ctx, isa.ID, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	}
	if old != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "ISA %s already exists", isa.ID)
	}

	checkpoint := r.memStore.Checkpoint()
	subs, err := mem.UpdateNotificationIdxsInCells(ctx, isa.Cells)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error updating notification indices")
	}

	ret, err := mem.InsertISA(ctx, isa)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error inserting ISA")
	}

	return &ISATransactionResult{Ret: ret, Subs: subs}, nil
}

func (r *repo) updateISATransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*ISATransactionResult, error) {
	var isa *ridmodels.IdentificationServiceArea
	err := json.Unmarshal(proposal.Value, &isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal payload")
	}

	old, err := mem.GetISA(ctx, isa.ID, true)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", isa.ID)
	case old.Owner != isa.Owner:
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"ISA owned by %s, but %s attempted to modify", old.Owner, isa.Owner)
	case !old.Version.Matches(isa.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"ISA currently at version %s but client specified %s", old.Version, isa.Version)
	}

	if err := isa.AdjustTimeRange(proposal.Timestamp, old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	checkpoint := r.memStore.Checkpoint()
	ret, err := mem.UpdateISA(ctx, isa)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error updating ISA")
	}

	cells := s2.CellUnionFromUnion(old.Cells, isa.Cells)
	geo.Levelify(&cells)
	subs, err := mem.UpdateNotificationIdxsInCells(ctx, cells)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error updating notification indices")
	}

	return &ISATransactionResult{Ret: ret, Subs: subs}, nil
}
