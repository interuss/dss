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
	constraintProcessingScope    = "utm.constraint_processing"
	conformanceMonitoringSAScope = "utm.conformance_monitoring_sa"
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
	return map[auth.Operation]auth.KeyClaimedScopesValidator{
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/CreateConstraintReference":        auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/CreateOperationalIntentReference": auth.RequireAnyScope(strategicCoordinationScope, conformanceMonitoringSAScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/CreateSubscription":               auth.RequireAnyScope(strategicCoordinationScope, constraintProcessingScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteConstraintReference":        auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteOperationalIntentReference": auth.RequireAnyScope(strategicCoordinationScope, conformanceMonitoringSAScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteSubscription":               auth.RequireAnyScope(strategicCoordinationScope, constraintProcessingScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetConstraintReference":           auth.RequireAnyScope(constraintManagementScope, constraintProcessingScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetOperationalIntentReference":    auth.RequireAnyScope(strategicCoordinationScope, conformanceMonitoringSAScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetSubscription":                  auth.RequireAnyScope(strategicCoordinationScope, constraintProcessingScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetUssAvailability":               auth.RequireAnyScope(strategicCoordinationScope, availabilityArbitrationScope, conformanceMonitoringSAScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/MakeDssReport":                    auth.RequireAnyScope(strategicCoordinationScope, constraintProcessingScope, constraintManagementScope, availabilityArbitrationScope, conformanceMonitoringSAScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QueryConstraintReferences":        auth.RequireAnyScope(constraintProcessingScope, constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QueryOperationalIntentReferences": auth.RequireAnyScope(strategicCoordinationScope, conformanceMonitoringSAScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QuerySubscriptions":               auth.RequireAnyScope(strategicCoordinationScope, constraintProcessingScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/SetUssAvailability":               auth.RequireAnyScope(availabilityArbitrationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/UpdateConstraintReference":        auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/UpdateOperationalIntentReference": auth.RequireAnyScope(strategicCoordinationScope, conformanceMonitoringSAScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/UpdateSubscription":               auth.RequireAnyScope(strategicCoordinationScope, constraintProcessingScope),
	}
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *scdpb.MakeDssReportRequest) (*scdpb.ErrorReport, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Not yet implemented")
}
