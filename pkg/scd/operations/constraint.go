package operations

import (
	"context"
	"errors"
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/pb/scdpb"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

func init() {
	Registry[restapi.DeleteConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		EncodePayload: encodeDeleteConstraint,
		DecodePayload: decodeDeleteConstraint,
		Execute:       executeDeleteConstraint,
	}
	Registry[restapi.GetConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		EncodePayload: encodeGetConstraint,
		DecodePayload: decodeGetConstraint,
		Execute:       executeGetConstraint,
		IsReadOnly:    true,
	}
	Registry[restapi.CreateConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		EncodePayload: encodeCreateConstraint,
		DecodePayload: decodeCreateConstraint,
		Execute:       executePutConstraint,
	}
	Registry[restapi.UpdateConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		EncodePayload: encodeUpdateConstraint,
		DecodePayload: decodeUpdateConstraint,
		Execute:       executePutConstraint,
	}
	Registry[restapi.QueryConstraintReferencesOperationID] = dssstore.OperationHandler[repos.Repository]{
		EncodePayload: encodeQueryConstraints,
		DecodePayload: decodeQueryConstraints,
		Execute:       executeQueryConstraintReferences,
		IsReadOnly:    true,
	}
}

type getConstraintPayload struct {
	ID      dssmodels.ID
	Manager dssmodels.Manager
}

func (p *getConstraintPayload) OperationID() string {
	return restapi.GetConstraintReferenceOperationID
}

// NewGetConstraintPayload builds the payload for a Constraint retrieval request.
func NewGetConstraintPayload(id dssmodels.ID, manager dssmodels.Manager) dssstore.OperationRequest {
	return &getConstraintPayload{ID: id, Manager: manager}
}

func encodeGetConstraint(request dssstore.OperationRequest) (*raftpb.Payload, error) {
	payload, ok := request.(*getConstraintPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetConstraintReferenceOperationID)
	}
	id, err := payload.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Constraint ID")
	}
	return ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetConstraintReferenceOperation{
		GetConstraintReferenceOperation: &scdpb.GetConstraintReferenceOperation{Id: id, Manager: payload.Manager.String()},
	}}), nil
}

func decodeGetConstraint(payload *raftpb.Payload) (dssstore.OperationRequest, error) {
	req := payload.GetScdPayload().GetGetConstraintReferenceOperation()
	if req == nil {
		return nil, stacktrace.NewError("unexpected payload %T for operation %q", payload.GetPayload(), restapi.GetConstraintReferenceOperationID)
	}
	decoded := &getConstraintPayload{Manager: dssmodels.Manager(req.GetManager())}
	if err := decoded.ID.FromProto(req.GetId()); err != nil {
		return nil, stacktrace.Propagate(err, "invalid Constraint ID")
	}
	return decoded, nil
}

func executeGetConstraint(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*getConstraintPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetConstraintReferenceOperationID)
	}
	id := payload.ID

	constraint, err := repo.GetConstraint(ctx, id)
	switch {
	case errors.Is(err, repos.ErrNotFound):
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
	case err != nil:
		return nil, stacktrace.Propagate(err, "Unable to get Constraint from repo")
	}

	if constraint.Manager != payload.Manager {
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

func encodeCreateConstraint(request dssstore.OperationRequest) (*raftpb.Payload, error) {
	payload, ok := request.(*createConstraintPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.CreateConstraintReferenceOperationID)
	}
	constraint, err := payload.Constraint.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode Constraint for operation %q", restapi.CreateConstraintReferenceOperationID)
	}
	return ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_CreateConstraintReferenceOperation{
		CreateConstraintReferenceOperation: &scdpb.CreateConstraintReferenceOperation{Constraint: constraint},
	}}), nil
}

func decodeCreateConstraint(payload *raftpb.Payload) (dssstore.OperationRequest, error) {
	req := payload.GetScdPayload().GetCreateConstraintReferenceOperation()
	if req == nil {
		return nil, stacktrace.NewError("unexpected payload %T for operation %q", payload.GetPayload(), restapi.CreateConstraintReferenceOperationID)
	}
	constraint := new(scdmodels.Constraint)
	if err := constraint.FromProto(req.GetConstraint()); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode Constraint for operation %q", restapi.CreateConstraintReferenceOperationID)
	}
	return &createConstraintPayload{Constraint: constraint}, nil
}

func encodeUpdateConstraint(request dssstore.OperationRequest) (*raftpb.Payload, error) {
	payload, ok := request.(*updateConstraintPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.UpdateConstraintReferenceOperationID)
	}
	constraint, err := payload.Constraint.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode Constraint for operation %q", restapi.UpdateConstraintReferenceOperationID)
	}
	return ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_UpdateConstraintReferenceOperation{
		UpdateConstraintReferenceOperation: &scdpb.UpdateConstraintReferenceOperation{
			Constraint: constraint,
			Ovn:        payload.Constraint.OVN.ToProto(),
		},
	}}), nil
}

func decodeUpdateConstraint(payload *raftpb.Payload) (dssstore.OperationRequest, error) {
	req := payload.GetScdPayload().GetUpdateConstraintReferenceOperation()
	if req == nil {
		return nil, stacktrace.NewError("unexpected payload %T for operation %q", payload.GetPayload(), restapi.UpdateConstraintReferenceOperationID)
	}
	constraint := new(scdmodels.Constraint)
	if err := constraint.FromProto(req.GetConstraint()); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode Constraint for operation %q", restapi.UpdateConstraintReferenceOperationID)
	}
	if err := constraint.OVN.FromProto(req.GetOvn()); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode Constraint OVN for operation %q", restapi.UpdateConstraintReferenceOperationID)
	}
	return &updateConstraintPayload{Constraint: constraint}, nil
}

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
	case errors.Is(err, repos.ErrNotFound):
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

type queryConstraintsPayload struct {
	AreaOfInterest *dssmodels.CellsVolume4D
	Manager        dssmodels.Manager
}

func (p *queryConstraintsPayload) OperationID() string {
	return restapi.QueryConstraintReferencesOperationID
}

// NewQueryConstraintsPayload builds the payload for a Constraints query.
func NewQueryConstraintsPayload(areaOfInterest *dssmodels.CellsVolume4D, manager dssmodels.Manager) dssstore.OperationRequest {
	return &queryConstraintsPayload{AreaOfInterest: areaOfInterest, Manager: manager}
}

func encodeQueryConstraints(request dssstore.OperationRequest) (*raftpb.Payload, error) {
	payload, ok := request.(*queryConstraintsPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.QueryConstraintReferencesOperationID)
	}
	if payload.AreaOfInterest == nil {
		return nil, stacktrace.NewError("missing area of interest")
	}
	return ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_QueryConstraintReferencesOperation{
		QueryConstraintReferencesOperation: &scdpb.QueryConstraintReferencesOperation{
			AreaOfInterest: payload.AreaOfInterest.ToProto(),
			Manager:        payload.Manager.String(),
		},
	}}), nil
}

func decodeQueryConstraints(payload *raftpb.Payload) (dssstore.OperationRequest, error) {
	req := payload.GetScdPayload().GetQueryConstraintReferencesOperation()
	if req == nil {
		return nil, stacktrace.NewError("unexpected payload %T for operation %q", payload.GetPayload(), restapi.QueryConstraintReferencesOperationID)
	}
	if req.GetAreaOfInterest() == nil {
		return nil, stacktrace.NewError("missing area of interest")
	}
	decoded := &queryConstraintsPayload{AreaOfInterest: new(dssmodels.CellsVolume4D), Manager: dssmodels.Manager(req.GetManager())}
	decoded.AreaOfInterest.FromProto(req.GetAreaOfInterest())
	return decoded, nil
}

func executeQueryConstraintReferences(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*queryConstraintsPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.QueryConstraintReferencesOperationID)
	}

	// Perform search query on Store
	constraints, err := repo.SearchConstraints(ctx, payload.AreaOfInterest)
	if err != nil {
		return nil, err
	}

	// Create response for client
	response := &restapi.QueryConstraintReferencesResponse{
		ConstraintReferences: make([]restapi.ConstraintReference, 0, len(constraints)),
	}
	for _, constraint := range constraints {
		p := constraint.ToRest()
		if constraint.Manager != payload.Manager {
			noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
			p.Ovn = &noOvnPhrase
		}
		response.ConstraintReferences = append(response.ConstraintReferences, *p)
	}

	return response, nil
}

type deleteConstraintPayload struct {
	ID      dssmodels.ID
	Manager dssmodels.Manager
	OVN     scdmodels.OVN
}

func (p *deleteConstraintPayload) OperationID() string {
	return restapi.DeleteConstraintReferenceOperationID
}

// NewDeleteConstraintPayload builds the payload for a Constraint deletion request.
func NewDeleteConstraintPayload(id dssmodels.ID, manager dssmodels.Manager, ovn scdmodels.OVN) dssstore.OperationRequest {
	return &deleteConstraintPayload{ID: id, Manager: manager, OVN: ovn}
}

func encodeDeleteConstraint(request dssstore.OperationRequest) (*raftpb.Payload, error) {
	payload, ok := request.(*deleteConstraintPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.DeleteConstraintReferenceOperationID)
	}
	id, err := payload.ID.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid Constraint ID")
	}
	return ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_DeleteConstraintReferenceOperation{
		DeleteConstraintReferenceOperation: &scdpb.DeleteConstraintReferenceOperation{
			Id:      id,
			Manager: payload.Manager.String(),
			Ovn:     payload.OVN.ToProto(),
		},
	}}), nil
}

func decodeDeleteConstraint(payload *raftpb.Payload) (dssstore.OperationRequest, error) {
	req := payload.GetScdPayload().GetDeleteConstraintReferenceOperation()
	if req == nil {
		return nil, stacktrace.NewError("unexpected payload %T for operation %q", payload.GetPayload(), restapi.DeleteConstraintReferenceOperationID)
	}
	decoded := &deleteConstraintPayload{Manager: dssmodels.Manager(req.GetManager())}
	if err := decoded.ID.FromProto(req.GetId()); err != nil {
		return nil, stacktrace.Propagate(err, "invalid Constraint ID")
	}
	if err := decoded.OVN.FromProto(req.GetOvn()); err != nil {
		return nil, stacktrace.Propagate(err, "invalid Constraint OVN")
	}
	return decoded, nil
}

func executeDeleteConstraint(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*deleteConstraintPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.DeleteConstraintReferenceOperationID)
	}
	id := payload.ID

	// Make sure deletion request is valid
	old, err := repo.GetConstraint(ctx, id)
	switch {
	case errors.Is(err, repos.ErrNotFound):
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
	case err != nil:
		return nil, stacktrace.Propagate(err, "Unable to get Constraint from repo")
	case old.Manager != payload.Manager:
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"Constraint owned by %s, but %s attempted to delete", old.Manager, payload.Manager)
	case old.OVN != payload.OVN:
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"Current version is %s but client specified version %s", old.OVN, payload.OVN)
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
