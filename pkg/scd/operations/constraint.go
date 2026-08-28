package operations

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
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
		Decode:  dssstore.DecodeJSON[*restapi.CreateConstraintReferenceRequest],
		Execute: executePutConstraint,
	}
	Registry[restapi.UpdateConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.UpdateConstraintReferenceRequest],
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

// executePutConstraint inserts or updates a Constraint.
// If ovn is empty (""), it will attempt to create a new Constraint.
func executePutConstraint(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	var (
		manager  string
		entityid restapi.EntityID
		ovn      restapi.EntityOVN
		params   *restapi.PutConstraintReferenceParameters
	)

	switch req := request.(type) {
	case *restapi.CreateConstraintReferenceRequest:
		manager, entityid, params = *req.Auth.ClientID, req.Entityid, req.Body
	case *restapi.UpdateConstraintReferenceRequest:
		manager, entityid, ovn, params = *req.Auth.ClientID, req.Entityid, req.Ovn, req.Body
	default:
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.CreateConstraintReferenceOperationID)
	}

	validParams, err := validateAndReturnConstraintUpsertParams(timestamp.MustGetRequestTimestamp(ctx), entityid, params)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Constraint upsert parameters")
	}

	version := scdmodels.VersionNumber(1)

	// Get existing Constraint, if any, and validate request
	old, err := repo.GetConstraint(ctx, validParams.id)
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
		if old.Manager != dssmodels.Manager(manager) {
			return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"Constraint owned by %s, but %s attempted to modify", old.Manager, manager)
		}
		if old.OVN != scdmodels.OVN(ovn) {
			return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"Current version is %s but client specified version %s", old.OVN, ovn)
		}
		version = old.Version + 1
	}

	// Compute total affected Volume4D for notification purposes
	var notifyVol4 *dssmodels.Volume4D
	if old == nil {
		notifyVol4 = validParams.uExtent
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
		notifyVol4, err = dssmodels.UnionVolumes4D(validParams.uExtent, oldVol4)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error constructing 4D volumes union")
		}
	}

	// Construct the new Constraint
	constraint := validParams.toConstraint(dssmodels.Manager(manager), version)

	// Upsert the Constraint
	constraint, err = repo.UpsertConstraint(ctx, constraint)
	if err != nil {
		return nil, err
	}

	// Find the Subscriptions interested in Constraints and increment their
	// notification indices.
	subs, err := repo.IncrementNotificationIndicesForConstraints(ctx, notifyVol4)
	if err != nil {
		return nil, err
	}

	// Return response to client
	return &restapi.ChangeConstraintReferenceResponse{
		ConstraintReference: *constraint.ToRest(),
		Subscribers:         makeSubscribersToNotify(subs),
	}, nil
}

type validConstraintParams struct {
	id         dssmodels.ID
	uExtent    *dssmodels.Volume4D
	cells      s2.CellUnion
	ussBaseURL string
}

func (vp *validConstraintParams) toConstraint(manager dssmodels.Manager, version scdmodels.VersionNumber) *scdmodels.Constraint {
	return &scdmodels.Constraint{
		ID:      vp.id,
		Manager: manager,
		Version: version,

		StartTime:     vp.uExtent.StartTime,
		EndTime:       vp.uExtent.EndTime,
		AltitudeLower: vp.uExtent.SpatialVolume.AltitudeLo,
		AltitudeUpper: vp.uExtent.SpatialVolume.AltitudeHi,

		USSBaseURL: vp.ussBaseURL,
		Cells:      vp.cells,
	}
}

// validateAndReturnConstraintUpsertParams performs validation of Constraint upsert requests and returns a validConstraintParams struct if successful.
// Note that this does NOT check for anything related to access controls: any error returned should be labeled as a dsserr.BadRequest.
func validateAndReturnConstraintUpsertParams(
	now time.Time,
	entityid restapi.EntityID,
	params *restapi.PutConstraintReferenceParameters,
) (*validConstraintParams, error) {
	var err error
	valid := &validConstraintParams{}
	valid.id, err = dssmodels.IDFromString(string(entityid))
	if err != nil {
		return nil, stacktrace.NewError("Invalid ID format: `%s`", entityid)
	}

	valid.ussBaseURL = string(params.UssBaseUrl)

	// Start and end times are required for each volume
	// The end time may not be in the past
	valid.uExtent, err = scdmodels.UnionVolumes4DFromSCDRest(
		params.Extents,
		scdmodels.WithRequireTimeBounds(),
		scdmodels.WithRequireEndTimeAfter(now),
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid extents")
	}

	valid.cells, err = valid.uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid area")
	}

	return valid, nil
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

	// Parse area of interest to common Volume4D
	vol4, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model")
	}

	// Perform search query on Store
	constraints, err := repo.SearchConstraints(ctx, vol4)
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
	subs, err := repo.IncrementNotificationIndicesForConstraints(ctx, &dssmodels.Volume4D{
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
		return nil, stacktrace.Propagate(err, "Unable to increment notification indices")
	}

	// Return response to client
	return &restapi.ChangeConstraintReferenceResponse{
		ConstraintReference: *old.ToRest(),
		Subscribers:         makeSubscribersToNotify(subs),
	}, nil
}

func makeSubscribersToNotify(subscriptions []*scdmodels.Subscription) []restapi.SubscriberToNotify {
	result := []restapi.SubscriberToNotify{}

	subscriptionsByURL := map[string][]restapi.SubscriptionState{}
	for _, sub := range subscriptions {
		subState := restapi.SubscriptionState{
			SubscriptionId:    restapi.SubscriptionID(sub.ID.String()),
			NotificationIndex: restapi.SubscriptionNotificationIndex(sub.NotificationIndex),
		}
		subscriptionsByURL[sub.USSBaseURL] = append(subscriptionsByURL[sub.USSBaseURL], subState)
	}
	for url, states := range subscriptionsByURL {
		result = append(result, restapi.SubscriberToNotify{
			UssBaseUrl:    restapi.SubscriptionUssBaseURL(url),
			Subscriptions: states,
		})
	}

	return result
}
