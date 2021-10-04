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
	"github.com/interuss/stacktrace"
	"google.golang.org/grpc/status"
)

// DeleteOperationalIntentReference deletes a single operational intent ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationalIntentReference(ctx context.Context, req *scdpb.DeleteOperationalIntentReferenceRequest) (*scdpb.ChangeOperationalIntentReferenceResponse, error) {
	// Retrieve OperationalIntent ID
	id, err := dssmodels.IDFromString(req.GetEntityid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityid())
	}

	// Retrieve ID of client making call
	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
	}

	var response *scdpb.ChangeOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get OperationalIntent to delete
		old, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationIntent from repo")
		}
		if old == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
		}

		// Validate deletion request
		if old.Manager != manager {
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"OperationalIntent owned by %s, but %s attempted to delete", old.Manager, manager)
		}

		// Get the Subscription supporting the OperationalIntent
		sub, err := r.GetSubscription(ctx, old.SubscriptionID)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
		}
		if sub == nil {
			return stacktrace.NewError("OperationalIntent's Subscription missing from repo")
		}

		removeImplicitSubscription := false
		if sub.ImplicitSubscription {
			// Get the Subscription's dependent OperationalIntents
			dependentOps, err := r.GetDependentOperationalIntents(ctx, sub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Could not find dependent OperationalIntents")
			}
			if len(dependentOps) == 0 {
				return stacktrace.NewError("An implicit Subscription had no dependent OperationalIntents")
			} else if len(dependentOps) == 1 {
				removeImplicitSubscription = true
			}
		}

		// Find Subscriptions that may overlap the OperationalIntent's Volume4D
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

		// Limit Subscription notifications to only those interested in OperationalIntents
		var subs repos.Subscriptions
		for _, s := range allsubs {
			if s.NotifyForOperationalIntents {
				subs = append(subs, s)
			}
		}

		// Increment notification indices for Subscriptions to be notified
		if err := subs.IncrementNotificationIndices(ctx, r); err != nil {
			return stacktrace.Propagate(err, "Unable to increment notification indices")
		}

		// Delete OperationalIntent from repo
		if err := r.DeleteOperationalIntent(ctx, id); err != nil {
			return stacktrace.Propagate(err, "Unable to delete OperationalIntent from repo")
		}

		if removeImplicitSubscription {
			// Automatically remove a now-unused implicit Subscription
			err = r.DeleteSubscription(ctx, sub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to delete associated implicit Subscription")
			}
		}

		// Convert deleted OperationalIntent to proto
		opProto, err := old.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert OperationalIntent to proto")
		}

		// Return response to client
		response = &scdpb.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: opProto,
			Subscribers:                makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// GetOperationalIntentReference returns a single operation intent ref for the given ID.
func (a *Server) GetOperationalIntentReference(ctx context.Context, req *scdpb.GetOperationalIntentReferenceRequest) (*scdpb.GetOperationalIntentReferenceResponse, error) {
	id, err := dssmodels.IDFromString(req.GetEntityid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityid())
	}

	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
	}

	var response *scdpb.GetOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		op, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationalIntent from repo")
		}
		if op == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
		}

		if op.Manager != manager {
			op.OVN = scdmodels.OVN("")
		}

		p, err := op.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert OperationalIntent to proto")
		}

		response = &scdpb.GetOperationalIntentReferenceResponse{
			OperationalIntentReference: p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// QueryOperationalIntentsReferences queries existing operational intent refs in the given
// bounds.
func (a *Server) QueryOperationalIntentReferences(ctx context.Context, req *scdpb.QueryOperationalIntentReferencesRequest) (*scdpb.QueryOperationalIntentReferenceResponse, error) {
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
	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
	}

	var response *scdpb.QueryOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		ops, err := r.SearchOperationalIntents(ctx, vol4)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to query for OperationalIntents in repo")
		}

		// Create response for client
		response = &scdpb.QueryOperationalIntentReferenceResponse{}
		for _, op := range ops {
			p, err := op.ToProto()
			if err != nil {
				return stacktrace.Propagate(err, "Could not convert OperationalIntent model to proto")
			}
			if op.Manager != manager {
				p.Ovn = ""
			}
			response.OperationalIntentReferences = append(response.OperationalIntentReferences, p)
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

func (a *Server) CreateOperationalIntentReference(ctx context.Context, req *scdpb.CreateOperationalIntentReferenceRequest) (*scdpb.ChangeOperationalIntentReferenceResponse, error) {
	return a.PutOperationalIntentReference(ctx, req.GetEntityid(), "", req.GetParams())
}

func (a *Server) UpdateOperationalIntentReference(ctx context.Context, req *scdpb.UpdateOperationalIntentReferenceRequest) (*scdpb.ChangeOperationalIntentReferenceResponse, error) {
	return a.PutOperationalIntentReference(ctx, req.GetEntityid(), req.Ovn, req.GetParams())
}

// PutOperationalIntentReference inserts or updates an Operational Intent.
// If the ovn argument is empty (""), it will attempt to create a new Operational Intent.
func (a *Server) PutOperationalIntentReference(ctx context.Context, entityid string, ovn string, params *scdpb.PutOperationalIntentReferenceParameters) (*scdpb.ChangeOperationalIntentReferenceResponse, error) {
	id, err := dssmodels.IDFromString(entityid)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", entityid)
	}

	// Retrieve ID of client making call
	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
	}

	var (
		extents = make([]*dssmodels.Volume4D, len(params.GetExtents()))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	err = scdmodels.ValidateUSSBaseURL(params.UssBaseUrl)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
	}

	state := scdmodels.OperationalIntentState(params.State)
	if !state.IsValidInDSS() {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid OperationalIntent state: %s", params.State)
	}

	for idx, extent := range params.GetExtents() {
		cExtent, err := dssmodels.Volume4DFromSCDProto(extent)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to parse extent %d", idx)
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
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "OperationalIntents may not end in the past")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	if uExtent.EndTime.Before(*uExtent.StartTime) {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "End time is past the start time")
	}

	if ovn == "" && params.State != "Accepted" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid state for initial version: `%s`", params.State)
	}

	subscriptionID, err := dssmodels.IDFromOptionalString(params.GetSubscriptionId())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format for Subscription ID: `%s`", params.GetSubscriptionId())
	}

	var response *scdpb.ChangeOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		var version int32 // Version of the Operational Intent (0 means creation requested).

		// Get existing OperationalIntent, if any, and validate request
		old, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get OperationalIntent from repo")
		}
		if old != nil {
			if old.Manager != manager {
				return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
					"OperationalIntent owned by %s, but %s attempted to modify", old.Manager, manager)
			}
			if old.OVN != scdmodels.OVN(ovn) {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
					"Current version is %s but client specified version %s", old.OVN, ovn)
			}

			version = int32(old.Version)
		} else {
			if ovn != "" {
				return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent does not exist and therefore is not version %s", ovn)
			}

			version = 0
		}

		var sub *scdmodels.Subscription
		if subscriptionID.Empty() {
			// Create implicit Subscription
			err := scdmodels.ValidateUSSBaseURL(params.GetNewSubscription().GetUssBaseUrl())
			if err != nil {
				return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate USS base URL")
			}

			sub, err = r.UpsertSubscription(ctx, &scdmodels.Subscription{
				ID:                          dssmodels.ID(uuid.New().String()),
				Manager:                     manager,
				StartTime:                   uExtent.StartTime,
				EndTime:                     uExtent.EndTime,
				AltitudeLo:                  uExtent.SpatialVolume.AltitudeLo,
				AltitudeHi:                  uExtent.SpatialVolume.AltitudeHi,
				Cells:                       cells,
				USSBaseURL:                  params.GetNewSubscription().GetUssBaseUrl(),
				NotifyForOperationalIntents: true,
				NotifyForConstraints:        params.GetNewSubscription().GetNotifyForConstraints(),
				ImplicitSubscription:        true,
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
			if sub.Manager != manager {
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Specificed Subscription is owned by different client"),
					"Subscription %s owned by %s, but %s attempted to use it for an OperationalIntent", subscriptionID, sub.Manager, manager)
			}
			updateSub := false
			if sub.StartTime != nil && sub.StartTime.After(*uExtent.StartTime) {
				if sub.ImplicitSubscription {
					sub.StartTime = uExtent.StartTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the OperationalIntent starts")
				}
			}
			if sub.EndTime != nil && sub.EndTime.Before(*uExtent.EndTime) {
				if sub.ImplicitSubscription {
					sub.EndTime = uExtent.EndTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the OperationalIntent ends")
				}
			}
			if !sub.Cells.Contains(cells) {
				if sub.ImplicitSubscription {
					sub.Cells = s2.CellUnionFromUnion(sub.Cells, cells)
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not cover entire spatial area of the OperationalIntent")
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

			// Identify OperationalIntents missing from the key
			var missingOps []*scdmodels.OperationalIntent
			relevantOps, err := r.SearchOperationalIntents(ctx, uExtent)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to SearchOperations")
			}
			for _, relevantOp := range relevantOps {
				if _, ok := key[relevantOp.OVN]; !ok {
					if relevantOp.Manager != manager {
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
						if relevantConstraint.Manager != manager {
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

		// Construct the new OperationalIntent
		op := &scdmodels.OperationalIntent{
			ID:      id,
			Manager: manager,
			Version: scdmodels.VersionNumber(version + 1),

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

		// Upsert the OperationalIntent
		op, err = r.UpsertOperationalIntent(ctx, op)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to upsert OperationalIntent in repo")
		}

		// Find Subscriptions that may need to be notified
		allsubs, err := r.SearchSubscriptions(ctx, notifyVol4)
		if err != nil {
			return err
		}

		// Limit Subscription notifications to only those interested in OperationalIntents
		var subs repos.Subscriptions
		for _, sub := range allsubs {
			if sub.NotifyForOperationalIntents {
				subs = append(subs, sub)
			}
		}

		// Increment notification indices for relevant Subscriptions
		err = subs.IncrementNotificationIndices(ctx, r)
		if err != nil {
			return err
		}

		// Convert upserted OperationalIntent to proto
		p, err := op.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert OperationalIntent to proto")
		}

		// Return response to client
		response = &scdpb.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: p,
			Subscribers:                makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}
