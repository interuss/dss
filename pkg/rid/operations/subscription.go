package operations

import (
	"context"

	ridv1 "github.com/interuss/dss/pkg/api/ridv1"
	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/locality"
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

func (p *insertSubscriptionPayload) subscription() *ridmodels.Subscription { return p.Subscription }

type updateSubscriptionPayload struct {
	Subscription *ridmodels.Subscription
}

func (p *updateSubscriptionPayload) OperationID() string { return ridv2.UpdateSubscriptionOperationID }

func (p *updateSubscriptionPayload) subscription() *ridmodels.Subscription { return p.Subscription }

// subscriptionPayload is implemented by insertSubscriptionPayload and updateSubscriptionPayload
type subscriptionPayload interface {
	dssstore.OperationRequest
	subscription() *ridmodels.Subscription
}

// NewInsertSubscriptionPayload performs the request validation that can be done ahead of the
// transaction for a Subscription creation request.
func NewInsertSubscriptionPayload(id dssmodels.ID, owner dssmodels.Owner, url string, extents *dssmodels.CellsVolume4D, allowHTTPBaseUrls bool) (dssstore.OperationRequest, error) {
	sub, err := newSubscription(id, owner, url, nil, extents, allowHTTPBaseUrls)
	if err != nil {
		return nil, err
	}
	return &insertSubscriptionPayload{Subscription: sub}, nil
}

// NewUpdateSubscriptionPayload performs the request validation that can be done ahead of the
// transaction for a Subscription update request.
func NewUpdateSubscriptionPayload(id dssmodels.ID, owner dssmodels.Owner, url string, version *dssmodels.Version, extents *dssmodels.CellsVolume4D, allowHTTPBaseUrls bool) (dssstore.OperationRequest, error) {
	sub, err := newSubscription(id, owner, url, version, extents, allowHTTPBaseUrls)
	if err != nil {
		return nil, err
	}
	return &updateSubscriptionPayload{Subscription: sub}, nil
}

// newSubscription performs the request validation that can be done ahead of the transaction.
func newSubscription(id dssmodels.ID, owner dssmodels.Owner, url string, version *dssmodels.Version, extents *dssmodels.CellsVolume4D, allowHTTPBaseUrls bool) (*ridmodels.Subscription, error) {
	if !allowHTTPBaseUrls {
		if err := ridmodels.ValidateURL(url); err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Subscription URL")
		}
	}
	return &ridmodels.Subscription{
		ID:            id,
		Owner:         owner,
		URL:           url,
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
		Execute: executePutSubscription,
	}
	Registry[ridv2.CreateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*insertSubscriptionPayload],
		Execute: executePutSubscription,
	}
	Registry[ridv1.UpdateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*updateSubscriptionPayload],
		Execute: executePutSubscription,
	}
	Registry[ridv2.UpdateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*updateSubscriptionPayload],
		Execute: executePutSubscription,
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

func executePutSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(subscriptionPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.CreateSubscriptionOperationID)
	}
	sub := payload.subscription()
	sub.Writer = locality.MustFromContext(ctx)

	old, err := repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := sub.AdjustTimeRange(timestamp.MustFromContext(ctx), old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	switch {
	case old == nil && sub.Version != nil:
		// The user wants to update an existing Subscription, but one wasn't found.
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", sub.ID.String())
	case old != nil && sub.Version == nil:
		// The user wants to create a new Subscription, but one already exists.
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", sub.ID.String())
	case old != nil && !sub.Version.Matches(old.Version):
		// The user wants to update a Subscription but the version doesn't match.
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", sub.Version),
			"Subscription currently at version %s but client specified %s", old.Version, sub.Version)
	case old != nil && old.Owner != sub.Owner:
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to modify", old.Owner, sub.Owner)
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

	if old == nil {
		ret, err := repo.InsertSubscription(ctx, sub)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error inserting Subscription into repo")
		}
		return ret, nil
	}

	ret, err := repo.UpdateSubscription(ctx, sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating Subscription in repo")
	}
	return ret, nil
}
