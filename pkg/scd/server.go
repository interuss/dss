package scd

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"
	"github.com/interuss/stacktrace"
)

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

// Server implements scdv1.Implementation.
type Server struct {
	Store      scdstore.Store
	Timeout    time.Duration
	EnableHTTP bool
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *restapi.MakeDssReportRequest,
) restapi.MakeDssReportResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.MakeDssReportResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.MakeDssReportResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	return restapi.MakeDssReportResponseSet{Response400: &restapi.ErrorResponse{
		Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Not yet implemented"))}}
}

func setAuthError(ctx context.Context, authErr error, resp401, resp403 **restapi.ErrorResponse, resp500 **api.InternalServerErrorBody) {
	switch stacktrace.GetCode(authErr) {
	case dsserr.Unauthenticated:
		*resp401 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authentication failed"))}
	case dsserr.PermissionDenied:
		*resp403 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authorization failed"))}
	default:
		*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}
	}
}
