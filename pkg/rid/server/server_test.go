package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/api/v1/ridpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/dss/pkg/geo/testdata"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var timeout = time.Second * 10

func mustTimestamp(ts *tspb.Timestamp) *time.Time {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		panic(err)
	}
	return &t
}

func mustPolygonToCellIDs(p *ridpb.GeoPolygon) s2.CellUnion {
	cells, err := dssmodels.GeoPolygonFromRIDProto(p).CalculateCovering()
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

func (ms *mockStore) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, s)
	return args.Get(0).(*ridmodels.Subscription), args.Error(1)
}

func (ms *mockStore) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id)
	return args.Get(0).(*ridmodels.Subscription), args.Error(1)
}

func (ms *mockStore) DeleteSubscription(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id, owner, version)
	return args.Get(0).(*ridmodels.Subscription), args.Error(1)
}

func (ms *mockStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, cells, owner)
	return args.Get(0).([]*ridmodels.Subscription), args.Error(1)
}

func (ms *mockStore) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id)
	return args.Get(0).(*ridmodels.IdentificationServiceArea), args.Error(1)
}

func (ms *mockStore) DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id, owner, version)
	return args.Get(0).(*ridmodels.IdentificationServiceArea), args.Get(1).([]*ridmodels.Subscription), args.Error(2)
}

func (ms *mockStore) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	args := ms.Called(ctx, isa)
	return args.Get(0).(*ridmodels.IdentificationServiceArea), args.Get(1).([]*ridmodels.Subscription), args.Error(2)
}

func (ms *mockStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, cells, earliest, latest)
	return args.Get(0).([]*ridmodels.IdentificationServiceArea), args.Error(1)
}

func TestDeleteSubscription(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")
	version, _ := dssmodels.VersionFromString("bar")

	for _, r := range []struct {
		name         string
		id           dssmodels.ID
		version      *dssmodels.Version
		subscription *ridmodels.Subscription
		err          error
	}{
		{
			name:         "subscription-is-returned-if-returned-from-store",
			id:           dssmodels.ID(uuid.New().String()),
			version:      version,
			subscription: &ridmodels.Subscription{},
		},
		{
			name:    "error-is-returned-if-returned-from-store",
			id:      dssmodels.ID(uuid.New().String()),
			version: version,
			err:     errors.New("failed to look up subscription for ID"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			store.On("DeleteSubscription", mock.Anything, r.id, mock.Anything, r.version).Return(
				r.subscription, r.err,
			)
			s := &Server{
				Store: store,
			}

			_, err := s.DeleteSubscription(ctx, &ridpb.DeleteSubscriptionRequest{
				Id: r.id.String(), Version: r.version.String(),
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
		id               dssmodels.ID
		callbacks        *ridpb.SubscriptionCallbacks
		extents          *ridpb.Volume4D
		wantSubscription *ridmodels.Subscription
		wantErr          error
	}{
		{
			name: "success",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &ridpb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: testdata.LoopVolume4D,
			wantSubscription: &ridmodels.Subscription{
				ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				Owner:      "foo",
				URL:        "https://example.com",
				StartTime:  mustTimestamp(testdata.LoopVolume4D.GetTimeStart()),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.GetTimeEnd()),
				AltitudeHi: &testdata.LoopVolume3D.AltitudeHi,
				AltitudeLo: &testdata.LoopVolume3D.AltitudeLo,
				Cells:      mustPolygonToCellIDs(testdata.LoopPolygon),
			},
		},
		{
			name: "missing-extents",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &ridpb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			wantErr: dsserr.BadRequest("missing required extents"),
		},
		{
			name: "missing-extents-spatial-volume",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &ridpb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: &ridpb.Volume4D{},
			wantErr: dsserr.BadRequest("bad extents: missing required spatial_volume"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &ridpb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: &ridpb.Volume4D{
				SpatialVolume: &ridpb.Volume3D{},
			},
			wantErr: dsserr.BadRequest("bad extents: spatial_volume missing required footprint"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &ridpb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: &ridpb.Volume4D{
				SpatialVolume: &ridpb.Volume3D{
					Footprint: &ridpb.GeoPolygon{},
				},
			},
			wantErr: dsserr.BadRequest("bad extents: not enough points in polygon"),
		},
		{
			name:    "missing-callbacks",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			wantErr: dsserr.BadRequest("missing required callbacks"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			if r.wantErr == nil {
				store.On("SearchISAs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*ridmodels.IdentificationServiceArea(nil), nil)
				store.On("InsertSubscription", mock.Anything, r.wantSubscription).Return(
					r.wantSubscription, nil,
				)
			}
			s := &Server{
				Store: store,
			}

			_, err := s.CreateSubscription(ctx, &ridpb.CreateSubscriptionRequest{
				Id: r.id.String(),
				Params: &ridpb.CreateSubscriptionParameters{
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

	isas := []*ridmodels.IdentificationServiceArea{
		{
			ID:    dssmodels.ID("8265221b-9528-4d45-900d-59a148e13850"),
			Owner: dssmodels.Owner("me-myself-and-i"),
			URL:   "https://no/place/like/home",
		},
	}

	cells := mustPolygonToCellIDs(testdata.LoopPolygon)
	sub := &ridmodels.Subscription{
		ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
		Owner:      "foo",
		URL:        "https://example.com",
		StartTime:  mustTimestamp(testdata.LoopVolume4D.GetTimeStart()),
		EndTime:    mustTimestamp(testdata.LoopVolume4D.GetTimeEnd()),
		AltitudeHi: &testdata.LoopVolume3D.AltitudeHi,
		AltitudeLo: &testdata.LoopVolume3D.AltitudeLo,
		Cells:      cells,
	}

	store := &mockStore{}
	store.On("SearchISAs", mock.Anything, cells, mock.Anything, mock.Anything).Return(isas, nil)
	store.On("InsertSubscription", mock.Anything, sub).Return(sub, nil)
	s := &Server{
		Store: store,
	}

	resp, err := s.CreateSubscription(ctx, &ridpb.CreateSubscriptionRequest{
		Id: sub.ID.String(),
		Params: &ridpb.CreateSubscriptionParameters{
			Callbacks: &ridpb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: sub.URL,
			},
			Extents: testdata.LoopVolume4D,
		},
	})
	require.Nil(t, err)
	require.True(t, store.AssertExpectations(t))

	require.Equal(t, []*ridpb.IdentificationServiceArea{
		{
			FlightsUrl: "https://no/place/like/home",
			Id:         "8265221b-9528-4d45-900d-59a148e13850",
			Owner:      "me-myself-and-i",
		},
	}, resp.ServiceAreas)
}

func TestGetSubscription(t *testing.T) {
	for _, r := range []struct {
		name         string
		id           dssmodels.ID
		subscription *ridmodels.Subscription
		err          error
	}{
		{
			name:         "subscription-is-returned-if-returned-from-store",
			id:           dssmodels.ID(uuid.New().String()),
			subscription: &ridmodels.Subscription{},
		},
		{
			name: "error-is-returned-if-returned-from-store",
			id:   dssmodels.ID(uuid.New().String()),
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

			_, err := s.GetSubscription(context.Background(), &ridpb.GetSubscriptionRequest{
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

	_, err := s.SearchSubscriptions(ctx, &ridpb.SearchSubscriptionsRequest{
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

	_, err := s.SearchSubscriptions(ctx, &ridpb.SearchSubscriptionsRequest{
		Area: testdata.LoopWithOddNumberOfCoordinates,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestSearchSubscriptions(t *testing.T) {
	var (
		owner = dssmodels.Owner("foo")
		ctx   = auth.ContextWithOwner(context.Background(), owner)
		ms    = &mockStore{}
		s     = &Server{
			Store: ms,
		}
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ms.On("SearchSubscriptions", mock.Anything, mock.Anything, owner).Return(
		[]*ridmodels.Subscription{
			{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             owner,
				URL:               "https://no/place/like/home",
				NotificationIndex: 42,
			},
		}, error(nil),
	)
	resp, err := s.SearchSubscriptions(ctx, &ridpb.SearchSubscriptionsRequest{
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
		id         dssmodels.ID
		extents    *ridpb.Volume4D
		flightsURL string
		wantISA    *ridmodels.IdentificationServiceArea
		wantErr    error
	}{
		{
			name:       "success",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    testdata.LoopVolume4D,
			flightsURL: "https://example.com",
			wantISA: &ridmodels.IdentificationServiceArea{
				ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				URL:        "https://example.com",
				Owner:      "foo",
				Cells:      mustPolygonToCellIDs(testdata.LoopPolygon),
				StartTime:  mustTimestamp(testdata.LoopVolume4D.GetTimeStart()),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.GetTimeEnd()),
				AltitudeHi: &testdata.LoopVolume3D.AltitudeHi,
				AltitudeLo: &testdata.LoopVolume3D.AltitudeLo,
			},
		},
		{
			name:       "missing-extents",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			flightsURL: "https://example.com",
			wantErr:    dsserr.BadRequest("missing required extents"),
		},
		{
			name:       "missing-extents-spatial-volume",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    &ridpb.Volume4D{},
			flightsURL: "https://example.com",
			wantErr:    dsserr.BadRequest("bad extents: missing required spatial_volume"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &ridpb.Volume4D{
				SpatialVolume: &ridpb.Volume3D{},
			},
			flightsURL: "https://example.com",
			wantErr:    dsserr.BadRequest("bad extents: spatial_volume missing required footprint"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &ridpb.Volume4D{
				SpatialVolume: &ridpb.Volume3D{
					Footprint: &ridpb.GeoPolygon{},
				},
			},
			flightsURL: "https://example.com",
			wantErr:    dsserr.BadRequest("bad extents: not enough points in polygon"),
		},
		{
			name:    "missing-flights-url",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			wantErr: dsserr.BadRequest("missing required flightsURL"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			if r.wantISA != nil {
				store.On("InsertISA", mock.Anything, r.wantISA).Return(
					r.wantISA, []*ridmodels.Subscription(nil), nil)
			}
			s := &Server{
				Store: store,
			}

			_, err := s.CreateIdentificationServiceArea(ctx, &ridpb.CreateIdentificationServiceAreaRequest{
				Id: r.id.String(),
				Params: &ridpb.CreateIdentificationServiceAreaParameters{
					Extents:    r.extents,
					FlightsUrl: r.flightsURL,
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

	_, err := s.DeleteIdentificationServiceArea(context.Background(), &ridpb.DeleteIdentificationServiceAreaRequest{
		Id: id,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestDeleteIdentificationServiceArea(t *testing.T) {
	var (
		owner      = dssmodels.Owner("foo")
		id         = dssmodels.ID(uuid.New().String())
		version, _ = dssmodels.VersionFromString("bar")
		ctx        = auth.ContextWithOwner(context.Background(), owner)
		ms         = &mockStore{}
		s          = &Server{
			Store: ms,
		}
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ms.On("DeleteISA", mock.Anything, id, owner, mock.Anything).Return(
		&ridmodels.IdentificationServiceArea{
			ID:      dssmodels.ID(id),
			Owner:   dssmodels.Owner("me-myself-and-i"),
			URL:     "https://no/place/like/home",
			Version: version,
		},
		[]*ridmodels.Subscription{
			{
				NotificationIndex: 42,
				URL:               "https://no/place/like/home",
			},
		}, error(nil),
	)
	resp, err := s.DeleteIdentificationServiceArea(ctx, &ridpb.DeleteIdentificationServiceAreaRequest{
		Id: id.String(), Version: version.String(),
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

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ms.On("SearchISAs", mock.Anything, mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).Return(
		[]*ridmodels.IdentificationServiceArea{
			{
				ID:    dssmodels.ID(uuid.New().String()),
				Owner: dssmodels.Owner("me-myself-and-i"),
				URL:   "https://no/place/like/home",
			},
		}, error(nil),
	)
	resp, err := s.SearchIdentificationServiceAreas(ctx, &ridpb.SearchIdentificationServiceAreasRequest{
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
