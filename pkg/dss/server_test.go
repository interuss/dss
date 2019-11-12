package dss

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/dss/auth"
	"github.com/interuss/dss/pkg/dss/geo"
	"github.com/interuss/dss/pkg/dss/geo/testdata"
	"github.com/interuss/dss/pkg/dss/models"
	dspb "github.com/interuss/dss/pkg/dssproto"
	dsserr "github.com/interuss/dss/pkg/errors"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mustTimestamp(ts *tspb.Timestamp) *time.Time {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		panic(err)
	}
	return &t
}

func mustGeoPolygonToCellIDs(p *dspb.GeoPolygon) s2.CellUnion {
	cells, err := geo.GeoPolygonToCellIDs(p)
	if err != nil {
		panic(err)
	}
	return cells
}

type mockStore struct {
	mock.Mock
}

func (ms *mockStore) Close() error {
	return ms.Called().Error(0)
}

func (ms *mockStore) InsertSubscription(ctx context.Context, s models.Subscription) (*models.Subscription, error) {
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

func (ms *mockStore) InsertISA(ctx context.Context, isa models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error) {
	args := ms.Called(ctx, isa)
	return args.Get(0).(*models.IdentificationServiceArea), args.Get(1).([]*models.Subscription), args.Error(2)
}

func (ms *mockStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*models.IdentificationServiceArea, error) {
	args := ms.Called(ctx, cells, earliest, latest)
	return args.Get(0).([]*models.IdentificationServiceArea), args.Error(1)
}

func TestDeleteSubscription(t *testing.T) {
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

			_, err := s.DeleteSubscription(ctx, &dspb.DeleteSubscriptionRequest{
				Id: r.id.String(),
			})
			require.Equal(t, r.err, err)
			require.True(t, store.AssertExpectations(t))
		})
	}
}

func TestCreateSubscription(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")

	for _, r := range []struct {
		name             string
		id               models.ID
		callbacks        *dspb.SubscriptionCallbacks
		extents          *dspb.Volume4D
		wantSubscription models.Subscription
		wantErr          error
	}{
		{
			name: "success",
			id:   models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &dspb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: testdata.LoopVolume4D,
			wantSubscription: models.Subscription{
				ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				Owner:      "foo",
				Url:        "https://example.com",
				StartTime:  mustTimestamp(testdata.LoopVolume4D.GetTimeStart()),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.GetTimeEnd()),
				AltitudeHi: &testdata.LoopVolume3D.AltitudeHi,
				AltitudeLo: &testdata.LoopVolume3D.AltitudeLo,
				Cells:      mustGeoPolygonToCellIDs(testdata.LoopPolygon),
			},
		},
		{
			name: "missing-extents",
			id:   models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &dspb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			wantErr: dsserr.BadRequest("missing required extents"),
		},
		{
			name: "missing-extents-spatial-volume",
			id:   models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &dspb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: &dspb.Volume4D{},
			wantErr: dsserr.BadRequest("bad extents: missing required spatial_volume"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &dspb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: &dspb.Volume4D{
				SpatialVolume: &dspb.Volume3D{},
			},
			wantErr: dsserr.BadRequest("bad extents: spatial_volume missing required footprint"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &dspb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: &dspb.Volume4D{
				SpatialVolume: &dspb.Volume3D{
					Footprint: &dspb.GeoPolygon{},
				},
			},
			wantErr: dsserr.BadRequest("bad extents: not enough points in polygon"),
		},
		{
			name:    "missing-callbacks",
			id:      models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			wantErr: dsserr.BadRequest("missing required callbacks"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			if r.wantErr == nil {
				store.On("SearchISAs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*models.IdentificationServiceArea(nil), nil)
				store.On("InsertSubscription", mock.Anything, r.wantSubscription).Return(
					&r.wantSubscription, nil,
				)
			}
			s := &Server{
				Store: store,
			}

			_, err := s.CreateSubscription(ctx, &dspb.CreateSubscriptionRequest{
				Id: r.id.String(),
				Params: &dspb.CreateSubscriptionParameters{
					Callbacks: r.callbacks,
					Extents:   r.extents,
				},
			})
			require.Equal(t, r.wantErr, err)
			require.True(t, store.AssertExpectations(t))
		})
	}
}

func TestCreateSubscriptionResponseIncludesISAs(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")

	isas := []*models.IdentificationServiceArea{
		&models.IdentificationServiceArea{
			ID:    models.ID("8265221b-9528-4d45-900d-59a148e13850"),
			Owner: models.Owner("me-myself-and-i"),
			Url:   "https://no/place/like/home",
		},
	}

	cells := mustGeoPolygonToCellIDs(testdata.LoopPolygon)
	sub := models.Subscription{
		ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
		Owner:      "foo",
		Url:        "https://example.com",
		StartTime:  mustTimestamp(testdata.LoopVolume4D.GetTimeStart()),
		EndTime:    mustTimestamp(testdata.LoopVolume4D.GetTimeEnd()),
		AltitudeHi: &testdata.LoopVolume3D.AltitudeHi,
		AltitudeLo: &testdata.LoopVolume3D.AltitudeLo,
		Cells:      cells,
	}

	store := &mockStore{}
	store.On("SearchISAs", mock.Anything, cells, mock.Anything, mock.Anything).Return(isas, nil)
	store.On("InsertSubscription", mock.Anything, sub).Return(&sub, nil)
	s := &Server{
		Store: store,
	}

	resp, err := s.CreateSubscription(ctx, &dspb.CreateSubscriptionRequest{
		Id: sub.ID.String(),
		Params: &dspb.CreateSubscriptionParameters{
			Callbacks: &dspb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: sub.Url,
			},
			Extents: testdata.LoopVolume4D,
		},
	})
	require.Nil(t, err)
	require.True(t, store.AssertExpectations(t))

	require.Equal(t, []*dspb.IdentificationServiceArea{
		&dspb.IdentificationServiceArea{
			FlightsUrl: "https://no/place/like/home",
			Id:         "8265221b-9528-4d45-900d-59a148e13850",
			Owner:      "me-myself-and-i",
		},
	}, resp.ServiceAreas)
}

func TestGetSubscription(t *testing.T) {
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

			_, err := s.GetSubscription(context.Background(), &dspb.GetSubscriptionRequest{
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

func TestSearchSubscriptions(t *testing.T) {
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
	resp, err := s.SearchSubscriptions(ctx, &dspb.SearchSubscriptionsRequest{
		Area: testdata.Loop,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Subscriptions, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestCreateISA(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")

	for _, r := range []struct {
		name       string
		id         models.ID
		extents    *dspb.Volume4D
		flightsUrl string
		wantISA    *models.IdentificationServiceArea
		wantErr    error
	}{
		{
			name:       "success",
			id:         models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    testdata.LoopVolume4D,
			flightsUrl: "https://example.com",
			wantISA: &models.IdentificationServiceArea{
				ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				Url:        "https://example.com",
				Owner:      "foo",
				Cells:      mustGeoPolygonToCellIDs(testdata.LoopPolygon),
				StartTime:  mustTimestamp(testdata.LoopVolume4D.GetTimeStart()),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.GetTimeEnd()),
				AltitudeHi: &testdata.LoopVolume3D.AltitudeHi,
				AltitudeLo: &testdata.LoopVolume3D.AltitudeLo,
			},
		},
		{
			name:       "missing-extents",
			id:         models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			flightsUrl: "https://example.com",
			wantErr:    dsserr.BadRequest("missing required extents"),
		},
		{
			name:       "missing-extents-spatial-volume",
			id:         models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    &dspb.Volume4D{},
			flightsUrl: "https://example.com",
			wantErr:    dsserr.BadRequest("bad extents: missing required spatial_volume"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &dspb.Volume4D{
				SpatialVolume: &dspb.Volume3D{},
			},
			flightsUrl: "https://example.com",
			wantErr:    dsserr.BadRequest("bad extents: spatial_volume missing required footprint"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &dspb.Volume4D{
				SpatialVolume: &dspb.Volume3D{
					Footprint: &dspb.GeoPolygon{},
				},
			},
			flightsUrl: "https://example.com",
			wantErr:    dsserr.BadRequest("bad extents: not enough points in polygon"),
		},
		{
			name:    "missing-flights-url",
			id:      models.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			wantErr: dsserr.BadRequest("missing required flights_url"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			if r.wantISA != nil {
				store.On("InsertISA", mock.Anything, *r.wantISA).Return(
					r.wantISA, []*models.Subscription(nil), nil)
			}
			s := &Server{
				Store: store,
			}

			_, err := s.CreateIdentificationServiceArea(ctx, &dspb.CreateIdentificationServiceAreaRequest{
				Id: r.id.String(),
				Params: &dspb.CreateIdentificationServiceAreaParameters{
					Extents:    r.extents,
					FlightsUrl: r.flightsUrl,
				},
			})
			require.Equal(t, r.wantErr, err)
			require.True(t, store.AssertExpectations(t))
		})
	}
}

func TestDeleteIdentificationServiceAreaRequiresOwnerInContext(t *testing.T) {
	var (
		id = uuid.New().String()
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

func TestDeleteIdentificationServiceArea(t *testing.T) {
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
	resp, err := s.DeleteIdentificationServiceArea(ctx, &dspb.DeleteIdentificationServiceAreaRequest{
		Id: id.String(),
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Subscribers, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestSearchIdentificationServiceAreas(t *testing.T) {
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
	resp, err := s.SearchIdentificationServiceAreas(ctx, &dspb.SearchIdentificationServiceAreasRequest{
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
