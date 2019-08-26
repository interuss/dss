package dss

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo/testdata"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	mock.Mock
}

func (ms *mockStore) Close() error {
	return ms.Called().Error(0)
}

func (ms *mockStore) InsertSubscription(ctx context.Context, s *models.Subscription) (*models.Subscription, error) {
	args := ms.Called(ctx, s)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (ms *mockStore) GetSubscription(ctx context.Context, id models.ID) (*models.Subscription, error) {
	args := ms.Called(ctx, id)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (ms *mockStore) DeleteSubscription(ctx context.Context, id models.ID, owner models.Owner, version *models.Version) (*models.Subscription, error) {
	args := ms.Called(ctx, id, owner, version)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (ms *mockStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner models.Owner) ([]*models.Subscription, error) {
	args := ms.Called(ctx, cells, owner)
	return args.Get(0).([]*models.Subscription), args.Error(1)
}

func (ms *mockStore) GetISA(ctx context.Context, id models.ID) (*models.IdentificationServiceArea, error) {
	args := ms.Called(ctx, id)
	return args.Get(0).(*models.IdentificationServiceArea), args.Error(1)
}

func (ms *mockStore) DeleteISA(ctx context.Context, id models.ID, owner models.Owner, version *models.Version) (*models.IdentificationServiceArea, []*models.Subscription, error) {
	args := ms.Called(ctx, id, owner, version)
	return args.Get(0).(*models.IdentificationServiceArea), args.Get(1).([]*models.Subscription), args.Error(2)
}

func (ms *mockStore) InsertISA(ctx context.Context, isa *models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error) {
	args := ms.Called(ctx, isa)
	return args.Get(0).(*models.IdentificationServiceArea), args.Get(1).([]*models.Subscription), args.Error(2)
}

func (ms *mockStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*models.IdentificationServiceArea, error) {
	args := ms.Called(ctx, cells, earliest, latest)
	return args.Get(0).([]*models.IdentificationServiceArea), args.Error(1)
}

func TestDeleteSubscriptionCallsIntoMockStore(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")

	for _, r := range []struct {
		name         string
		id           models.ID
		subscription *models.Subscription
		err          error
	}{
		{
			name:         "subscription-is-returned-if-returned-from-store",
			id:           models.ID(uuid.New().String()),
			subscription: &models.Subscription{},
		},
		{
			name: "error-is-returned-if-returned-from-store",
			id:   models.ID(uuid.New().String()),
			err:  errors.New("failed to look up subscription for ID"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			store.On("DeleteSubscription", mock.Anything, r.id, mock.Anything, mock.Anything).Return(
				r.subscription, r.err,
			)
			s := &Server{
				Store: store,
			}

			_, err := s.DeleteV1DssSubscriptionsIdVersion(ctx, &dspb.DeleteV1DssSubscriptionsIdVersionRequest{
				Id: r.id.String(),
			})
			require.Equal(t, r.err, err)
			require.True(t, store.AssertExpectations(t))
		})
	}
}

func TestGetSubscriptionCallsIntoMockStore(t *testing.T) {
	for _, r := range []struct {
		name         string
		id           models.ID
		subscription *models.Subscription
		err          error
	}{
		{
			name:         "subscription-is-returned-if-returned-from-store",
			id:           models.ID(uuid.New().String()),
			subscription: &models.Subscription{},
		},
		{
			name: "error-is-returned-if-returned-from-store",
			id:   models.ID(uuid.New().String()),
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

			_, err := s.GetV1DssSubscriptionsId(context.Background(), &dspb.GetV1DssSubscriptionsIdRequest{
				Id: r.id.String(),
			})
			require.Equal(t, r.err, err)
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

	_, err := s.GetV1DssSubscriptions(ctx, &dspb.GetV1DssSubscriptionsRequest{
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

	_, err := s.GetV1DssSubscriptions(ctx, &dspb.GetV1DssSubscriptionsRequest{
		Area: testdata.LoopWithOddNumberOfCoordinates,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestSearchSubscriptionsCallsIntoStore(t *testing.T) {
	var (
		owner = models.Owner("foo")
		ctx   = auth.ContextWithOwner(context.Background(), owner)
		ms    = &mockStore{}
		s     = &Server{
			Store: ms,
		}
	)

	ms.On("SearchSubscriptions", mock.Anything, mock.Anything, owner).Return(
		[]*models.Subscription{
			{
				ID:                models.ID(uuid.New().String()),
				Owner:             owner,
				Url:               "https://no/place/like/home",
				NotificationIndex: 42,
			},
		}, error(nil),
	)
	resp, err := s.GetV1DssSubscriptions(ctx, &dspb.GetV1DssSubscriptionsRequest{
		Area: testdata.Loop,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Subscriptions, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestDeleteIdentificationServiceAreaRequiresOwnerInContext(t *testing.T) {
	var (
		id = uuid.New().String()
		ms = &mockStore{}
		s  = &Server{
			Store: ms,
		}
	)

	_, err := s.DeleteV1DssIdentificationServiceAreasIdVersion(context.Background(), &dspb.DeleteV1DssIdentificationServiceAreasIdVersionRequest{
		Id: id,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestDeleteIdentificationServiceAreaCallsIntoStore(t *testing.T) {
	var (
		owner = models.Owner("foo")
		id    = models.ID(uuid.New().String())
		ctx   = auth.ContextWithOwner(context.Background(), owner)
		ms    = &mockStore{}
		s     = &Server{
			Store: ms,
		}
	)

	ms.On("DeleteISA", ctx, id, owner, mock.Anything).Return(
		&models.IdentificationServiceArea{
			ID:    models.ID(id),
			Owner: models.Owner("me-myself-and-i"),
			Url:   "https://no/place/like/home",
		},
		[]*models.Subscription{
			{
				NotificationIndex: 42,
				Url:               "https://no/place/like/home",
			},
		}, error(nil),
	)
	resp, err := s.DeleteV1DssIdentificationServiceAreasIdVersion(ctx, &dspb.DeleteV1DssIdentificationServiceAreasIdVersionRequest{
		Id: id.String(),
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Subscribers, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestSearchIdentificationServiceAreasCallsIntoStore(t *testing.T) {
	var (
		ctx = context.Background()
		ms  = &mockStore{}
		s   = &Server{
			Store: ms,
		}
	)

	ms.On("SearchISAs", ctx, mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).Return(
		[]*models.IdentificationServiceArea{
			{
				ID:    models.ID(uuid.New().String()),
				Owner: models.Owner("me-myself-and-i"),
				Url:   "https://no/place/like/home",
			},
		}, error(nil),
	)
	resp, err := s.GetV1DssIdentificationServiceAreas(ctx, &dspb.GetV1DssIdentificationServiceAreasRequest{
		Area: testdata.Loop,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.ServiceAreas, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestDefaultRegionCovererProducesResults(t *testing.T) {
	cover, err := geo.AreaToCellIDs(testdata.Loop)
	require.NoError(t, err)
	require.NotNil(t, cover)
}
