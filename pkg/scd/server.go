package scd

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"
	"github.com/interuss/stacktrace"
)

const (
	strategicCoordinationScope   = "utm.strategic_coordination"
	constraintManagementScope    = "utm.constraint_management"
	constraintIngestionScope     = "utm.constraint_ingestion"
	availabilityArbitrationScope = "utm.availability_arbitration"
)

func makeSubscribersToNotify(subscriptions []*scdmodels.Subscription) []*scdpb.SubscriberToNotify {
	result := []*scdpb.SubscriberToNotify{}

	subscriptionsByURL := map[string][]*scdpb.SubscriptionState{}
	for _, sub := range subscriptions {
		subState := &scdpb.SubscriptionState{
			SubscriptionId:    sub.ID.String(),
			NotificationIndex: int32(sub.NotificationIndex),
		}
		subscriptionsByURL[sub.USSBaseURL] = append(subscriptionsByURL[sub.USSBaseURL], subState)
	}
	for url, states := range subscriptionsByURL {
		result = append(result, &scdpb.SubscriberToNotify{
			UssBaseUrl:    url,
			Subscriptions: states,
		})
	}

	return result
}

// Server implements scdpb.DiscoveryAndSynchronizationService.
type Server struct {
	Store      scdstore.Store
	Timeout    time.Duration
	EnableHTTP bool
}

// AuthScopes returns a map of endpoint to required Oauth scope.
func (a *Server) AuthScopes() map[auth.Operation]auth.KeyClaimedScopesValidator {
	// TODO: replace with correct scopes
	return map[auth.Operation]auth.KeyClaimedScopesValidator{
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/CreateConstraintReference":        auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/CreateOperationalIntentReference": auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/CreateSubscription":               auth.RequireAnyScope(strategicCoordinationScope, constraintIngestionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteConstraintReference":        auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteOperationalIntentReference": auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteSubscription":               auth.RequireAnyScope(strategicCoordinationScope, constraintIngestionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetConstraintReference":           auth.RequireAnyScope(constraintManagementScope, constraintIngestionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetOperationalIntentReference":    auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetSubscription":                  auth.RequireAnyScope(strategicCoordinationScope, constraintIngestionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetUssAvailability":               auth.RequireAnyScope(strategicCoordinationScope, availabilityArbitrationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/MakeDssReport":                    auth.RequireAnyScope(strategicCoordinationScope, constraintIngestionScope, constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QueryConstraintReferences":        auth.RequireAnyScope(constraintIngestionScope, constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QueryOperationalIntentReferences": auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QuerySubscriptions":               auth.RequireAnyScope(strategicCoordinationScope, constraintIngestionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/SetUssAvailability":               auth.RequireAnyScope(availabilityArbitrationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/UpdateConstraintReference":        auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/UpdateOperationalIntentReference": auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/UpdateSubscription":               auth.RequireAnyScope(strategicCoordinationScope, constraintIngestionScope),
	}
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *scdpb.MakeDssReportRequest) (*scdpb.ErrorReport, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Not yet implemented")
}
