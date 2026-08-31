package operations

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
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

// putSubscriptionPayload carries an already-validated rid Subscription create/update request:
// the HTTP handler (v1 or v2) does all request-format validation and computes the S2 covering
// once, so it isn't recomputed by every raft node during apply.
type putSubscriptionPayload struct {
	Op         string
	ID         dssmodels.ID
	Owner      dssmodels.Owner
	URL        string
	Version    *dssmodels.Version
	Cells      s2.CellUnion
	StartTime  *time.Time
	EndTime    *time.Time
	AltitudeLo *float32
	AltitudeHi *float32
}

func (p *putSubscriptionPayload) OperationID() string { return p.Op }

// NewPutSubscriptionPayload builds a putSubscriptionPayload. id, owner, url, and version are
// assumed already validated by the caller; extents is used only to compute the S2 covering.
func NewPutSubscriptionPayload(opID string, id dssmodels.ID, owner dssmodels.Owner, url string, version *dssmodels.Version, extents *dssmodels.Volume4D) (*putSubscriptionPayload, error) {
	cells, err := extents.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	return &putSubscriptionPayload{
		Op:         opID,
		ID:         id,
		Owner:      owner,
		URL:        url,
		Version:    version,
		Cells:      cells,
		StartTime:  extents.StartTime,
		EndTime:    extents.EndTime,
		AltitudeLo: extents.SpatialVolume.AltitudeLo,
		AltitudeHi: extents.SpatialVolume.AltitudeHi,
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
		Decode:  dssstore.DecodeJSON[*putSubscriptionPayload],
		Execute: executeInsertSubscription,
	}
	Registry[ridv2.CreateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*putSubscriptionPayload],
		Execute: executeInsertSubscription,
	}
	Registry[ridv1.UpdateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*putSubscriptionPayload],
		Execute: executeUpdateSubscription,
	}
	Registry[ridv2.UpdateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*putSubscriptionPayload],
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

func executeInsertSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*putSubscriptionPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.CreateSubscriptionOperationID)
	}

	sub := &ridmodels.Subscription{
		ID:         payload.ID,
		Owner:      payload.Owner,
		URL:        payload.URL,
		Cells:      payload.Cells,
		StartTime:  payload.StartTime,
		EndTime:    payload.EndTime,
		AltitudeLo: payload.AltitudeLo,
		AltitudeHi: payload.AltitudeHi,
		Writer:     locality.MustFromContext(ctx),
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

func executeUpdateSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*putSubscriptionPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.UpdateSubscriptionOperationID)
	}

	sub := &ridmodels.Subscription{
		ID:         payload.ID,
		Owner:      payload.Owner,
		URL:        payload.URL,
		Version:    payload.Version,
		Cells:      payload.Cells,
		StartTime:  payload.StartTime,
		EndTime:    payload.EndTime,
		AltitudeLo: payload.AltitudeLo,
		AltitudeHi: payload.AltitudeHi,
		Writer:     locality.MustFromContext(ctx),
	}

	return updateSubscription(ctx, repo, sub)
}

func updateSubscription(ctx context.Context, repo repos.Repository, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	old, err := repo.GetSubscription(ctx, sub.ID)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	case old == nil:
		// The user wants to update an existing subscription, but one wasn't found.
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", sub.ID.String())
	case !sub.Version.Matches(old.Version):
		// The user wants to update a subscription but the version doesn't match.
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", sub.Version),
			"Subscription currently at version %s but client specified %s", old.Version, sub.Version)
	case old.Owner != sub.Owner:
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to update", old.Owner, sub.Owner)
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := sub.AdjustTimeRange(timestamp.MustFromContext(ctx), old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
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

	ret, err := repo.UpdateSubscription(ctx, sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating Subscription in repo")
	}
	return ret, nil
}
