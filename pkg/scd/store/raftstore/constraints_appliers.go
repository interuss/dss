package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

type UpsertConstraintTransactionPayload struct {
	Manager    dssmodels.Manager
	ID         dssmodels.ID
	Ovn        scdmodels.OVN
	USSBaseURL string
	StartTime  *time.Time
	EndTime    *time.Time
	AltitudeLo *float32
	AltitudeHi *float32
	Cells      s2.CellUnion
}

func (r *repo) deleteConstraintTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.ChangeConstraintReferenceResponse, error) {
	var req *restapi.DeleteConstraintReferenceRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal delete constraint reference request")
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	ovn := scdmodels.OVN(req.Ovn)
	if ovn == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing OVN for constraint to modify")
	}

	if req.Auth.ClientID == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager")
	}

	old, err := mem.GetConstraint(ctx, id)
	switch {
	case err == pgx.ErrNoRows:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id)
	case err != nil:
		return nil, stacktrace.Propagate(err, "Unable to get Constraint from repo")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id)
	case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"Constraint owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	case old.OVN != ovn:
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"Current version is %s but client specified version %s", old.OVN, ovn)
	}

	notifyVolume := &dssmodels.Volume4D{
		StartTime: old.StartTime,
		EndTime:   old.EndTime,
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeHi: old.AltitudeUpper,
			AltitudeLo: old.AltitudeLower,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return old.Cells, nil
			}),
		},
	}

	cp := r.memStore.Checkpoint()

	if err := mem.DeleteConstraint(ctx, id); err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.Propagate(err, "Unable to delete Constraint from repo")
	}

	subs, err := mem.IncrementNotificationIndicesForConstraints(ctx, notifyVolume)
	if err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.Propagate(err, "Unable to increment notification indices")
	}

	return &restapi.ChangeConstraintReferenceResponse{
		ConstraintReference: *old.ToRest(),
		Subscribers:         repos.MakeSubscribersToNotify(subs),
	}, nil
}

func (r *repo) getConstraintTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.GetConstraintReferenceResponse, error) {
	var req *restapi.GetConstraintReferenceRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal get constraint reference request")
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	if req.Auth.ClientID == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager")
	}

	constraint, err := mem.GetConstraint(ctx, id)
	if err == pgx.ErrNoRows || constraint == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id)
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get Constraint from repo")
	}

	if constraint.Manager != dssmodels.Manager(*req.Auth.ClientID) {
		constraint.OVN = scdmodels.NoOvnPhrase
	}

	return &restapi.GetConstraintReferenceResponse{
		ConstraintReference: *constraint.ToRest(),
	}, nil
}

func (r *repo) queryConstraintTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.QueryConstraintReferencesResponse, error) {
	var req *restapi.QueryConstraintReferencesRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal query constraint references request")
	}

	if req.Auth.ClientID == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager")
	}

	vol4, err := dssmodels.Volume4DFromSCDRest(req.Body.AreaOfInterest)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry")
	}

	constraints, err := mem.SearchConstraints(ctx, vol4)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to query for Constraints in repo")
	}

	response := &restapi.QueryConstraintReferencesResponse{
		ConstraintReferences: make([]restapi.ConstraintReference, 0, len(constraints)),
	}
	for _, constraint := range constraints {
		p := constraint.ToRest()
		if constraint.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
			p.Ovn = &noOvnPhrase
		}
		response.ConstraintReferences = append(response.ConstraintReferences, *p)
	}

	return response, nil
}

func (r *repo) upsertConstraintTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.ChangeConstraintReferenceResponse, error) {
	var payload *UpsertConstraintTransactionPayload
	if err := json.Unmarshal(proposal.Value, &payload); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal upsert constraint request")
	}

	uExtent := &dssmodels.Volume4D{
		StartTime: payload.StartTime,
		EndTime:   payload.EndTime,
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeHi: payload.AltitudeHi,
			AltitudeLo: payload.AltitudeLo,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return payload.Cells, nil
			}),
		},
	}

	old, err := mem.GetConstraint(ctx, payload.ID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, stacktrace.Propagate(err, "Could not get Constraint from repo")
	}

	version := scdmodels.VersionNumber(1)
	if old == nil {
		if payload.Ovn != "" {
			return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Old version %s does not exist", payload.Ovn)
		}
	} else {
		if old.Manager != payload.Manager {
			return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"Constraint owned by %s, but %s attempted to modify", old.Manager, payload.Manager)
		}
		if old.OVN != payload.Ovn {
			return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"Current version is %s but client specified version %s", old.OVN, payload.Ovn)
		}
		version = old.Version + 1
	}

	var notifyVol4 *dssmodels.Volume4D
	if old == nil {
		notifyVol4 = uExtent
	} else {
		oldVol4 := &dssmodels.Volume4D{
			StartTime: old.StartTime,
			EndTime:   old.EndTime,
			SpatialVolume: &dssmodels.Volume3D{
				AltitudeHi: old.AltitudeUpper,
				AltitudeLo: old.AltitudeLower,
				Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
					return old.Cells, nil
				}),
			},
		}
		notifyVol4, err = dssmodels.UnionVolumes4D(uExtent, oldVol4)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error constructing 4D volumes union")
		}
	}

	constraint := &scdmodels.Constraint{
		ID:            payload.ID,
		Manager:       payload.Manager,
		Version:       version,
		StartTime:     uExtent.StartTime,
		EndTime:       uExtent.EndTime,
		AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
		AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,
		USSBaseURL:    payload.USSBaseURL,
		Cells:         payload.Cells,
	}

	cp := r.memStore.Checkpoint()

	constraint, err = mem.UpsertConstraint(ctx, constraint)
	if err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.Propagate(err, "Failed to upsert Constraint in repo")
	}

	subs, err := mem.IncrementNotificationIndicesForConstraints(ctx, notifyVol4)
	if err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.Propagate(err, "Unable to increment notification indices")
	}

	return &restapi.ChangeConstraintReferenceResponse{
		ConstraintReference: *constraint.ToRest(),
		Subscribers:         repos.MakeSubscribersToNotify(subs),
	}, nil
}
