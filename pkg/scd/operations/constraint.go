package operations

import (
	"context"
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

func init() {
	Registry[restapi.DeleteConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.DeleteConstraintReferenceRequest],
		Execute: executeDeleteConstraint,
	}
	Registry[restapi.GetConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.GetConstraintReferenceRequest],
		Execute:    executeGetConstraint,
		IsReadOnly: true,
	}
	Registry[restapi.CreateConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*createConstraintPayload],
		Execute: executePutConstraint,
	}
	Registry[restapi.UpdateConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*updateConstraintPayload],
		Execute: executePutConstraint,
	}
	Registry[restapi.QueryConstraintReferencesOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.QueryConstraintReferencesRequest],
		Execute:    executeQueryConstraintReferences,
		IsReadOnly: true,
	}
}

func executeGetConstraint(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.GetConstraintReferenceRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetConstraintReferenceOperationID)
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	constraint, err := repo.GetConstraint(ctx, id)
	switch {
	case err == pgx.ErrNoRows:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
	case err != nil:
		return nil, stacktrace.Propagate(err, "Unable to get Constraint from repo")
	}

	if constraint.Manager != dssmodels.Manager(*req.Auth.ClientID) {
		constraint.OVN = scdmodels.NoOvnPhrase
	}

	// Return response to client
	return &restapi.GetConstraintReferenceResponse{
		ConstraintReference: *constraint.ToRest(),
	}, nil
}

type createConstraintPayload struct {
	Constraint *scdmodels.Constraint
}

func (p *createConstraintPayload) OperationID() string {
	return restapi.CreateConstraintReferenceOperationID
}

func (p *createConstraintPayload) constraint() *scdmodels.Constraint { return p.Constraint }

type updateConstraintPayload struct {
	Constraint *scdmodels.Constraint
}

func (p *updateConstraintPayload) OperationID() string {
	return restapi.UpdateConstraintReferenceOperationID
}

func (p *updateConstraintPayload) constraint() *scdmodels.Constraint { return p.Constraint }

// constraintPayload is implemented by createConstraintPayload and updateConstraintPayload
type constraintPayload interface {
	dssstore.OperationRequest
	constraint() *scdmodels.Constraint
}

// NewCreateConstraintPayload performs the request validation that can be done ahead of the
// transaction for a Constraint creation request.
func NewCreateConstraintPayload(entityid restapi.EntityID, manager dssmodels.Manager, params *restapi.PutConstraintReferenceParameters, allowHTTPBaseUrls bool, now time.Time) (dssstore.OperationRequest, error) {
	constraint, err := newConstraint(entityid, "", manager, params, allowHTTPBaseUrls, now)
	if err != nil {
		return nil, err
	}
	return &createConstraintPayload{Constraint: constraint}, nil
}

// NewUpdateConstraintPayload performs the request validation that can be done ahead of the
// transaction for a Constraint update request.
func NewUpdateConstraintPayload(entityid restapi.EntityID, ovn restapi.EntityOVN, manager dssmodels.Manager, params *restapi.PutConstraintReferenceParameters, allowHTTPBaseUrls bool, now time.Time) (dssstore.OperationRequest, error) {
	constraint, err := newConstraint(entityid, ovn, manager, params, allowHTTPBaseUrls, now)
	if err != nil {
		return nil, err
	}
	return &updateConstraintPayload{Constraint: constraint}, nil
}

// newConstraint performs the request validation that can be done ahead of the transaction.
func newConstraint(entityid restapi.EntityID, ovn restapi.EntityOVN, manager dssmodels.Manager, params *restapi.PutConstraintReferenceParameters, allowHTTPBaseUrls bool, now time.Time) (*scdmodels.Constraint, error) {
	id, err := dssmodels.IDFromString(string(entityid))
	if err != nil {
		return nil, stacktrace.NewError("Invalid ID format: `%s`", entityid)
	}

	if !allowHTTPBaseUrls {
		if err := scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl)); err != nil {
			return nil, stacktrace.Propagate(err, "Failed to validate base URL")
		}
	}

	// Start and end times are required for each volume
	// The end time may not be in the past
	volume, err := scdmodels.UnionCellsVolume4DFromSCDRest(
		params.Extents,
		scdmodels.WithRequireCellsTimeBounds(),
		scdmodels.WithRequireCellsEndTimeAfter(now),
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid extents")
	}

	return &scdmodels.Constraint{
		ID:            id,
		Manager:       manager,
		OVN:           scdmodels.OVN(ovn),
		USSBaseURL:    string(params.UssBaseUrl),
		CellsVolume4D: volume,
	}, nil
}

// executePutConstraint inserts or updates a Constraint.
// If ovn is empty (""), it will attempt to create a new Constraint.
func executePutConstraint(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(constraintPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.CreateConstraintReferenceOperationID)
	}
	constraint := payload.constraint()
	ovn := constraint.OVN

	version := scdmodels.VersionNumber(1)

	// Get existing Constraint, if any, and validate request
	old, err := repo.GetConstraint(ctx, constraint.ID)
	switch {
	case err == pgx.ErrNoRows:
		// No existing Constraint; verify that creation was requested
		if ovn != "" {
			return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Old version %s does not exist", ovn)
		}
	case err != nil:
		return nil, stacktrace.Propagate(err, "Could not get Constraint from repo")
	}
	if old != nil {
		if old.Manager != constraint.Manager {
			return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"Constraint owned by %s, but %s attempted to modify", old.Manager, constraint.Manager)
		}
		if old.OVN != ovn {
			return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"Current version is %s but client specified version %s", old.OVN, ovn)
		}
		version = old.Version + 1
	}
	constraint.Version = version

	// Compute total affected cellsVolume for notification purposes
	notifyCellsVol4D := constraint.CellsVolume4D
	if old != nil {
		notifyCellsVol4D = dssmodels.UnionCellsVolumes4D(constraint.CellsVolume4D, old.CellsVolume4D)
	}

	// Upsert the Constraint
	constraint, err = repo.UpsertConstraint(ctx, constraint)
	if err != nil {
		return nil, err
	}

	// Find the Subscriptions interested in Constraints and increment their
	// notification indices.
	subs, err := repo.IncrementNotificationIndicesForConstraints(ctx, notifyCellsVol4D)
	if err != nil {
		return nil, err
	}

	// Return response to client
	return &restapi.ChangeConstraintReferenceResponse{
		ConstraintReference: *constraint.ToRest(),
		Subscribers:         makeSubscribersToNotify(subs),
	}, nil
}

func executeQueryConstraintReferences(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.QueryConstraintReferencesRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.QueryConstraintReferencesOperationID)
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to a cells-native volume
	cellsVol4, err := scdmodels.CellsVolume4DFromSCDRest(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model")
	}

	// Perform search query on Store
	constraints, err := repo.SearchConstraints(ctx, cellsVol4)
	if err != nil {
		return nil, err
	}

	// Create response for client
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

func executeDeleteConstraint(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.DeleteConstraintReferenceRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.DeleteConstraintReferenceOperationID)
	}

	// Retrieve Constraint ID
	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	// Make sure deletion request is valid
	old, err := repo.GetConstraint(ctx, id)
	switch {
	case err == pgx.ErrNoRows:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
	case err != nil:
		return nil, stacktrace.Propagate(err, "Unable to get Constraint from repo")
	case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"Constraint owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	case old.OVN != scdmodels.OVN(req.Ovn):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"Current version is %s but client specified version %s", old.OVN, scdmodels.OVN(req.Ovn))
	}

	// Delete Constraint in repo
	err = repo.DeleteConstraint(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to delete Constraint from repo")
	}

	// Find the Subscriptions interested in Constraints and increment their
	// notification indices.
	subs, err := repo.IncrementNotificationIndicesForConstraints(ctx, old.CellsVolume4D)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to increment notification indices")
	}

	// Return response to client
	return &restapi.ChangeConstraintReferenceResponse{
		ConstraintReference: *old.ToRest(),
		Subscribers:         makeSubscribersToNotify(subs),
	}, nil
}
