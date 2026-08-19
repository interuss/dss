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
	Registry[restapi.DeleteSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.DeleteSubscriptionRequest],
		Execute: ExecuteDeleteSubscription,
	}
}

func ExecuteDeleteSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.DeleteSubscriptionRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.DeleteSubscriptionOperationID)
	}

	id, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid subscription ID: %s", req.Subscriptionid)
	}

	// Check to make sure it's ok to delete this Subscription
	old, err := repo.GetSubscription(ctx, id)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Could not get Subscription from repo")
	case old == nil: // Return a 404 here.
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
	case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	case old.Version != scdmodels.OVN(req.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", scdmodels.OVN(req.Version))
	}

	// Get dependent Operations
	dependentOps, err := repo.GetDependentOperationalIntents(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not find dependent Operations")
	}
	if len(dependentOps) > 0 {
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscriptions with dependent Operations may not be removed"),
			"Subscription had %d dependent Operations", len(dependentOps))
	}

	// Delete Subscription in repo
	err = repo.DeleteSubscription(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not delete Subscription from repo")
	}

	// Convert deleted Subscription to REST
	p, err := old.ToRest(dependentOps)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error converting Subscription model to REST")
	}

	return &restapi.DeleteSubscriptionResponse{Subscription: *p}, nil
}
