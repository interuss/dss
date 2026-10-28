package operations

import (
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/pb/scdpb"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// Registry maps operation IDs to their handlers
var Registry = map[string]dssstore.OperationHandler[repos.Repository]{}

// ProposalPayload wraps an scd payload in the payload of a proposal.
func ProposalPayload(payload *scdpb.Payload) *raftpb.Payload {
	return &raftpb.Payload{Payload: &raftpb.Payload_ScdPayload{ScdPayload: payload}}
}

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
