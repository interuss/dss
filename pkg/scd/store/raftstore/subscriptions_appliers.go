package raftstore

import (
	"context"
	"encoding/json"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
)

type UpsertSubscriptionTransactionPayload struct {
	Subreq  *scdmodels.Subscription `json:"subreq"`
	Extents *dssmodels.Volume4D     `json:"extents"`
}

func (r *repo) upsertSubscriptionTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.PutSubscriptionResponse, error) {
	var payload *UpsertSubscriptionTransactionPayload
	if err := json.Unmarshal(proposal.Value, &payload); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal upsert subscription transaction payload")
	}

	old, err := mem.GetSubscription(ctx, payload.Subreq.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get existing subscription from repo")
	}

	if err := payload.Subreq.AdjustTimeRange(proposal.Timestamp, old); err != nil {
		return nil, stacktrace.Propagate(err, "failed to adjust subscription time range")
	}

	var dependentOpIds []dssmodels.ID

	if old == nil {
		if payload.Subreq.Version.String() != "" {
			return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", payload.Subreq.ID.String())
		}
	} else {
		switch {
		case payload.Subreq.Version.String() == "":
			return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", payload.Subreq.ID.String())
		case payload.Subreq.Version.String() != old.Version.String():
			return nil, stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", payload.Subreq.Version),
				"Current version is %s but client specified version %s", old.Version, payload.Subreq.Version)
		case old.Manager != payload.Subreq.Manager:
			return nil, stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
				"Subscription owned by %s, but %s attempted to modify", old.Manager, payload.Subreq.Manager)
		}

		payload.Subreq.NotificationIndex = old.NotificationIndex

		dependentOpIds, err = mem.GetDependentOperationalIntents(ctx, payload.Subreq.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not find dependent Operation Ids")
		}

		var operations []*scdmodels.OperationalIntent
		for _, opID := range dependentOpIds {
			operation, err := mem.GetOperationalIntent(ctx, opID)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Could not retrieve dependent Operation %s", opID)
			}
			operations = append(operations, operation)
		}

		if err := payload.Subreq.ValidateDependentOps(operations); err != nil {
			return nil, err
		}
	}

	cp := r.memStore.Checkpoint()

	sub, err := mem.UpsertSubscription(ctx, payload.Subreq)
	if err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.Propagate(err, "Failed to upsert Subscription in repo")
	}
	if sub == nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.NewError("UpsertSubscription returned no Subscription for ID: %s", payload.Subreq.ID)
	}

	var relevantOperations []*scdmodels.OperationalIntent
	if len(sub.Cells) > 0 {
		ops, err := mem.SearchOperationalIntents(ctx, &dssmodels.Volume4D{
			StartTime: sub.StartTime,
			EndTime:   sub.EndTime,
			SpatialVolume: &dssmodels.Volume3D{
				AltitudeLo: sub.AltitudeLo,
				AltitudeHi: sub.AltitudeHi,
				Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
					return sub.Cells, nil
				}),
			},
		})
		if err != nil {
			if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
				return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
			}
			return nil, stacktrace.Propagate(err, "Could not search Operations in repo")
		}
		relevantOperations = ops
	}

	p, err := sub.ToRest(dependentOpIds)
	if err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.Propagate(err, "Could not convert Subscription to REST model")
	}

	result := &restapi.PutSubscriptionResponse{
		Subscription: *p,
	}

	if sub.NotifyForOperationalIntents {
		opIntentRefs := make([]restapi.OperationalIntentReference, 0, len(relevantOperations))
		for _, op := range relevantOperations {
			if op.Manager != dssmodels.Manager(payload.Subreq.Manager) {
				op.OVN = scdmodels.NoOvnPhrase
			}

			opIntentRefs = append(opIntentRefs, *op.ToRest())
		}
		result.OperationalIntentReferences = &opIntentRefs
	}

	if sub.NotifyForConstraints {
		constraints, err := mem.SearchConstraints(ctx, payload.Extents)
		if err != nil {
			if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
				return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
			}
			return nil, stacktrace.Propagate(err, "Could not search Constraints in repo")
		}

		constraintRefs := make([]restapi.ConstraintReference, 0, len(constraints))
		for _, constraint := range constraints {
			p := constraint.ToRest()
			if constraint.Manager != dssmodels.Manager(payload.Subreq.Manager) {
				noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
				p.Ovn = &noOvnPhrase
			}

			constraintRefs = append(constraintRefs, *p)
		}
		result.ConstraintReferences = &constraintRefs
	}

	return result, nil
}

func (r *repo) getSubscriptionTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.GetSubscriptionResponse, error) {
	var req *restapi.GetSubscriptionRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal get subscription request")
	}

	id, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Subscriptionid)
	}

	if req.Auth.ClientID == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner")
	}

	sub, err := mem.GetSubscription(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not get Subscription from repo")
	}
	if sub == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
	}

	if dssmodels.Manager(*req.Auth.ClientID) != sub.Manager {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"Subscription owned by %s, but %s attempted to view", sub.Manager, *req.Auth.ClientID)
	}

	dependentOps, err := mem.GetDependentOperationalIntents(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not find dependent Operations")
	}

	p, err := sub.ToRest(dependentOps)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to convert Subscription to REST")
	}

	return &restapi.GetSubscriptionResponse{Subscription: *p}, nil
}

func (r *repo) querySubscriptionTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.QuerySubscriptionsResponse, error) {
	var req *restapi.QuerySubscriptionsRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal query subscriptions request")
	}

	if req.Auth.ClientID == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner")
	}

	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	vol4, err := dssmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model")
	}

	subs, err := mem.SearchSubscriptions(ctx, vol4)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error searching Subscriptions in repo")
	}

	response := &restapi.QuerySubscriptionsResponse{
		Subscriptions: make([]restapi.Subscription, 0),
	}
	for _, sub := range subs {
		if sub.EndTime.Before(proposal.Timestamp) || sub.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			continue
		}
		dependentOps, err := mem.GetDependentOperationalIntents(ctx, sub.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not find dependent Operations")
		}
		p, err := sub.ToRest(dependentOps)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting Subscription model to REST")
		}
		response.Subscriptions = append(response.Subscriptions, *p)
	}

	return response, nil
}

func (r *repo) deleteSubscriptionTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.DeleteSubscriptionResponse, error) {
	var req *restapi.DeleteSubscriptionRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal delete subscription request")
	}

	id, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Subscriptionid)
	}

	version := scdmodels.OVN(req.Version)
	if version == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing version")
	}

	if req.Auth.ClientID == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner")
	}

	old, err := mem.GetSubscription(ctx, id)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Could not get Subscription from repo")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
	case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"Subscription owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	case old.Version != version:
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", version)
	}

	dependentOps, err := mem.GetDependentOperationalIntents(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not find dependent Operations")
	}
	if len(dependentOps) > 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscriptions with dependent Operations may not be removed")
	}

	cp := r.memStore.Checkpoint()

	if err = mem.DeleteSubscription(ctx, id); err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return nil, stacktrace.Propagate(err, "Could not delete Subscription from repo")
	}

	p, err := old.ToRest(dependentOps)
	if err != nil {
		if restoreErr := r.memStore.Restore(cp); restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}

		return nil, stacktrace.Propagate(err, "Error converting Subscription model to REST")
	}

	return &restapi.DeleteSubscriptionResponse{Subscription: *p}, nil
}
