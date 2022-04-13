package server

import (
	"context"
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
	"github.com/google/uuid"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var timeout = time.Second * 10

func mustTimestamp(ts *tspb.Timestamp) *time.Time {
	t := ts.AsTime()
	err := ts.CheckValid()
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

type mockApp struct {
	mock.Mock
}

func (ma *mockApp) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, s)
	return args.Get(0).(*ridmodels.Subscription), args.Error(1)
}

func (ma *mockApp) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, s)
	return args.Get(0).(*ridmodels.Subscription), args.Error(1)
}

func (ma *mockApp) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, id)
	return args.Get(0).(*ridmodels.Subscription), args.Error(1)
}

func (ma *mockApp) DeleteSubscription(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, id, owner, version)
	return args.Get(0).(*ridmodels.Subscription), args.Error(1)
}

func (ma *mockApp) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, cells, owner)
	return args.Get(0).([]*ridmodels.Subscription), args.Error(1)
}

func (ma *mockApp) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, id)
	return args.Get(0).(*ridmodels.IdentificationServiceArea), args.Error(1)
}

func (ma *mockApp) DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, id, owner, version)
	return args.Get(0).(*ridmodels.IdentificationServiceArea), args.Get(1).([]*ridmodels.Subscription), args.Error(2)
}

func (ma *mockApp) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	args := ma.Called(ctx, isa)
	return args.Get(0).(*ridmodels.IdentificationServiceArea), args.Get(1).([]*ridmodels.Subscription), args.Error(2)
}

func (ma *mockApp) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	args := ma.Called(ctx, isa)
	return args.Get(0).(*ridmodels.IdentificationServiceArea), args.Get(1).([]*ridmodels.Subscription), args.Error(2)
}

func (ma *mockApp) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ma.Called(ctx, cells, earliest, latest)
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
		wantErr      stacktrace.ErrorCode
	}{
		{
			name:         "subscription-is-returned-if-returned-from-app",
			id:           dssmodels.ID(uuid.New().String()),
			version:      version,
			subscription: &ridmodels.Subscription{},
		},
		{
			name:    "error-is-returned-if-returned-from-app",
			id:      dssmodels.ID(uuid.New().String()),
			version: version,
			wantErr: dsserr.NotFound,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}
			ma.On("DeleteSubscription", mock.Anything, r.id, mock.Anything, r.version).Return(
				r.subscription, stacktrace.NewErrorWithCode(r.wantErr, "Expected error"),
			)
			s := &Server{
				App: ma,
			}

			_, err := s.DeleteSubscription(ctx, &ridpb.DeleteSubscriptionRequest{
				Id: r.id.String(), Version: r.version.String(),
			})
			if r.wantErr != stacktrace.ErrorCode(0) {
				require.Equal(t, stacktrace.GetCode(err), r.wantErr)
			}
			require.True(t, ma.AssertExpectations(t))
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
		wantErr          stacktrace.ErrorCode
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
			wantErr: dsserr.BadRequest,
		},
		{
			name: "missing-extents-spatial-volume",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: &ridpb.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: "https://example.com",
			},
			extents: &ridpb.Volume4D{},
			wantErr: dsserr.BadRequest,
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
			wantErr: dsserr.BadRequest,
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
			wantErr: dsserr.BadRequest,
		},
		{
			name:    "missing-callbacks",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			wantErr: dsserr.BadRequest,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}
			if r.wantErr == stacktrace.ErrorCode(0) {
				ma.On("SearchISAs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*ridmodels.IdentificationServiceArea(nil), nil)
				ma.On("InsertSubscription", mock.Anything, r.wantSubscription).Return(
					r.wantSubscription, nil,
				)
			}
			s := &Server{App: ma}

			_, err := s.CreateSubscription(ctx, &ridpb.CreateSubscriptionRequest{
				Id: r.id.String(),
				Params: &ridpb.CreateSubscriptionParameters{
					Callbacks: r.callbacks,
					Extents:   r.extents,
				},
			})
			if r.wantErr != stacktrace.ErrorCode(0) {
				require.Equal(t, stacktrace.GetCode(err), r.wantErr)
			}
			require.True(t, ma.AssertExpectations(t))
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

	ma := &mockApp{}

	ma.On("SearchISAs", mock.Anything, cells, mock.Anything, mock.Anything).Return(isas, nil)
	ma.On("InsertSubscription", mock.Anything, sub).Return(sub, nil)
	s := &Server{
		App: ma,
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
	require.True(t, ma.AssertExpectations(t))

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
		err          stacktrace.ErrorCode
	}{
		{
			name:         "subscription-is-returned-if-returned-from-app",
			id:           dssmodels.ID(uuid.New().String()),
			subscription: &ridmodels.Subscription{},
		},
		{
			name: "error-is-returned-if-returned-from-app",
			id:   dssmodels.ID(uuid.New().String()),
			err:  dsserr.NotFound,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}

			ma.On("GetSubscription", mock.Anything, r.id).Return(
				r.subscription, stacktrace.NewErrorWithCode(r.err, "Expected error"),
			)
			s := &Server{
				App: ma,
			}

			_, err := s.GetSubscription(context.Background(), &ridpb.GetSubscriptionRequest{
				Id: r.id.String(),
			})
			require.Equal(t, stacktrace.GetCode(err), r.err)
			require.True(t, ma.AssertExpectations(t))
		})
	}
}

func TestSearchSubscriptionsFailsIfOwnerMissingFromContext(t *testing.T) {
	var (
		ctx = context.Background()
		ma  = &mockApp{}
		s   = &Server{
			App: ma,
		}
	)

	_, err := s.SearchSubscriptions(ctx, &ridpb.SearchSubscriptionsRequest{
		Area: testdata.Loop,
	})

	require.Error(t, err)
	require.True(t, ma.AssertExpectations(t))
}

func TestSearchSubscriptionsFailsForInvalidArea(t *testing.T) {
	var (
		ctx = auth.ContextWithOwner(context.Background(), "foo")
		ma  = &mockApp{}
		s   = &Server{
			App: ma,
		}
	)

	_, err := s.SearchSubscriptions(ctx, &ridpb.SearchSubscriptionsRequest{
		Area: testdata.LoopWithOddNumberOfCoordinates,
	})

	require.Error(t, err)
	require.True(t, ma.AssertExpectations(t))
}

func TestSearchSubscriptions(t *testing.T) {
	var (
		owner = dssmodels.Owner("foo")
		ctx   = auth.ContextWithOwner(context.Background(), owner)
		ma    = &mockApp{}
		s     = &Server{
			App: ma,
		}
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ma.On("SearchSubscriptionsByOwner", mock.Anything, mock.Anything, owner).Return(
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
	require.True(t, ma.AssertExpectations(t))
}

func TestCreateISA(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")

	for _, r := range []struct {
		name       string
		id         dssmodels.ID
		extents    *ridpb.Volume4D
		flightsURL string
		wantISA    *ridmodels.IdentificationServiceArea
		wantErr    stacktrace.ErrorCode
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
			wantErr:    dsserr.BadRequest,
		},
		{
			name:       "missing-extents-spatial-volume",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    &ridpb.Volume4D{},
			flightsURL: "https://example.com",
			wantErr:    dsserr.BadRequest,
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &ridpb.Volume4D{
				SpatialVolume: &ridpb.Volume3D{},
			},
			flightsURL: "https://example.com",
			wantErr:    dsserr.BadRequest,
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
			wantErr:    dsserr.BadRequest,
		},
		{
			name:    "missing-flights-url",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			wantErr: dsserr.BadRequest,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}
			if r.wantISA != nil {
				ma.On("InsertISA", mock.Anything, r.wantISA).Return(
					r.wantISA, []*ridmodels.Subscription(nil), nil)
			}
			s := &Server{
				App: ma,
			}

			_, err := s.CreateIdentificationServiceArea(ctx, &ridpb.CreateIdentificationServiceAreaRequest{
				Id: r.id.String(),
				Params: &ridpb.CreateIdentificationServiceAreaParameters{
					Extents:    r.extents,
					FlightsUrl: r.flightsURL,
				},
			})
			if r.wantErr != stacktrace.ErrorCode(0) {
				require.Equal(t, stacktrace.GetCode(err), r.wantErr)
			}
			require.True(t, ma.AssertExpectations(t))
		})
	}
}

func TestUpdateISA(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")
	version, _ := dssmodels.VersionFromString("bar")
	for _, r := range []struct {
		name       string
		id         dssmodels.ID
		extents    *ridpb.Volume4D
		flightsURL string
		wantISA    *ridmodels.IdentificationServiceArea
		wantErr    stacktrace.ErrorCode
		version    *dssmodels.Version
	}{
		{
			name:       "success",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    testdata.LoopVolume4D,
			flightsURL: "https://example.com",
			version:    version,
			wantISA: &ridmodels.IdentificationServiceArea{
				ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				URL:        "https://example.com",
				Owner:      "foo",
				Cells:      mustPolygonToCellIDs(testdata.LoopPolygon),
				StartTime:  mustTimestamp(testdata.LoopVolume4D.GetTimeStart()),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.GetTimeEnd()),
				AltitudeHi: &testdata.LoopVolume3D.AltitudeHi,
				AltitudeLo: &testdata.LoopVolume3D.AltitudeLo,
				Writer:     "locality value",
				Version:    version,
			},
		},
		{
			name:    "missing-flights-url",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			version: version,
			wantErr: dsserr.BadRequest,
		},
		{
			name:       "missing-extents",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			flightsURL: "https://example.com",
			version:    version,
			wantErr:    dsserr.BadRequest,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}
			if r.wantISA != nil {
				ma.On("UpdateISA", mock.Anything, r.wantISA).Return(
					r.wantISA, []*ridmodels.Subscription(nil), nil)
			}
			s := &Server{
				App:      ma,
				Locality: "locality value",
			}
			_, err := s.UpdateIdentificationServiceArea(ctx, &ridpb.UpdateIdentificationServiceAreaRequest{
				Id:      r.id.String(),
				Version: r.version.String(),
				Params: &ridpb.UpdateIdentificationServiceAreaParameters{
					Extents:    r.extents,
					FlightsUrl: r.flightsURL,
				},
			})
			if r.wantErr != stacktrace.ErrorCode(0) {
				require.Equal(t, stacktrace.GetCode(err), r.wantErr)
			}
			require.True(t, ma.AssertExpectations(t))
		})
	}
}

func TestDeleteIdentificationServiceAreaRequiresOwnerInContext(t *testing.T) {
	var (
		id = uuid.New().String()
		ma = &mockApp{}

		s = &Server{
			App: ma,
		}
	)

	_, err := s.DeleteIdentificationServiceArea(context.Background(), &ridpb.DeleteIdentificationServiceAreaRequest{
		Id: id,
	})

	require.Error(t, err)
	require.True(t, ma.AssertExpectations(t))
}

func TestDeleteIdentificationServiceArea(t *testing.T) {
	var (
		owner      = dssmodels.Owner("foo")
		id         = dssmodels.ID(uuid.New().String())
		version, _ = dssmodels.VersionFromString("bar")
		ctx        = auth.ContextWithOwner(context.Background(), owner)
		ma         = &mockApp{}

		s = &Server{
			App: ma,
		}
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ma.On("DeleteISA", mock.Anything, id, owner, mock.Anything).Return(
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
	require.True(t, ma.AssertExpectations(t))
}

func TestSearchIdentificationServiceAreas(t *testing.T) {
	var (
		ctx = context.Background()
		ma  = &mockApp{}

		s = &Server{
			App: ma,
		}
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ma.On("SearchISAs", mock.Anything, mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).Return(
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
	require.True(t, ma.AssertExpectations(t))
}

func TestDefaultRegionCovererProducesResults(t *testing.T) {
	cover, err := geo.AreaToCellIDs(testdata.Loop)
	require.NoError(t, err)
	require.NotNil(t, cover)
}
