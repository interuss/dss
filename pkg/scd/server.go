package scd

import (
	"context"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"
)

const (
	strategicCoordinationScope = "utm.strategic_coordination"
	constraintManagementScope  = "utm.constraint_management"
	constraintConsumptionScope = "utm.constraint_consumption"
)

func dssErrorOfAreaError(err error) error {
	switch err.(type) {
	case *geo.ErrAreaTooLarge:
		return dsserr.AreaTooLarge(err.Error())
	default:
		return dsserr.BadRequest(fmt.Sprintf("bad area: %s", err))
	}
}

func makeSubscribersToNotify(subscriptions []*scdmodels.Subscription) []*scdpb.SubscriberToNotify {
	result := []*scdpb.SubscriberToNotify{}

	subscriptionsByURL := map[string][]*scdpb.SubscriptionState{}
	for _, sub := range subscriptions {
		subState := &scdpb.SubscriptionState{
			SubscriptionId:    sub.ID.String(),
			NotificationIndex: int32(sub.NotificationIndex),
		}
		subscriptionsByURL[sub.BaseURL] = append(subscriptionsByURL[sub.BaseURL], subState)
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
	Store   scdstore.Store
	Timeout time.Duration
}

// AuthScopes returns a map of endpoint to required Oauth scope.
func (a *Server) AuthScopes() map[auth.Operation]auth.KeyClaimedScopesValidator {
	// TODO: replace with correct scopes
	return map[auth.Operation]auth.KeyClaimedScopesValidator{
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteConstraintReference": auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteOperationReference":  auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteSubscription":        auth.RequireAnyScope(strategicCoordinationScope, constraintConsumptionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetConstraintReference":    auth.RequireAnyScope(strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetOperationReference":     auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetSubscription":           auth.RequireAnyScope(strategicCoordinationScope, constraintConsumptionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/MakeDssReport":             auth.RequireAnyScope(strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/PutConstraintReference":    auth.RequireAnyScope(constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/PutOperationReference":     auth.RequireAnyScope(strategicCoordinationScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/PutSubscription":           auth.RequireAnyScope(strategicCoordinationScope, constraintConsumptionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QueryConstraintReferences": auth.RequireAnyScope(strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QuerySubscriptions":        auth.RequireAnyScope(strategicCoordinationScope, constraintConsumptionScope),
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/SearchOperationReferences": auth.RequireAnyScope(strategicCoordinationScope),
	}
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *scdpb.MakeDssReportRequest) (*scdpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
