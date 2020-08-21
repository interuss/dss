package scd

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scderr "github.com/interuss/dss/pkg/scd/errors"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc/status"
)

// DeleteOperationReference deletes a single operation ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationReference(ctx context.Context, req *scdpb.DeleteOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	// Retrieve Operation ID
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityuuid())
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.ChangeOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get Operation to delete
		old, err := r.GetOperation(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get Operation from repo")
		}
		if old == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Operation %s not found", id)
		}

		// Validate deletion request
		if old.Owner != owner {
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Operation is owned by different client"),
				"Operation owned by %s, but %s attempted to delete", old.Owner, owner)
		}

		// Get the Subscription supporting the Operation
		sub, err := r.GetSubscription(ctx, old.SubscriptionID)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get Operation's Subscription from repo")
		}
		if sub == nil {
			return stacktrace.NewError("Operation's Subscription missing from repo")
		}

		removeImplicitSubscription := false
		if sub.ImplicitSubscription {
			// Get the Subscription's dependent Operations
			dependentOps, err := r.GetDependentOperations(ctx, sub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Could not find dependent Operations")
			}
			if len(dependentOps) == 0 {
				return stacktrace.NewError("An implicit Subscription had no dependent Operations")
			} else if len(dependentOps) == 1 {
				removeImplicitSubscription = true
			}
		}

		// Find Subscriptions that may overlap the Operation's Volume4D
		allsubs, err := r.SearchSubscriptions(ctx, &dssmodels.Volume4D{
			StartTime: old.StartTime,
			EndTime:   old.EndTime,
			SpatialVolume: &dssmodels.Volume3D{
				AltitudeHi: old.AltitudeUpper,
				AltitudeLo: old.AltitudeLower,
				Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
					return old.Cells, nil
				}),
			}})
		if err != nil {
			return stacktrace.Propagate(err, "Unable to search Subscriptions in repo")
		}

		// Limit Subscription notifications to only those interested in Operations
		var subs repos.Subscriptions
		for _, s := range allsubs {
			if s.NotifyForOperations {
				subs = append(subs, s)
			}
		}

		// Increment notification indices for Subscriptions to be notified
		if err := subs.IncrementNotificationIndices(ctx, r); err != nil {
			return stacktrace.Propagate(err, "Unable to increment notification indices")
		}

		// Delete Operation from repo
		if err := r.DeleteOperation(ctx, id); err != nil {
			return stacktrace.Propagate(err, "Unable to delete Operation from repo")
		}

		if removeImplicitSubscription {
			// Automatically remove a now-unused implicit Subscription
			err = r.DeleteSubscription(ctx, sub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to delete associated implicit Subscription")
			}
		}

		// Convert deleted Operation to proto
		opProto, err := old.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Operation to proto")
		}

		// Return response to client
		response = &scdpb.ChangeOperationReferenceResponse{
			OperationReference: opProto,
			Subscribers:        makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// GetOperationReference returns a single operation ref for the given ID.
func (a *Server) GetOperationReference(ctx context.Context, req *scdpb.GetOperationReferenceRequest) (*scdpb.GetOperationReferenceResponse, error) {
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityuuid())
	}

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.GetOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		op, err := r.GetOperation(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get Operation from repo")
		}
		if op == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Operation %s not found", id)
		}

		if op.Owner != owner {
			op.OVN = scdmodels.OVN("")
		}

		p, err := op.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Operation to proto")
		}

		response = &scdpb.GetOperationReferenceResponse{
			OperationReference: p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// SearchOperationReferences queries existing operation refs in the given
// bounds.
func (a *Server) SearchOperationReferences(ctx context.Context, req *scdpb.SearchOperationReferencesRequest) (*scdpb.SearchOperationReferenceResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDProto(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.SearchOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		ops, err := r.SearchOperations(ctx, vol4)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to search for Operations in repo")
		}

		// Create response for client
		response = &scdpb.SearchOperationReferenceResponse{}
		for _, op := range ops {
			p, err := op.ToProto()
			if err != nil {
				return stacktrace.Propagate(err, "Could not convert Operation model to proto")
			}
			if op.Owner != owner {
				p.Ovn = ""
			}
			response.OperationReferences = append(response.OperationReferences, p)
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// PutOperationReference creates a single operation ref.
func (a *Server) PutOperationReference(ctx context.Context, req *scdpb.PutOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityuuid())
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var (
		params  = req.GetParams()
		extents = make([]*dssmodels.Volume4D, len(params.GetExtents()))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	err = scdmodels.ValidateUSSBaseURL(params.UssBaseUrl)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
	}

	state := scdmodels.OperationState(params.State)
	if !state.IsValidInDSS() {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid Operation state: %s", params.State)
	}

	for idx, extent := range params.GetExtents() {
		cExtent, err := dssmodels.Volume4DFromSCDProto(extent)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to parse extents")
		}
		extents[idx] = cExtent
	}
	uExtent, err := dssmodels.UnionVolumes4D(extents...)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to union extents")
	}

	if uExtent.StartTime == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_start from extents")
	}
	if uExtent.EndTime == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_end from extents")
	}

	if time.Now().After(*uExtent.EndTime) {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Operations may not end in the past")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid area")
	}

	if uExtent.EndTime.Before(*uExtent.StartTime) {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "End time is past the start time")
	}

	if params.OldVersion == 0 && params.State != "Accepted" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid state for version 0: `%s`", params.State)
	}

	subscriptionID, err := dssmodels.IDFromOptionalString(params.GetSubscriptionId())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format for Subscription ID: `%s`", params.GetSubscriptionId())
	}

	var response *scdpb.ChangeOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get existing Operation, if any, and validate request
		old, err := r.GetOperation(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get Operation from repo")
		}
		if old != nil {
			if old.Owner != owner {
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Operation is owned by different client"),
					"Operation owned by %s, but %s attempted to modify", old.Owner, owner)
			}
			if old.Version != scdmodels.Version(params.OldVersion) {
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Old version %d is not the current version", params.OldVersion),
					"Current version is %d but client specified version %d", old.Version, params.OldVersion)
			}
		} else {
			if params.OldVersion != 0 {
				return stacktrace.NewErrorWithCode(dsserr.NotFound, "Operation does not exist and therefore is not version %d", params.OldVersion)
			}
		}

		var sub *scdmodels.Subscription
		if subscriptionID.Empty() {
			// Create implicit Subscription
			err := scdmodels.ValidateUSSBaseURL(params.GetNewSubscription().GetUssBaseUrl())
			if err != nil {
				return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate USS base URL")
			}

			sub, err = r.UpsertSubscription(ctx, &scdmodels.Subscription{
				ID:         dssmodels.ID(uuid.New().String()),
				Owner:      owner,
				StartTime:  uExtent.StartTime,
				EndTime:    uExtent.EndTime,
				AltitudeLo: uExtent.SpatialVolume.AltitudeLo,
				AltitudeHi: uExtent.SpatialVolume.AltitudeHi,
				Cells:      cells,

				BaseURL:              params.GetNewSubscription().GetUssBaseUrl(),
				NotifyForOperations:  true,
				NotifyForConstraints: params.GetNewSubscription().GetNotifyForConstraints(),
				ImplicitSubscription: true,
			})
			if err != nil {
				return stacktrace.Propagate(err, "Failed to create implicit subscription")
			}
			subscriptionID = sub.ID
		} else {
			// Use existing Subscription
			sub, err = r.GetSubscription(ctx, subscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to get Subscription")
			}
			if sub == nil {
				return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription does not exist")
			}
			if sub.Owner != owner {
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Specificed Subscription is owned by different client"),
					"Subscription %s owned by %s, but %s attempted to use it for an Operation", subscriptionID, sub.Owner, owner)
			}
			updateSub := false
			if sub.StartTime != nil && sub.StartTime.After(*uExtent.StartTime) {
				if sub.ImplicitSubscription {
					sub.StartTime = uExtent.StartTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the Operation starts")
				}
			}
			if sub.EndTime != nil && sub.EndTime.Before(*uExtent.EndTime) {
				if sub.ImplicitSubscription {
					sub.EndTime = uExtent.EndTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the Operation ends")
				}
			}
			if !sub.Cells.Contains(cells) {
				if sub.ImplicitSubscription {
					sub.Cells = s2.CellUnionFromUnion(sub.Cells, cells)
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not cover entire spatial area of the Operation")
				}
			}
			if updateSub {
				sub, err = r.UpsertSubscription(ctx, sub)
				if err != nil {
					return stacktrace.Propagate(err, "Failed to update existing Subscription")
				}
			}
		}

		if state.RequiresKey() {
			// Construct a hash set of OVNs as the key
			key := map[scdmodels.OVN]bool{}
			for _, ovn := range params.GetKey() {
				key[scdmodels.OVN(ovn)] = true
			}

			// Identify Operations missing from the key
			var missingOps []*scdmodels.Operation
			relevantOps, err := r.SearchOperations(ctx, uExtent)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to SearchOperations")
			}
			for _, relevantOp := range relevantOps {
				if _, ok := key[relevantOp.OVN]; !ok {
					if relevantOp.Owner != owner {
						relevantOp.OVN = ""
					}
					missingOps = append(missingOps, relevantOp)
				}
			}

			// Identify Constraints missing from the key
			var missingConstraints []*scdmodels.Constraint
			if sub.NotifyForConstraints {
				constraints, err := r.SearchConstraints(ctx, uExtent)
				if err != nil {
					return stacktrace.Propagate(err, "Unable to SearchConstraints")
				}
				for _, relevantConstraint := range constraints {
					if _, ok := key[relevantConstraint.OVN]; !ok {
						if relevantConstraint.Owner != owner {
							relevantConstraint.OVN = ""
						}
						missingConstraints = append(missingConstraints, relevantConstraint)
					}
				}
			}

			// If the client is missing some OVNs, provide the pointers to the
			// information they need
			if len(missingOps) > 0 || len(missingConstraints) > 0 {
				p, err := scderr.MissingOVNsErrorResponse(missingOps, missingConstraints)
				if err != nil {
					return stacktrace.Propagate(err, "Failed to construct missing OVNs error message")
				}
				return stacktrace.Propagate(status.ErrorProto(p), "Missing OVNs")
			}
		}

		// Construct the new Operation
		op := &scdmodels.Operation{
			ID:      id,
			Owner:   owner,
			Version: scdmodels.Version(params.OldVersion + 1),

			StartTime:     uExtent.StartTime,
			EndTime:       uExtent.EndTime,
			AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,
			Cells:         cells,

			USSBaseURL:     params.UssBaseUrl,
			SubscriptionID: sub.ID,
			State:          state,
		}
		err = op.ValidateTimeRange()
		if err != nil {
			return stacktrace.Propagate(err, "Error validating time range")
		}

		// Compute total affected Volume4D for notification purposes
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
				}}
			notifyVol4, err = dssmodels.UnionVolumes4D(uExtent, oldVol4)
			if err != nil {
				return stacktrace.Propagate(err, "Error constructing 4D volumes union")
			}
		}

		// Upsert the Operation
		op, err = r.UpsertOperation(ctx, op)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to upsert operation in repo")
		}

		// Find Subscriptions that may need to be notified
		allsubs, err := r.SearchSubscriptions(ctx, notifyVol4)
		if err != nil {
			return err
		}

		// Limit Subscription notifications to only those interested in Operations
		var subs repos.Subscriptions
		for _, sub := range allsubs {
			if sub.NotifyForOperations {
				subs = append(subs, sub)
			}
		}

		// Increment notification indices for relevant Subscriptions
		err = subs.IncrementNotificationIndices(ctx, r)
		if err != nil {
			return err
		}

		// Convert upserted Operation to proto
		p, err := op.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Operation to proto")
		}

		// Return response to client
		response = &scdpb.ChangeOperationReferenceResponse{
			OperationReference: p,
			Subscribers:        makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}
