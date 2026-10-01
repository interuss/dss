package scd

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/stretchr/testify/require"
)

func setUSSAvailability(t *testing.T, srv *Server, ctx context.Context, ussID string, state restapi.UssAvailabilityState) {
	t.Helper()
	arbitrator := "availability_arbitrator"
	arbScopes := []string{string(restapi.UtmAvailabilityArbitrationScope)}

	getResp := srv.GetUssAvailability(ctx, &restapi.GetUssAvailabilityRequest{
		UssId: ussID,
		Auth: api.AuthorizationResult{
			ClientID: &arbitrator,
			Scopes:   arbScopes,
		},
	})
	require.NotNil(t, getResp.Response200)

	setResp := srv.SetUssAvailability(ctx, &restapi.SetUssAvailabilityRequest{
		UssId: ussID,
		Auth: api.AuthorizationResult{
			ClientID: &arbitrator,
			Scopes:   arbScopes,
		},
		Body: &restapi.SetUssAvailabilityStatusParameters{
			Availability: state,
			OldVersion:   getResp.Response200.Version,
		},
	})
	require.NotNil(t, setResp.Response200)
	require.Equal(t, state, setResp.Response200.Status.Availability)
}

func createSubscriber(t *testing.T, srv *Server, ctx context.Context, subscriber string, vol restapi.Volume4D) restapi.SubscriptionID {
	t.Helper()
	subID := restapi.SubscriptionID(uuid.NewString())
	notifyOI := true
	notifyCon := false
	resp := srv.CreateSubscription(ctx, &restapi.CreateSubscriptionRequest{
		Subscriptionid: subID,
		Auth: api.AuthorizationResult{
			ClientID: &subscriber,
			Scopes:   []string{string(restapi.UtmStrategicCoordinationScope)},
		},
		Body: &restapi.PutSubscriptionParameters{
			Extents:                     vol,
			UssBaseUrl:                  "https://subscriber.example.com/uss",
			NotifyForOperationalIntents: &notifyOI,
			NotifyForConstraints:        &notifyCon,
		},
	})
	require.NotNil(t, resp.Response200)
	require.Equal(t, restapi.SubscriptionNotificationIndex(0), resp.Response200.Subscription.NotificationIndex)
	return subID
}

func getSubscriptionNotificationIndex(t *testing.T, srv *Server, ctx context.Context, subscriber string, subID restapi.SubscriptionID) restapi.SubscriptionNotificationIndex {
	t.Helper()
	resp := srv.GetSubscription(ctx, &restapi.GetSubscriptionRequest{
		Subscriptionid: subID,
		Auth: api.AuthorizationResult{
			ClientID: &subscriber,
			Scopes:   []string{string(restapi.UtmStrategicCoordinationScope)},
		},
	})
	require.NotNil(t, resp.Response200)
	return resp.Response200.Subscription.NotificationIndex
}

// This test is AI-generated and has not been closely inspected by a human.
func TestOperationalIntentMutationsEnforceUSSAvailability(t *testing.T) {
	strategicScopes := []string{string(restapi.UtmStrategicCoordinationScope)}
	cmsaScopes := []string{
		string(restapi.UtmStrategicCoordinationScope),
		string(restapi.UtmConformanceMonitoringSaScope),
	}

	t.Run("create_rejected_when_down_and_allowed_when_unknown_or_normal", func(t *testing.T) {
		srv, ctx := newTestServer(t)
		vol := testVolume4D(timestamp.MustFromContext(ctx))
		manager := "uss_a"
		observer := "uss_b"

		observerSubID := createSubscriber(t, srv, ctx, observer, vol)

		// Unknown availability (default) allows creating an Accepted OIR.
		oiUnknownID := restapi.EntityID(uuid.NewString())
		createUnknownResp := srv.CreateOperationalIntentReference(ctx, &restapi.CreateOperationalIntentReferenceRequest{
			Entityid: oiUnknownID,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:    []restapi.Volume4D{vol},
				State:      restapi.OperationalIntentState_Accepted,
				UssBaseUrl: "https://uss-a.example.com/uss",
			},
		})
		require.NotNil(t, createUnknownResp.Response201)
		require.Nil(t, createUnknownResp.Response412)
		require.Equal(t, restapi.SubscriptionNotificationIndex(1), getSubscriptionNotificationIndex(t, srv, ctx, observer, observerSubID))

		// Mark manager Down: creating an Accepted OIR must fail with 412 and have no side effects.
		setUSSAvailability(t, srv, ctx, manager, restapi.UssAvailabilityState_Down)

		oiDownID := restapi.EntityID(uuid.NewString())
		key := restapi.Key{*createUnknownResp.Response201.OperationalIntentReference.Ovn}
		createDownResp := srv.CreateOperationalIntentReference(ctx, &restapi.CreateOperationalIntentReferenceRequest{
			Entityid: oiDownID,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:    []restapi.Volume4D{vol},
				State:      restapi.OperationalIntentState_Accepted,
				UssBaseUrl: "https://uss-a.example.com/uss",
				Key:        &key,
			},
		})
		require.NotNil(t, createDownResp.Response412)
		require.Nil(t, createDownResp.Response201)

		// Verify rejected OIR was not created and subscriber notification index did not advance.
		getDownOI := srv.GetOperationalIntentReference(ctx, &restapi.GetOperationalIntentReferenceRequest{
			Entityid: oiDownID,
			Auth: api.AuthorizationResult{
				ClientID: &observer,
				Scopes:   strategicScopes,
			},
		})
		require.NotNil(t, getDownOI.Response404)
		require.Equal(t, restapi.SubscriptionNotificationIndex(1), getSubscriptionNotificationIndex(t, srv, ctx, observer, observerSubID))

		// Restore manager to Normal: creation succeeds.
		setUSSAvailability(t, srv, ctx, manager, restapi.UssAvailabilityState_Normal)

		createNormalResp := srv.CreateOperationalIntentReference(ctx, &restapi.CreateOperationalIntentReferenceRequest{
			Entityid: oiDownID,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:    []restapi.Volume4D{vol},
				State:      restapi.OperationalIntentState_Accepted,
				UssBaseUrl: "https://uss-a.example.com/uss",
				Key:        &key,
			},
		})
		require.NotNil(t, createNormalResp.Response201)
		require.Nil(t, createNormalResp.Response412)
		require.Equal(t, restapi.SubscriptionNotificationIndex(2), getSubscriptionNotificationIndex(t, srv, ctx, observer, observerSubID))
	})

	t.Run("update_and_delete_restrictions_when_down", func(t *testing.T) {
		srv, ctx := newTestServer(t)
		vol := testVolume4D(timestamp.MustFromContext(ctx))
		manager := "uss_a"
		observer := "uss_b"

		observerSubID := createSubscriber(t, srv, ctx, observer, vol)

		// Create subscription for manager to support Activated / off-nominal states.
		managerSubIDStr := uuid.NewString()
		managerSubID := restapi.SubscriptionID(managerSubIDStr)
		managerSubEntityID := restapi.EntityID(managerSubIDStr)
		notifyOI := true
		notifyCon := false
		subResp := srv.CreateSubscription(ctx, &restapi.CreateSubscriptionRequest{
			Subscriptionid: managerSubID,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
			Body: &restapi.PutSubscriptionParameters{
				Extents:                     vol,
				UssBaseUrl:                  "https://uss-a.example.com/uss",
				NotifyForOperationalIntents: &notifyOI,
				NotifyForConstraints:        &notifyCon,
			},
		})
		require.NotNil(t, subResp.Response200)

		// Create an Accepted OIR while manager is Normal.
		setUSSAvailability(t, srv, ctx, manager, restapi.UssAvailabilityState_Normal)
		oiID := restapi.EntityID(uuid.NewString())
		createResp := srv.CreateOperationalIntentReference(ctx, &restapi.CreateOperationalIntentReferenceRequest{
			Entityid: oiID,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:        []restapi.Volume4D{vol},
				State:          restapi.OperationalIntentState_Accepted,
				UssBaseUrl:     "https://uss-a.example.com/uss",
				SubscriptionId: &managerSubEntityID,
			},
		})
		require.NotNil(t, createResp.Response201)
		currentOVN := *createResp.Response201.OperationalIntentReference.Ovn
		require.Equal(t, restapi.SubscriptionNotificationIndex(1), getSubscriptionNotificationIndex(t, srv, ctx, observer, observerSubID))

		// Mark manager Down.
		setUSSAvailability(t, srv, ctx, manager, restapi.UssAvailabilityState_Down)

		// 1. Update resulting in Accepted while Down -> 412.
		updateAcceptedResp := srv.UpdateOperationalIntentReference(ctx, &restapi.UpdateOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:        []restapi.Volume4D{vol},
				State:          restapi.OperationalIntentState_Accepted,
				UssBaseUrl:     "https://uss-a-updated.example.com/uss",
				SubscriptionId: &managerSubEntityID,
			},
		})
		require.NotNil(t, updateAcceptedResp.Response412)
		require.Nil(t, updateAcceptedResp.Response200)

		// 2. Update resulting in Activated while Down -> 412.
		updateActivatedResp := srv.UpdateOperationalIntentReference(ctx, &restapi.UpdateOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:        []restapi.Volume4D{vol},
				State:          restapi.OperationalIntentState_Activated,
				UssBaseUrl:     "https://uss-a.example.com/uss",
				SubscriptionId: &managerSubEntityID,
			},
		})
		require.NotNil(t, updateActivatedResp.Response412)
		require.Nil(t, updateActivatedResp.Response200)

		// 3. Delete while Down -> 412.
		deleteDownResp := srv.DeleteOperationalIntentReference(ctx, &restapi.DeleteOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
		})
		require.NotNil(t, deleteDownResp.Response412)
		require.Nil(t, deleteDownResp.Response200)

		// Verify OIR was not mutated or deleted, and subscriber notification index is still 1.
		getOI := srv.GetOperationalIntentReference(ctx, &restapi.GetOperationalIntentReferenceRequest{
			Entityid: oiID,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   strategicScopes,
			},
		})
		require.NotNil(t, getOI.Response200)
		require.Equal(t, restapi.EntityVersion(1), getOI.Response200.OperationalIntentReference.Version)
		require.Equal(t, restapi.OperationalIntentState_Accepted, getOI.Response200.OperationalIntentReference.State)
		require.Equal(t, restapi.OperationalIntentUssBaseURL("https://uss-a.example.com/uss"), getOI.Response200.OperationalIntentReference.UssBaseUrl)
		require.Equal(t, restapi.SubscriptionNotificationIndex(1), getSubscriptionNotificationIndex(t, srv, ctx, observer, observerSubID))

		// 4. Transition to Nonconforming and Contingent while Down is permitted with CMSA scope.
		updateNonconfResp := srv.UpdateOperationalIntentReference(ctx, &restapi.UpdateOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   cmsaScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:        []restapi.Volume4D{vol},
				State:          restapi.OperationalIntentState_Nonconforming,
				UssBaseUrl:     "https://uss-a.example.com/uss",
				SubscriptionId: &managerSubEntityID,
			},
		})
		require.NotNil(t, updateNonconfResp.Response200)
		require.Nil(t, updateNonconfResp.Response412)
		currentOVN = *updateNonconfResp.Response200.OperationalIntentReference.Ovn
		require.Equal(t, restapi.SubscriptionNotificationIndex(2), getSubscriptionNotificationIndex(t, srv, ctx, observer, observerSubID))

		// Transitioning from Nonconforming back to Activated while Down is still rejected with 412.
		reactivateDownResp := srv.UpdateOperationalIntentReference(ctx, &restapi.UpdateOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   cmsaScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:        []restapi.Volume4D{vol},
				State:          restapi.OperationalIntentState_Activated,
				UssBaseUrl:     "https://uss-a.example.com/uss",
				SubscriptionId: &managerSubEntityID,
			},
		})
		require.NotNil(t, reactivateDownResp.Response412)
		require.Nil(t, reactivateDownResp.Response200)

		// Transitioning to Contingent while Down is permitted.
		updateContingentResp := srv.UpdateOperationalIntentReference(ctx, &restapi.UpdateOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   cmsaScopes,
			},
			Body: &restapi.PutOperationalIntentReferenceParameters{
				Extents:        []restapi.Volume4D{vol},
				State:          restapi.OperationalIntentState_Contingent,
				UssBaseUrl:     "https://uss-a.example.com/uss",
				SubscriptionId: &managerSubEntityID,
			},
		})
		require.NotNil(t, updateContingentResp.Response200)
		require.Nil(t, updateContingentResp.Response412)
		currentOVN = *updateContingentResp.Response200.OperationalIntentReference.Ovn

		// Deleting a Contingent OIR while Down is still rejected with 412.
		deleteContingentDownResp := srv.DeleteOperationalIntentReference(ctx, &restapi.DeleteOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   cmsaScopes,
			},
		})
		require.NotNil(t, deleteContingentDownResp.Response412)
		require.Nil(t, deleteContingentDownResp.Response200)

		// Restore manager to Normal: deletion succeeds.
		setUSSAvailability(t, srv, ctx, manager, restapi.UssAvailabilityState_Normal)
		deleteNormalResp := srv.DeleteOperationalIntentReference(ctx, &restapi.DeleteOperationalIntentReferenceRequest{
			Entityid: oiID,
			Ovn:      currentOVN,
			Auth: api.AuthorizationResult{
				ClientID: &manager,
				Scopes:   cmsaScopes,
			},
		})
		require.NotNil(t, deleteNormalResp.Response200)
		require.Nil(t, deleteNormalResp.Response412)
	})
}
