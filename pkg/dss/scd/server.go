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
	//"DeleteConstraintReference": {readISAScope}, //{constraintManagementScope},
	//"DeleteOperationReference":  {readISAScope}, //{strategicCoordinationScope},
	// TODO: De-duplicate operation names
	//"DeleteSubscription":               {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"GetConstraintReference": {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
	//"GetOperationReference":  {readISAScope}, //{strategicCoordinationScope},
	//"GetSubscription":                  {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"MakeDssReport":          {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
	//"PutConstraintReference": {readISAScope}, //{constraintManagementScope},
	//"PutOperationReference":  {readISAScope}, //{strategicCoordinationScope},
	//"PutSubscription":                  {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"QueryConstraintReferences": {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
	//"QuerySubscriptions":        {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"SearchOperationReferences": {readISAScope}, //{strategicCoordinationScope},
	return nil
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *scdpb.MakeDssReportRequest) (*scdpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
