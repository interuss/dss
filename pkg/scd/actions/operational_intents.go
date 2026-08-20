package actions

import (
	"context"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

func init() {
	Registry[restapi.GetOperationalIntentReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.GetOperationalIntentReferenceRequest],
		Execute:    ExecuteGetOperationalIntentReference,
		IsReadOnly: true,
	}
	Registry[restapi.QueryOperationalIntentReferencesOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.QueryOperationalIntentReferencesRequest],
		Execute:    ExecuteQueryOperationalIntentReferences,
		IsReadOnly: true,
	}
}

func ExecuteGetOperationalIntentReference(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.GetOperationalIntentReferenceRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetOperationalIntentReferenceOperationID)
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	op, err := repo.GetOperationalIntent(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get OperationalIntent from repo")
	}
	if op == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
	}

	if op.Manager != dssmodels.Manager(*req.Auth.ClientID) {
		op.OVN = scdmodels.NoOvnPhrase
	}

	return &restapi.GetOperationalIntentReferenceResponse{
		OperationalIntentReference: *op.ToRest(),
	}, nil
}

func ExecuteQueryOperationalIntentReferences(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.QueryOperationalIntentReferencesRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.QueryOperationalIntentReferencesOperationID)
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry")
	}

	// Perform search query on Store
	ops, err := repo.SearchOperationalIntents(ctx, vol4)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to query for OperationalIntents in repo")
	}

	// Create response for client
	response := &restapi.QueryOperationalIntentReferenceResponse{
		OperationalIntentReferences: make([]restapi.OperationalIntentReference, 0, len(ops)),
	}
	for _, op := range ops {
		p := op.ToRest()
		if op.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
			p.Ovn = &noOvnPhrase
		}
		response.OperationalIntentReferences = append(response.OperationalIntentReferences, *p)
	}

	return response, nil
}
