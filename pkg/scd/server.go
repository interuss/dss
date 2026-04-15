package scd

import (
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"
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
	Store             scdstore.Store
	DSSReportHandler  ReceivedReportHandler
	AllowHTTPBaseUrls bool
}
