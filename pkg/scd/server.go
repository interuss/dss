package scd

import (
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"
)

func makeSubscribersToNotify(subscriptions []*scdmodels.Subscription) []restapi.SubscriberToNotify {
	result := []restapi.SubscriberToNotify{}

	type uss struct {
		manager string
		baseUrl restapi.SubscriptionUssBaseURL
	}
	subscriptionsByUSS := map[uss][]restapi.SubscriptionState{}
	for _, sub := range subscriptions {
		subState := restapi.SubscriptionState{
			SubscriptionId:    restapi.SubscriptionID(sub.ID.String()),
			NotificationIndex: restapi.SubscriptionNotificationIndex(sub.NotificationIndex),
		}
		uss := uss{
			manager: sub.Manager.String(),
			baseUrl: restapi.SubscriptionUssBaseURL(sub.USSBaseURL),
		}
		subscriptionsByUSS[uss] = append(subscriptionsByUSS[uss], subState)
	}
	for uss, states := range subscriptionsByUSS {
		result = append(result, restapi.SubscriberToNotify{
			Manager:       uss.manager,
			UssBaseUrl:    uss.baseUrl,
			Subscriptions: states,
		})
	}

	return result
}

// Server implements scdv1.Implementation.
type Server struct {
	Store             scdstore.Store
	DSSReportHandler  ReceivedReportHandler
	AllowHTTPBaseUrls bool
}
