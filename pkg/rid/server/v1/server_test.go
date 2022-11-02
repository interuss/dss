package v1

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/ridv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/dss/pkg/geo/testdata"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	apiv1 "github.com/interuss/dss/pkg/rid/models/api/v1"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var timeout = time.Second * 10

func mustTimestamp(ts *string) *time.Time {
	t, err := time.Parse(time.RFC3339Nano, *ts)
	if err != nil {
		panic(err)
	}
	return &t
}

func mustPolygonToCellIDs(p *restapi.GeoPolygon) s2.CellUnion {
	cells, err := apiv1.FromGeoPolygon(p).CalculateCovering()
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
	var respSet restapi.DeleteSubscriptionResponseSet
	for _, r := range []struct {
		name         string
		id           dssmodels.ID
		version      *dssmodels.Version
		subscription *ridmodels.Subscription
		appErr       stacktrace.ErrorCode
		wantErr      **restapi.ErrorResponse
	}{
		{
			name:         "subscription-is-returned-if-returned-from-app",
			id:           dssmodels.ID(uuid.New().String()),
			version:      testdata.Version,
			subscription: &ridmodels.Subscription{},
		},
		{
			name:    "error-is-returned-if-returned-from-app",
			id:      dssmodels.ID(uuid.New().String()),
			version: testdata.Version,
			appErr:  dsserr.NotFound,
			wantErr: &respSet.Response404,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}
			if r.appErr == stacktrace.ErrorCode(0) {
				ma.On("DeleteSubscription", mock.Anything, r.id, mock.Anything, r.version).Return(
					r.subscription, nil,
				)
			} else {
				ma.On("DeleteSubscription", mock.Anything, r.id, mock.Anything, r.version).Return(
					(*ridmodels.Subscription)(nil), stacktrace.NewErrorWithCode(r.appErr, "Expected error"),
				)
			}

			s := &Server{
				App: ma,
			}

			respSet = s.DeleteSubscription(context.Background(), &restapi.DeleteSubscriptionRequest{
				Id: restapi.SubscriptionUUID(r.id.String()), Version: r.version.String(), Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
			})
			if r.wantErr != nil {
				require.NotNil(t, *r.wantErr)
			} else {
				require.NotNil(t, respSet.Response200)
			}
			require.True(t, ma.AssertExpectations(t))
		})
	}
}

func TestCreateSubscription(t *testing.T) {
	var respSet restapi.CreateSubscriptionResponseSet
	for _, r := range []struct {
		name             string
		id               dssmodels.ID
		callbacks        restapi.SubscriptionCallbacks
		extents          restapi.Volume4D
		wantSubscription *ridmodels.Subscription
		appErr           stacktrace.ErrorCode
		wantErr          **restapi.ErrorResponse
	}{
		{
			name:      "success",
			id:        dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: restapi.SubscriptionCallbacks{IdentificationServiceAreaUrl: &testdata.CallbackURL},
			extents:   testdata.LoopVolume4D,
			wantSubscription: &ridmodels.Subscription{
				ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				Owner:      "foo",
				URL:        "https://example.com",
				StartTime:  mustTimestamp(testdata.LoopVolume4D.TimeStart),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.TimeEnd),
				AltitudeHi: (*float32)(testdata.LoopVolume3D.AltitudeHi),
				AltitudeLo: (*float32)(testdata.LoopVolume3D.AltitudeLo),
				Cells:      mustPolygonToCellIDs(&testdata.LoopPolygon),
			},
		},
		{
			name:      "missing-extents",
			id:        dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: restapi.SubscriptionCallbacks{IdentificationServiceAreaUrl: &testdata.CallbackURL},
			appErr:    dsserr.BadRequest,
			wantErr:   &respSet.Response400,
		},
		{
			name:      "missing-extents-spatial-volume",
			id:        dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: restapi.SubscriptionCallbacks{IdentificationServiceAreaUrl: &testdata.CallbackURL},
			extents:   restapi.Volume4D{},
			appErr:    dsserr.BadRequest,
			wantErr:   &respSet.Response400,
		},
		{
			name:      "missing-spatial-volume-footprint",
			id:        dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: restapi.SubscriptionCallbacks{IdentificationServiceAreaUrl: &testdata.CallbackURL},
			extents: restapi.Volume4D{
				SpatialVolume: restapi.Volume3D{},
			},
			appErr:  dsserr.BadRequest,
			wantErr: &respSet.Response400,
		},
		{
			name:      "missing-spatial-volume-footprint",
			id:        dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			callbacks: restapi.SubscriptionCallbacks{IdentificationServiceAreaUrl: &testdata.CallbackURL},
			extents: restapi.Volume4D{
				SpatialVolume: restapi.Volume3D{
					Footprint: restapi.GeoPolygon{},
				},
			},
			appErr:  dsserr.BadRequest,
			wantErr: &respSet.Response400,
		},
		{
			name:    "missing-callbacks",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			appErr:  dsserr.BadRequest,
			wantErr: &respSet.Response400,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}
			if r.appErr == stacktrace.ErrorCode(0) {
				ma.On("SearchISAs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*ridmodels.IdentificationServiceArea(nil), nil)
				ma.On("InsertSubscription", mock.Anything, r.wantSubscription).Return(
					r.wantSubscription, nil,
				)
			}
			s := &Server{App: ma}

			respSet = s.CreateSubscription(context.Background(), &restapi.CreateSubscriptionRequest{
				Id: restapi.SubscriptionUUID(r.id.String()),
				Body: &restapi.CreateSubscriptionParameters{
					Callbacks: r.callbacks,
					Extents:   r.extents,
				},
				Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
			})
			if r.wantErr != nil {
				require.NotNil(t, *r.wantErr)
			} else {
				require.NotNil(t, respSet.Response200)
			}
			require.True(t, ma.AssertExpectations(t))
		})
	}
}

func TestCreateSubscriptionResponseIncludesISAs(t *testing.T) {
	isas := []*ridmodels.IdentificationServiceArea{
		{
			ID:    dssmodels.ID("8265221b-9528-4d45-900d-59a148e13850"),
			Owner: dssmodels.Owner("me-myself-and-i"),
			URL:   "https://no/place/like/home",
		},
	}

	cells := mustPolygonToCellIDs(&testdata.LoopPolygon)
	sub := &ridmodels.Subscription{
		ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
		Owner:      "foo",
		URL:        string(testdata.CallbackURL),
		StartTime:  mustTimestamp(testdata.LoopVolume4D.TimeStart),
		EndTime:    mustTimestamp(testdata.LoopVolume4D.TimeEnd),
		AltitudeHi: (*float32)(testdata.LoopVolume3D.AltitudeHi),
		AltitudeLo: (*float32)(testdata.LoopVolume3D.AltitudeLo),
		Cells:      cells,
	}

	ma := &mockApp{}

	ma.On("SearchISAs", mock.Anything, cells, mock.Anything, mock.Anything).Return(isas, nil)
	ma.On("InsertSubscription", mock.Anything, sub).Return(sub, nil)
	s := &Server{
		App: ma,
	}

	respSet := s.CreateSubscription(context.Background(), &restapi.CreateSubscriptionRequest{
		Id: restapi.SubscriptionUUID(sub.ID.String()),
		Body: &restapi.CreateSubscriptionParameters{
			Callbacks: restapi.SubscriptionCallbacks{
				IdentificationServiceAreaUrl: &testdata.CallbackURL,
			},
			Extents: testdata.LoopVolume4D,
		},
		Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
	})
	require.NotNil(t, respSet.Response200)
	require.True(t, ma.AssertExpectations(t))

	require.Equal(t, []restapi.IdentificationServiceArea{
		{
			FlightsUrl: "https://no/place/like/home",
			Id:         "8265221b-9528-4d45-900d-59a148e13850",
			Owner:      "me-myself-and-i",
		},
	}, *respSet.Response200.ServiceAreas)
}

func TestGetSubscription(t *testing.T) {
	var respSet restapi.GetSubscriptionResponseSet
	for _, r := range []struct {
		name         string
		id           dssmodels.ID
		subscription *ridmodels.Subscription
		wantErr      **restapi.ErrorResponse
	}{
		{
			name:         "subscription-is-returned-if-returned-from-app",
			id:           dssmodels.ID(uuid.New().String()),
			subscription: &ridmodels.Subscription{},
		},
		{
			name:    "error-is-returned-if-returned-from-app",
			id:      dssmodels.ID(uuid.New().String()),
			wantErr: &respSet.Response404,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ma := &mockApp{}
			ma.On("GetSubscription", mock.Anything, r.id).Return(
				r.subscription, nil,
			)
			s := &Server{
				App: ma,
			}

			respSet = s.GetSubscription(context.Background(), &restapi.GetSubscriptionRequest{
				Id: restapi.SubscriptionUUID(r.id.String()),
			})
			if r.wantErr != nil {
				require.NotNil(t, *r.wantErr)
			} else {
				require.NotNil(t, respSet.Response200)
			}
			require.True(t, ma.AssertExpectations(t))
		})
	}
}

func TestSearchSubscriptionsFailsIfOwnerMissingFromContext(t *testing.T) {
	var (
		ma = &mockApp{}
		s  = &Server{
			App: ma,
		}
	)

	respSet := s.SearchSubscriptions(context.Background(), &restapi.SearchSubscriptionsRequest{
		Area: (*restapi.GeoPolygonString)(&testdata.Loop),
	})

	require.NotNil(t, respSet.Response403)
	require.True(t, ma.AssertExpectations(t))
}

func TestSearchSubscriptionsFailsForInvalidArea(t *testing.T) {
	var (
		ma = &mockApp{}
		s  = &Server{
			App: ma,
		}
	)

	respSet := s.SearchSubscriptions(context.Background(), &restapi.SearchSubscriptionsRequest{
		Area: (*restapi.GeoPolygonString)(&testdata.LoopWithOddNumberOfCoordinates),
		Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
	})

	require.NotNil(t, respSet.Response400)
	require.True(t, ma.AssertExpectations(t))
}

func TestSearchSubscriptions(t *testing.T) {
	var (
		ma = &mockApp{}
		s  = &Server{
			App: ma,
		}
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ma.On("SearchSubscriptionsByOwner", mock.Anything, mock.Anything, dssmodels.Owner(testdata.Owner)).Return(
		[]*ridmodels.Subscription{
			{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             dssmodels.Owner(testdata.Owner),
				URL:               "https://no/place/like/home",
				NotificationIndex: 42,
			},
		}, error(nil),
	)
	respSet := s.SearchSubscriptions(ctx, &restapi.SearchSubscriptionsRequest{
		Area: (*restapi.GeoPolygonString)(&testdata.Loop),
		Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
	})

	require.NotNil(t, respSet.Response200)
	require.Len(t, respSet.Response200.Subscriptions, 1)
	require.True(t, ma.AssertExpectations(t))
}

func TestCreateISA(t *testing.T) {
	var respSet restapi.CreateIdentificationServiceAreaResponseSet
	for _, r := range []struct {
		name       string
		id         dssmodels.ID
		extents    restapi.Volume4D
		flightsURL string
		wantISA    *ridmodels.IdentificationServiceArea
		appErr     stacktrace.ErrorCode
		wantErr    **restapi.ErrorResponse
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
				Cells:      mustPolygonToCellIDs(&testdata.LoopPolygon),
				StartTime:  mustTimestamp(testdata.LoopVolume4D.TimeStart),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.TimeEnd),
				AltitudeHi: (*float32)(testdata.LoopVolume3D.AltitudeHi),
				AltitudeLo: (*float32)(testdata.LoopVolume3D.AltitudeLo),
			},
		},
		{
			name:       "missing-extents",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			flightsURL: "https://example.com",
			appErr:     dsserr.BadRequest,
			wantErr:    &respSet.Response400,
		},
		{
			name:       "missing-extents-spatial-volume",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    restapi.Volume4D{},
			flightsURL: "https://example.com",
			appErr:     dsserr.BadRequest,
			wantErr:    &respSet.Response400,
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: restapi.Volume4D{
				SpatialVolume: restapi.Volume3D{},
			},
			flightsURL: "https://example.com",
			appErr:     dsserr.BadRequest,
			wantErr:    &respSet.Response400,
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: restapi.Volume4D{
				SpatialVolume: restapi.Volume3D{
					Footprint: restapi.GeoPolygon{},
				},
			},
			flightsURL: "https://example.com",
			appErr:     dsserr.BadRequest,
			wantErr:    &respSet.Response400,
		},
		{
			name:    "missing-flights-url",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			appErr:  dsserr.BadRequest,
			wantErr: &respSet.Response400,
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

			respSet = s.CreateIdentificationServiceArea(context.Background(), &restapi.CreateIdentificationServiceAreaRequest{
				Id: restapi.EntityUUID(r.id.String()),
				Body: &restapi.CreateIdentificationServiceAreaParameters{
					Extents:    r.extents,
					FlightsUrl: restapi.RIDFlightsURL(r.flightsURL),
				},
				Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
			})
			if r.wantErr != nil {
				require.NotNil(t, *r.wantErr)
			} else {
				require.NotNil(t, respSet.Response200)
			}
			require.True(t, ma.AssertExpectations(t))
		})
	}
}

func TestUpdateISA(t *testing.T) {
	var respSet restapi.UpdateIdentificationServiceAreaResponseSet
	for _, r := range []struct {
		name       string
		id         dssmodels.ID
		extents    restapi.Volume4D
		flightsURL string
		wantISA    *ridmodels.IdentificationServiceArea
		appErr     stacktrace.ErrorCode
		wantErr    **restapi.ErrorResponse
		version    *dssmodels.Version
	}{
		{
			name:       "success",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents:    testdata.LoopVolume4D,
			flightsURL: "https://example.com",
			version:    testdata.Version,
			wantISA: &ridmodels.IdentificationServiceArea{
				ID:         "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				URL:        "https://example.com",
				Owner:      "foo",
				Cells:      mustPolygonToCellIDs(&testdata.LoopPolygon),
				StartTime:  mustTimestamp(testdata.LoopVolume4D.TimeStart),
				EndTime:    mustTimestamp(testdata.LoopVolume4D.TimeEnd),
				AltitudeHi: (*float32)(testdata.LoopVolume3D.AltitudeHi),
				AltitudeLo: (*float32)(testdata.LoopVolume3D.AltitudeLo),
				Writer:     "locality value",
				Version:    testdata.Version,
			},
		},
		{
			name:    "missing-flights-url",
			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			version: testdata.Version,
			appErr:  dsserr.BadRequest,
			wantErr: &respSet.Response400,
		},
		{
			name:       "missing-extents",
			id:         dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			flightsURL: "https://example.com",
			version:    testdata.Version,
			appErr:     dsserr.BadRequest,
			wantErr:    &respSet.Response400,
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
			respSet = s.UpdateIdentificationServiceArea(context.Background(), &restapi.UpdateIdentificationServiceAreaRequest{
				Id:      restapi.EntityUUID(r.id.String()),
				Version: r.version.String(),
				Body: &restapi.UpdateIdentificationServiceAreaParameters{
					Extents:    r.extents,
					FlightsUrl: restapi.RIDFlightsURL(r.flightsURL),
				},
				Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
			})
			if r.wantErr != nil {
				require.NotNil(t, *r.wantErr)
			} else {
				require.NotNil(t, respSet.Response200)
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

	respSet := s.DeleteIdentificationServiceArea(context.Background(), &restapi.DeleteIdentificationServiceAreaRequest{
		Id: restapi.EntityUUID(id),
	})

	require.NotNil(t, respSet.Response403)
	require.True(t, ma.AssertExpectations(t))
}

func TestDeleteIdentificationServiceArea(t *testing.T) {
	var (
		id = dssmodels.ID(uuid.New().String())
		ma = &mockApp{}

		s = &Server{
			App: ma,
		}
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ma.On("DeleteISA", mock.Anything, id, dssmodels.Owner(testdata.Owner), mock.Anything).Return(
		&ridmodels.IdentificationServiceArea{
			ID:      dssmodels.ID(id),
			Owner:   dssmodels.Owner("me-myself-and-i"),
			URL:     "https://no/place/like/home",
			Version: testdata.Version,
		},
		[]*ridmodels.Subscription{
			{
				NotificationIndex: 42,
				URL:               "https://no/place/like/home",
			},
		}, error(nil),
	)
	respSet := s.DeleteIdentificationServiceArea(ctx, &restapi.DeleteIdentificationServiceAreaRequest{
		Id: restapi.EntityUUID(id.String()), Version: testdata.Version.String(),
		Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
	})

	require.NotNil(t, respSet.Response200)
	require.Len(t, respSet.Response200.Subscribers, 1)
	require.True(t, ma.AssertExpectations(t))
}

func TestSearchIdentificationServiceAreas(t *testing.T) {
	var (
		ma = &mockApp{}

		s = &Server{
			App: ma,
		}
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
	respSet := s.SearchIdentificationServiceAreas(ctx, &restapi.SearchIdentificationServiceAreasRequest{
		Area: (*restapi.GeoPolygonString)(&testdata.Loop),
		Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
	})

	require.NotNil(t, respSet.Response200)
	require.Len(t, respSet.Response200.ServiceAreas, 1)
	require.True(t, ma.AssertExpectations(t))
}

func TestDefaultRegionCovererProducesResults(t *testing.T) {
	cover, err := geo.AreaToCellIDs(testdata.Loop)
	require.NoError(t, err)
	require.NotNil(t, cover)
}
