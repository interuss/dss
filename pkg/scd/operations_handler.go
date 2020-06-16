package scd

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scderr "github.com/interuss/dss/pkg/scd/errors"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"
	"google.golang.org/grpc/status"
)

// DeleteOperationReference deletes a single operation ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationReference(ctx context.Context, req *scdpb.DeleteOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	// Retrieve Operation ID
	idString := req.GetEntityuuid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Operation ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var response *scdpb.ChangeOperationReferenceResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		// Delete Operation in Store
		op, subs, err := store.DeleteOperation(ctx, id, owner)
		if err != nil {
			return err
		}
		if op == nil {
			return dsserr.Internal(fmt.Sprintf("DeleteOperation returned no Operation for ID: %s", id))
		}

		// Convert deleted Operation to proto
		opProto, err := op.ToProto()
		if err != nil {
			return dsserr.Internal(err.Error())
		}

		// Return response to client
		response = &scdpb.ChangeOperationReferenceResponse{
			OperationReference: opProto,
			Subscribers:        makeSubscribersToNotify(subs),
		}

		return nil
	}

	err := scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
	}

	return response, nil
}

// GetOperationReference returns a single operation ref for the given ID.
func (a *Server) GetOperationReference(ctx context.Context, req *scdpb.GetOperationReferenceRequest) (*scdpb.GetOperationReferenceResponse, error) {
	id := scdmodels.ID(req.GetEntityuuid())
	if id.Empty() {
		return nil, dsserr.BadRequest("missing Operation ID")
	}

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var response *scdpb.GetOperationReferenceResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		sub, err := store.GetOperation(ctx, id)
		if err != nil {
			return err
		}

		if sub.Owner != owner {
			sub.OVN = scdmodels.OVN("")
		}

		p, err := sub.ToProto()
		if err != nil {
			return dsserr.Internal(err.Error())
		}

		response = &scdpb.GetOperationReferenceResponse{
			OperationReference: p,
		}

		return nil
	}

	err := scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
	}

	return response, nil
}

// SearchOperationReferences queries existing operation refs in the given
// bounds.
func (a *Server) SearchOperationReferences(ctx context.Context, req *scdpb.SearchOperationReferencesRequest) (*scdpb.SearchOperationReferenceResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, dsserr.BadRequest("missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDProto(aoi)
	if err != nil {
		return nil, dsserr.Internal("failed to convert to internal geometry model")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var response *scdpb.SearchOperationReferenceResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		// Perform search query on Store
		ops, err := store.SearchOperations(ctx, vol4, owner)
		if err != nil {
			return err
		}

		// Create response for client
		response = &scdpb.SearchOperationReferenceResponse{}
		for _, op := range ops {
			p, err := op.ToProto()
			if err != nil {
				return dsserr.Internal("error converting Operation model to proto")
			}
			if op.Owner != owner {
				p.Ovn = ""
			}
			response.OperationReferences = append(response.OperationReferences, p)
		}

		return nil
	}

	err = scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
	}

	return response, nil
}

// PutOperationReference creates a single operation ref.
func (a *Server) PutOperationReference(ctx context.Context, req *scdpb.PutOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	id := scdmodels.ID(req.GetEntityuuid())
	if id.Empty() {
		return nil, dsserr.BadRequest("missing Operation ID")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var (
		params  = req.GetParams()
		extents = make([]*dssmodels.Volume4D, len(params.GetExtents()))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, dsserr.BadRequest("missing required UssBaseUrl")
	}

	for idx, extent := range params.GetExtents() {
		cExtent, err := dssmodels.Volume4DFromSCDProto(extent)
		if err != nil {
			return nil, dsserr.BadRequest(fmt.Sprintf("failed to parse extents: %s", err))
		}
		extents[idx] = cExtent
	}
	uExtent, err := dssmodels.UnionVolumes4D(extents...)
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("failed to union extents: %s", err))
	}

	if uExtent.StartTime == nil {
		return nil, dsserr.BadRequest("missing time_start from extents")
	}
	if uExtent.EndTime == nil {
		return nil, dsserr.BadRequest("missing time_end from extents")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, dssErrorOfAreaError(err)
	}

	subscriptionID := scdmodels.ID(params.GetSubscriptionId())

	var response *scdpb.ChangeOperationReferenceResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		if subscriptionID.Empty() {
			if err := scdmodels.ValidateUSSBaseURL(
				params.GetNewSubscription().GetUssBaseUrl(),
			); err != nil {
				return dsserr.BadRequest(err.Error())
			}

			sub, _, err := store.UpsertSubscription(ctx, &scdmodels.Subscription{
				ID:         scdmodels.ID(uuid.New().String()),
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
				return dsserr.Internal(fmt.Sprintf("failed to create implicit subscription: %s", err))
			}
			subscriptionID = sub.ID
		}

		key := []scdmodels.OVN{}
		for _, ovn := range params.GetKey() {
			key = append(key, scdmodels.OVN(ovn))
		}

		op, subs, err := store.UpsertOperation(ctx, &scdmodels.Operation{
			ID:      id,
			Owner:   owner,
			Version: scdmodels.Version(params.OldVersion),

			StartTime:     uExtent.StartTime,
			EndTime:       uExtent.EndTime,
			AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,
			Cells:         cells,

			USSBaseURL:     params.UssBaseUrl,
			SubscriptionID: subscriptionID,
			State:          scdmodels.OperationState(params.State),
		}, key)

		if err == scderr.MissingOVNsInternalError() {
			// The client is missing some OVNs; provide the pointers to the
			// information they need
			ops, err := store.SearchOperations(ctx, uExtent, owner)
			if err != nil {
				return err
			}
			success, err := scderr.MissingOVNsErrorResponse(ops)
			if !success {
				return dsserr.Internal(fmt.Sprintf("failed to construct missing OVNs error message: %s", err))
			}
			return err
		}

		if err != nil {
			if _, ok := status.FromError(err); ok {
				return err
			}
			return dsserr.Internal(fmt.Sprintf("failed to upsert operation: %s", err))
		}

		p, err := op.ToProto()
		if err != nil {
			return dsserr.Internal("could not convert Operation to proto")
		}

		response = &scdpb.ChangeOperationReferenceResponse{
			OperationReference: p,
			Subscribers:        makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
	}

	return response, nil
}
