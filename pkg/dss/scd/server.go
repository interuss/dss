package scd

import (
	"context"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/dss/auth"
	"github.com/interuss/dss/pkg/dss/geo"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	scdstore "github.com/interuss/dss/pkg/dss/scd/store"
	dsserr "github.com/interuss/dss/pkg/errors"
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
func (a *Server) AuthScopes() map[auth.Operation][]auth.Scope {
	// TODO: replace with correct scopes
	return map[auth.Operation][]auth.Scope{
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteConstraintReference": {constraintManagementScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteOperationReference":  {strategicCoordinationScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/DeleteSubscription":        {strategicCoordinationScope, constraintConsumptionScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetConstraintReference":    {strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetOperationReference":     {strategicCoordinationScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/GetSubscription":           {strategicCoordinationScope, constraintConsumptionScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/MakeDssReport":             {strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/PutConstraintReference":    {constraintManagementScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/PutOperationReference":     {strategicCoordinationScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/PutSubscription":           {strategicCoordinationScope, constraintConsumptionScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QueryConstraintReferences": {strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/QuerySubscriptions":        {strategicCoordinationScope, constraintConsumptionScope},
		"/scdpb.UTMAPIUSSDSSAndUSSUSSService/SearchOperationReferences": {strategicCoordinationScope},
	}
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *scdpb.MakeDssReportRequest) (*scdpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
