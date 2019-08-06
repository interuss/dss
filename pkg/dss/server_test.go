package dss

import (
	"context"
	"errors"
	"testing"

	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo/testdata"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	mock.Mock
}

func (ms *mockStore) Close() error {
	return ms.Called().Error(0)
}

func (ms *mockStore) GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error) {
	args := ms.Called(ctx, id)
	return args.Get(0).(*dspb.Subscription), args.Error(1)
}

func (ms *mockStore) DeleteSubscription(ctx context.Context, id, version string) (*dspb.Subscription, error) {
	args := ms.Called(ctx, id)
	return args.Get(0).(*dspb.Subscription), args.Error(1)
}

func (ms *mockStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*dspb.Subscription, error) {
	args := ms.Called(ctx, cells, owner)
	return args.Get(0).([]*dspb.Subscription), args.Error(1)
}

func (ms *mockStore) DeleteIdentificationServiceArea(ctx context.Context, id string, owner string) (*dspb.IdentificationServiceArea, []*dspb.SubscriberToNotify, error) {
	args := ms.Called(ctx, id, owner)
	return args.Get(0).(*dspb.IdentificationServiceArea), args.Get(1).([]*dspb.SubscriberToNotify), args.Error(2)
}

func TestDeleteSubscriptionCallsIntoMockStore(t *testing.T) {
	for _, r := range []struct {
		name         string
		id           string
		subscription *dspb.Subscription
		err          error
	}{
		{
			name:         "subscription-is-returned-if-returned-from-store",
			id:           uuid.NewV4().String(),
			subscription: &dspb.Subscription{},
		},
		{
			name: "error-is-returned-if-returned-from-store",
			id:   uuid.NewV4().String(),
			err:  errors.New("failed to look up subscription for ID"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			store.On("DeleteSubscription", mock.Anything, r.id).Return(
				r.subscription, r.err,
			)
			s := &Server{
				Store: store,
			}

			response, err := s.DeleteSubscription(context.Background(), &dspb.DeleteSubscriptionRequest{
				Id: r.id,
			})
			require.Equal(t, r.err, err)
			require.EqualValues(t, r.subscription, response.GetSubscription())
			require.True(t, store.AssertExpectations(t))
		})
	}
}

func TestGetSubscriptionCallsIntoMockStore(t *testing.T) {
	for _, r := range []struct {
		name         string
		id           string
		subscription *dspb.Subscription
		err          error
	}{
		{
			name:         "subscription-is-returned-if-returned-from-store",
			id:           uuid.NewV4().String(),
			subscription: &dspb.Subscription{},
		},
		{
			name: "error-is-returned-if-returned-from-store",
			id:   uuid.NewV4().String(),
			err:  errors.New("failed to look up subscription for ID"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			store.On("GetSubscription", mock.Anything, r.id).Return(
				r.subscription, r.err,
			)
			s := &Server{
				Store: store,
			}

			response, err := s.GetSubscription(context.Background(), &dspb.GetSubscriptionRequest{
				Id: r.id,
			})
			require.Equal(t, r.err, err)
			require.EqualValues(t, r.subscription, response.GetSubscription())
			require.True(t, store.AssertExpectations(t))
		})
	}
}

func TestSearchSubscriptionsFailsIfOwnerMissingFromContext(t *testing.T) {
	var (
		ctx = context.Background()
		ms  = &mockStore{}
		s   = &Server{
			Store: &mockStore{},
		}
	)

	_, err := s.SearchSubscriptions(ctx, &dspb.SearchSubscriptionsRequest{
		Area: testdata.Loop,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestSearchSubscriptionsFailsForInvalidArea(t *testing.T) {
	var (
		ctx = auth.ContextWithOwner(context.Background(), "foo")
		ms  = &mockStore{}
		s   = &Server{
			Store: &mockStore{},
		}
	)

	_, err := s.SearchSubscriptions(ctx, &dspb.SearchSubscriptionsRequest{
		Area: testdata.LoopWithOddNumberOfCoordinates,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestSearchSubscriptionsCallsIntoStore(t *testing.T) {
	var (
		ctx = auth.ContextWithOwner(context.Background(), "foo")
		ms  = &mockStore{}
		s   = &Server{
			Store: ms,
		}
	)

	ms.On("SearchSubscriptions", mock.Anything, mock.Anything, "foo").Return(
		[]*dspb.Subscription{
			{
				Id:    uuid.NewV4().String(),
				Owner: "me-myself-and-i",
				Callbacks: &dspb.SubscriptionCallbacks{
					IdentificationServiceAreaUrl: "https://no/place/like/home",
				},
				NotificationIndex: 42,
			},
		}, error(nil),
	)
	resp, err := s.SearchSubscriptions(ctx, &dspb.SearchSubscriptionsRequest{
		Area: testdata.Loop,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Subscriptions, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestDeleteIdentificationServiceAreaRequiresOwnerInContext(t *testing.T) {
	var (
		id = uuid.NewV4().String()
		ms = &mockStore{}
		s  = &Server{
			Store: ms,
		}
	)

	_, err := s.DeleteIdentificationServiceArea(context.Background(), &dspb.DeleteIdentificationServiceAreaRequest{
		Id: id,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestDeleteIdentificationServiceAreaCallsIntoStore(t *testing.T) {
	var (
		id  = uuid.NewV4().String()
		ctx = auth.ContextWithOwner(context.Background(), "foo")
		ms  = &mockStore{}
		s   = &Server{
			Store: ms,
		}
	)

	ms.On("DeleteIdentificationServiceArea", ctx, id, "foo").Return(
		&dspb.IdentificationServiceArea{
			Id:         id,
			Owner:      "me-myself-and-i",
			FlightsUrl: "https://no/place/like/home",
			Extents: &dspb.Volume4D{
				TimeStart: ptypes.TimestampNow(),
				TimeEnd:   ptypes.TimestampNow(),
			},
		},
		[]*dspb.SubscriberToNotify{
			&dspb.SubscriberToNotify{
				Subscriptions: []*dspb.SubscriptionState{
					&dspb.SubscriptionState{
						NotificationIndex: 42,
					},
				},
				Url: "https://no/place/like/home",
			},
		}, error(nil),
	)
	resp, err := s.DeleteIdentificationServiceArea(ctx, &dspb.DeleteIdentificationServiceAreaRequest{
		Id: id,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Subscribers, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestDefaultRegionCovererProducesResults(t *testing.T) {
	cover, err := geo.AreaToCellIDs(testdata.Loop)
	require.NoError(t, err)
	require.NotNil(t, cover)
}
