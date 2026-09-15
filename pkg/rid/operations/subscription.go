package operations

import (
	"context"

	ridv1 "github.com/interuss/dss/pkg/api/ridv1"
	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/locality"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	apiv1 "github.com/interuss/dss/pkg/rid/models/api/v1"
	apiv2 "github.com/interuss/dss/pkg/rid/models/api/v2"
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

// Defined in requirement DSS0030.
const maxSubscriptionsPerArea = 10

func init() {
	Registry[ridv1.DeleteSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*ridv1.DeleteSubscriptionRequest],
		Execute: executeDeleteSubscription,
	}
	Registry[ridv2.DeleteSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*ridv2.DeleteSubscriptionRequest],
		Execute: executeDeleteSubscription,
	}
	Registry[ridv1.CreateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*ridv1.CreateSubscriptionRequest],
		Execute: executeInsertSubscription,
	}
	Registry[ridv2.CreateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*ridv2.CreateSubscriptionRequest],
		Execute: executeInsertSubscription,
	}
}

func executeDeleteSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	var (
		rawID      string
		rawVersion string
		clientID   *string
	)

	switch req := request.(type) {
	case *ridv1.DeleteSubscriptionRequest:
		rawID, rawVersion, clientID = string(req.Id), req.Version, req.Auth.ClientID
	case *ridv2.DeleteSubscriptionRequest:
		rawID, rawVersion, clientID = string(req.Id), req.Version, req.Auth.ClientID
	default:
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.DeleteSubscriptionOperationID)
	}

	version, err := dssmodels.VersionFromString(rawVersion)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version")
	}
	id, err := dssmodels.IDFromString(rawID)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}
	owner := dssmodels.Owner(*clientID)

	old, err := repo.GetSubscription(ctx, id)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
	case !version.Matches(old.Version):
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", version),
			"Subscription currently at version %s but client specified %s", old.Version, version)
	case old.Owner != owner:
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to delete", old.Owner, owner)
	}

	ret, err := repo.DeleteSubscription(ctx, old)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error deleting Subscription from repo")
	}
	return ret, nil
}

func executeInsertSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	var (
		rawID    string
		url      string
		clientID *string
		extents  *dssmodels.Volume4D
	)

	switch req := request.(type) {
	case *ridv1.CreateSubscriptionRequest:
		if req.Body.Callbacks.IdentificationServiceAreaUrl == nil {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required callbacks")
		}
		if len(req.Body.Extents.SpatialVolume.Footprint.Vertices) == 0 {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents")
		}
		e, err := apiv1.FromVolume4D(&req.Body.Extents)
		if err != nil {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err))
		}
		rawID, url, clientID, extents = string(req.Id), string(*req.Body.Callbacks.IdentificationServiceAreaUrl), req.Auth.ClientID, e

	case *ridv2.CreateSubscriptionRequest:
		if req.Body.UssBaseUrl == "" {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL")
		}
		e, err := apiv2.FromVolume4D(&req.Body.Extents)
		if err != nil {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err))
		}
		rawID, url, clientID, extents = string(req.Id), string(req.Body.UssBaseUrl), req.Auth.ClientID, e

	default:
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.CreateSubscriptionOperationID)
	}

	id, err := dssmodels.IDFromString(rawID)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	sub := &ridmodels.Subscription{
		ID:     id,
		Owner:  dssmodels.Owner(*clientID),
		URL:    url,
		Writer: locality.MustFromContext(ctx),
	}
	if err := sub.SetExtents(extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	return InsertSubscription(ctx, repo, sub)
}

// InsertSubscription applies the business rules for inserting a new Subscription: it does not
// perform any request-format validation, which is the caller's responsibility.
func InsertSubscription(ctx context.Context, repo repos.Repository, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	// Validate and perhaps correct StartTime and EndTime.
	if err := sub.AdjustTimeRange(timestamp.MustFromContext(ctx), nil); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to adjust time range")
	}

	// ensure it doesn't exist yet
	old, err := repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	}
	if old != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", sub.ID)
	}

	// Check the user hasn't created too many subscriptions in this area.
	count, err := repo.MaxSubscriptionCountInCellsByOwner(ctx, sub.Cells, sub.Owner)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to fetch subscription count, rejecting request")
	}
	if count >= maxSubscriptionsPerArea {
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.Exhausted, "Too many existing subscriptions in this area already"),
			"%s had %d subscriptions in the area", sub.Owner, count)
	}

	ret, err := repo.InsertSubscription(ctx, sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error inserting Subscription into repo")
	}
	return ret, nil
}
