package actions

import (
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// Registry maps operation IDs to their handlers
var Registry = map[string]dssstore.OperationHandler[repos.Repository]{}

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
