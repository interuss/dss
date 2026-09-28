package operations

import (
	"context"

	ridv1 "github.com/interuss/dss/pkg/api/ridv1"
	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

// Defined in requirement DSS0030.
const maxSubscriptionsPerArea = 10

type insertSubscriptionPayload struct {
	Subscription *ridmodels.Subscription
}

func (p *insertSubscriptionPayload) OperationID() string { return ridv2.CreateSubscriptionOperationID }

type updateSubscriptionPayload struct {
	Subscription *ridmodels.Subscription
}

func (p *updateSubscriptionPayload) OperationID() string { return ridv2.UpdateSubscriptionOperationID }

// NewInsertSubscriptionPayload performs the request validation that can be done ahead of the
// transaction for a Subscription creation request.
func NewInsertSubscriptionPayload(id dssmodels.ID, owner dssmodels.Owner, url string, writer string, extents *dssmodels.CellsVolume4D, allowHTTPBaseUrls bool) (dssstore.OperationRequest, error) {
	sub, err := newSubscription(id, owner, url, writer, nil, extents, allowHTTPBaseUrls)
	if err != nil {
		return nil, err
	}
	return &insertSubscriptionPayload{Subscription: sub}, nil
}

// NewUpdateSubscriptionPayload performs the request validation that can be done ahead of the
// transaction for a Subscription update request.
func NewUpdateSubscriptionPayload(id dssmodels.ID, owner dssmodels.Owner, url string, writer string, version *dssmodels.Version, extents *dssmodels.CellsVolume4D, allowHTTPBaseUrls bool) (dssstore.OperationRequest, error) {
	sub, err := newSubscription(id, owner, url, writer, version, extents, allowHTTPBaseUrls)
	if err != nil {
		return nil, err
	}
	return &updateSubscriptionPayload{Subscription: sub}, nil
}

// newSubscription performs the request validation that can be done ahead of the transaction.
func newSubscription(id dssmodels.ID, owner dssmodels.Owner, url string, writer string, version *dssmodels.Version, extents *dssmodels.CellsVolume4D, allowHTTPBaseUrls bool) (*ridmodels.Subscription, error) {
	if !allowHTTPBaseUrls {
		if err := ridmodels.ValidateURL(url); err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Subscription URL")
		}
	}
	return &ridmodels.Subscription{
		ID:            id,
		Owner:         owner,
		URL:           url,
		Writer:        writer,
		Version:       version,
		CellsVolume4D: extents,
	}, nil
}

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
		Decode:  dssstore.DecodeJSON[*insertSubscriptionPayload],
		Execute: executeCreateSubscription,
	}
	Registry[ridv2.CreateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*insertSubscriptionPayload],
		Execute: executeCreateSubscription,
	}
	Registry[ridv1.UpdateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*updateSubscriptionPayload],
		Execute: executeUpdateSubscription,
	}
	Registry[ridv2.UpdateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*updateSubscriptionPayload],
		Execute: executeUpdateSubscription,
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

func executeCreateSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*insertSubscriptionPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.CreateSubscriptionOperationID)
	}
	sub := payload.Subscription

	// Validate and perhaps correct StartTime and EndTime.
	if err := sub.AdjustTimeRange(timestamp.MustFromContext(ctx), nil); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	old, err := repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	}
	if old != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", sub.ID.String())
	}

	if err := checkSubscriptionCount(ctx, repo, sub); err != nil {
		return nil, err
	}

	ret, err := repo.InsertSubscription(ctx, sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error inserting Subscription into repo")
	}
	return ret, nil
}

func executeUpdateSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*updateSubscriptionPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.UpdateSubscriptionOperationID)
	}
	sub := payload.Subscription

	old, err := repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := sub.AdjustTimeRange(timestamp.MustFromContext(ctx), old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	switch {
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", sub.ID.String())
	case !sub.Version.Matches(old.Version):
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", sub.Version),
			"Subscription currently at version %s but client specified %s", old.Version, sub.Version)
	case old.Owner != sub.Owner:
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to modify", old.Owner, sub.Owner)
	}

	if err := checkSubscriptionCount(ctx, repo, sub); err != nil {
		return nil, err
	}

	ret, err := repo.UpdateSubscription(ctx, sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating Subscription in repo")
	}
	return ret, nil
}

// checkSubscriptionCount checks if the owner already has too many Subscriptions in the area covered by sub.
func checkSubscriptionCount(ctx context.Context, repo repos.Repository, sub *ridmodels.Subscription) error {
	count, err := repo.MaxSubscriptionCountInCellsByOwner(ctx, sub.Cells, sub.Owner)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to fetch subscription count, rejecting request")
	}
	if count >= maxSubscriptionsPerArea {
		return stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.Exhausted, "Too many existing subscriptions in this area already"),
			"%s had %d subscriptions in the area", sub.Owner, count)
	}
	return nil
}
