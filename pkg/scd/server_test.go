package scd

import (
	"context"
	"errors"

	testdata "github.com/interuss/dss/pkg/geo/testdata/scd"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	"testing"
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"

	"github.com/interuss/dss/pkg/auth"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
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

func mustPolygonToCellIDs(p *scdpb.Polygon) s2.CellUnion {
	cells, err := dssmodels.GeoPolygonFromSCDProto(p).CalculateCovering()
	if err != nil {
		panic(err)
	}
	return cells
}

func float32p(v float32) *float32 {
	return &v
}

type mockStore struct {
	mock.Mock
}

func (ms *mockStore) Close() error {
	return ms.Called().Error(0)
}

func (ms *mockStore) DeleteSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner, version scdmodels.Version) (*scdmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id, owner, version)
	return args.Get(0).(*scdmodels.Subscription), args.Error(1)
}

func (ms *mockStore) GetSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id)
	return args.Get(0).(*scdmodels.Subscription), args.Error(1)
}

func (ms *mockStore) UpsertSubscription(ctx context.Context, subscription *scdmodels.Subscription) (*scdmodels.Subscription, []*scdmodels.Operation, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, subscription)
	return args.Get(0).(*scdmodels.Subscription), args.Get(1).([]*scdmodels.Operation), args.Error(2)
}

func (ms *mockStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*scdmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, cells, owner)
	return args.Get(0).([]*scdmodels.Subscription), args.Error(1)
}

func (ms *mockStore) GetOperation(ctx context.Context, id scdmodels.ID) (*scdmodels.Operation, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id)
	return args.Get(0).(*scdmodels.Operation), args.Error(1)
}

func (ms *mockStore) DeleteOperation(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, id, owner)
	return args.Get(0).(*scdmodels.Operation), args.Get(1).([]*scdmodels.Subscription), args.Error(2)
}

func (ms *mockStore) UpsertOperation(ctx context.Context, operation *scdmodels.Operation, key []scdmodels.OVN) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, operation, key)
	return args.Get(0).(*scdmodels.Operation), args.Get(1).([]*scdmodels.Subscription), args.Error(2)
}

func (ms *mockStore) SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := ms.Called(ctx, v4d, owner)
	return args.Get(0).([]*scdmodels.Operation), args.Error(1)
}

var validSubscription = &scdmodels.Subscription{
	ID:                   "4348c8e5-0b1c-43cf-9114-2e67a4532765",
	Owner:                "foo",
	Version:              0,
	BaseURL:              "https://example.com",
	NotifyForOperations:  true,
	NotifyForConstraints: true,
}

func TestDeleteSubscription(t *testing.T) {
	for _, r := range []struct {
		name             string
		id               scdmodels.ID
		owner            dssmodels.Owner
		version          scdmodels.Version
		wantSubscription *scdmodels.Subscription
		wantErr          error
	}{
		{
			name:    "missing owner",
			id:      scdmodels.ID(uuid.New().String()),
			version: scdmodels.Version(0),
			wantErr: dsserr.PermissionDenied("missing owner from context"),
		},
		{
			name:             "subscription-is-returned-if-returned-from-store",
			id:               scdmodels.ID(uuid.New().String()),
			owner:            dssmodels.Owner("foo"),
			version:          scdmodels.Version(0),
			wantSubscription: validSubscription,
		},
		{
			name:    "error-is-returned-if-returned-from-store",
			id:      scdmodels.ID(uuid.New().String()),
			owner:   dssmodels.Owner("foo"),
			version: scdmodels.Version(0),
			wantErr: errors.New("failed to look up subscription for ID"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ctx := context.Background()
			if r.owner != "" {
				ctx = auth.ContextWithOwner(ctx, r.owner)
			}
			store := &mockStore{}
			store.On("DeleteSubscription", mock.Anything, r.id, mock.Anything, r.version).Return(
				r.wantSubscription, r.wantErr,
			)
			s := &Server{
				Store: store,
			}

			resp, err := s.DeleteSubscription(ctx, &scdpb.DeleteSubscriptionRequest{
				Subscriptionid: r.id.String(),
			})
			require.Equal(t, r.wantErr, err)
			if r.wantErr == nil {
				require.NotNil(t, resp.Subscription)
				require.True(t, store.AssertExpectations(t))
			}
		})
	}
}

func TestGetSubscription(t *testing.T) {
	for _, r := range []struct {
		name             string
		id               scdmodels.ID
		owner            dssmodels.Owner
		wantSubscription *scdmodels.Subscription
		wantErr          error
	}{
		{
			name:    "missing-owner",
			id:      scdmodels.ID(uuid.New().String()),
			wantErr: dsserr.PermissionDenied("missing owner from context"),
		},
		{
			name:    "missing-id",
			owner:   dssmodels.Owner("foo"),
			wantErr: dsserr.BadRequest("missing Subscription ID"),
		},
		{
			name:             "subscription-is-returned-if-returned-from-store",
			id:               scdmodels.ID(uuid.New().String()),
			owner:            dssmodels.Owner("foo"),
			wantSubscription: validSubscription,
		},
		{
			name:    "error-is-returned-if-returned-from-store",
			id:      scdmodels.ID(uuid.New().String()),
			owner:   dssmodels.Owner("foo"),
			wantErr: errors.New("failed to look up subscription for ID"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ctx := context.Background()
			if r.owner != "" {
				ctx = auth.ContextWithOwner(ctx, r.owner)
			}
			store := &mockStore{}
			store.On("GetSubscription", mock.Anything, r.id).Return(
				r.wantSubscription, r.wantErr,
			)
			s := &Server{
				Store: store,
			}

			resp, err := s.GetSubscription(ctx, &scdpb.GetSubscriptionRequest{
				Subscriptionid: r.id.String(),
			})
			require.Equal(t, r.wantErr, err)
			if r.wantErr == nil {
				require.NotNil(t, resp.Subscription)
				require.True(t, store.AssertExpectations(t))
			}
		})
	}
}

func TestQuerySubscriptions(t *testing.T) {
	validAoi := testdata.LoopVolume4D
	//invalidAoiWithOnlyTwoPoints, _ := scdpb.FromVolume4D(testdata.LoopVolume4DWithOnlyTwoPoints)
	for _, r := range []struct {
		name              string
		owner             dssmodels.Owner
		aoi               *scdpb.Volume4D
		wantSubscriptions []*scdmodels.Subscription
		wantErr           error
	}{
		{
			name:    "missing-owner",
			aoi:     validAoi,
			wantErr: dsserr.PermissionDenied("missing owner from context"),
		},
		{
			name:  "success",
			owner: dssmodels.Owner("foo"),
			aoi:   validAoi,
			wantSubscriptions: []*scdmodels.Subscription{
				{
					ID:                   "4348c8e5-0b1c-43cf-9114-2e67a4532765",
					Owner:                "foo",
					Version:              0,
					BaseURL:              "https://example.com",
					NotifyForOperations:  true,
					NotifyForConstraints: true,
				},
			},
		},
		{
			name:    "missing-aoi",
			owner:   dssmodels.Owner("foo"),
			wantErr: dsserr.BadRequest("missing area_of_interest"),
		},
		/* {
		     name: "missing-spatial-polygon",
		     owner: dssmodels.Owner("foo"),
		     aoi: &scdpb.Volume4D{
		       Volume: &scdpb.Volume3D{
		         OutlinePolygon: &scdpb.Polygon{},
		       },
		     },
		     wantErr: dsserr.BadRequest("bad extents: not enough points in polygon"),
		   },
		   {
		     name: "missing-spatial-polygon",
		     owner: dssmodels.Owner("foo"),
		     aoi: invalidAoiWithOnlyTwoPoints,
		     wantErr: dsserr.BadRequest("bad extents: not enough points in polygon"),
		   },*/
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			ctx := context.Background()
			if r.owner != "" {
				ctx = auth.ContextWithOwner(ctx, r.owner)
			}
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			if r.wantErr == nil {
				store.On("SearchSubscriptions", mock.Anything, mock.Anything, r.owner).Return(
					r.wantSubscriptions, nil,
				)
			}
			s := &Server{
				Store: store,
			}

			resp, err := s.QuerySubscriptions(ctx, &scdpb.QuerySubscriptionsRequest{
				Params: &scdpb.SearchSubscriptionParameters{
					AreaOfInterest: r.aoi,
				},
			})
			require.Equal(t, r.wantErr, err)
			if r.wantErr == nil {
				require.Len(t, resp.Subscriptions, 1)
				require.True(t, store.AssertExpectations(t))
			}
		})
	}
}

func TestPutSubscription(t *testing.T) {
	extentsLoop := testdata.LoopVolume4D
	for _, r := range []struct {
		name             string
		id               dssmodels.ID
		owner            dssmodels.Owner
		oldVersion       int32
		ussBaseURL       string
		extents          *scdpb.Volume4D
		wantSubscription *scdmodels.Subscription
		wantErr          error
	}{
		{
			name:             "missing-owner",
			id:               dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			oldVersion:       0,
			ussBaseURL:       "https://example.com",
			extents:          extentsLoop,
			wantSubscription: validSubscription,
			wantErr:          dsserr.PermissionDenied("missing owner from context"),
		},
		{
			name:             "missing-id",
			owner:            dssmodels.Owner("foo"),
			oldVersion:       0,
			ussBaseURL:       "https://example.com",
			extents:          extentsLoop,
			wantSubscription: validSubscription,
			wantErr:          dsserr.BadRequest("missing Subscription ID"),
		},
		{
			name:             "success",
			id:               dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			owner:            dssmodels.Owner("foo"),
			oldVersion:       0,
			ussBaseURL:       "https://example.com",
			extents:          extentsLoop,
			wantSubscription: validSubscription,
		},
		/*{
		  			name: "missing-extents",
		  			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
		        owner: dssmodels.Owner("foo"),
		  			ussBaseURL: "https://example.com",
		  			wantErr: dsserr.BadRequest("missing required extents"),
		  		},
		      {
		  			name: "missing-extents-spatial-volume",
		  			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
		        owner: dssmodels.Owner("foo"),
		        ussBaseURL: "https://example.com",
		  			extents: &scdpb.Volume4D{},
		  			wantErr: dsserr.BadRequest("bad extents: missing required spatial_volume"),
		  		},
		  		{
		  			name: "missing-spatial-volume-footprint",
		  			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
		        owner: dssmodels.Owner("foo"),
		        ussBaseURL: "https://example.com",
		  			extents: &scdpb.Volume4D{
		  				Volume: &scdpb.Volume3D{},
		  			},
		  			wantErr: dsserr.BadRequest("bad extents: spatial_volume missing required footprint"),
		  		},
		  		{
		  			name: "missing-spatial-polygon",
		  			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
		        owner: dssmodels.Owner("foo"),
		        ussBaseURL: "https://example.com",
		  			extents: &scdpb.Volume4D{
		  				Volume: &scdpb.Volume3D{
		  					OutlinePolygon: &scdpb.Polygon{},
		  				},
		  			},
		  			wantErr: dsserr.BadRequest("bad extents: not enough points in polygon"),
		  		},
		  		{
		  			name:    "missing-baseurl",
		  			id:      dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
		        owner: dssmodels.Owner("foo"),
		        extents: extentsLoop,
		  			wantErr: dsserr.BadRequest("missing required callbacks"),
		  		},*/
	} {
		t.Run(r.name, func(t *testing.T) {
			sub := *r.wantSubscription

			if r.extents != nil {
				v4d, err := dssmodels.Volume4DFromSCDProto(r.extents)
				require.NoError(t, err)

				cells, err := v4d.CalculateSpatialCovering()
				require.NoError(t, err)

				sub.StartTime = v4d.StartTime
				sub.EndTime = v4d.EndTime
				sub.AltitudeHi = v4d.SpatialVolume.AltitudeHi
				sub.AltitudeLo = v4d.SpatialVolume.AltitudeLo
				sub.Cells = cells
			}

			ctx := context.Background()
			if r.owner != "" {
				ctx = auth.ContextWithOwner(ctx, r.owner)
			}
			store := &mockStore{}
			if r.wantErr == nil {
				store.On("UpsertSubscription", mock.Anything, &sub).Return(
					r.wantSubscription, []*scdmodels.Operation(nil), nil,
				)
			}
			s := &Server{
				Store: store,
			}

			resp, err := s.PutSubscription(ctx, &scdpb.PutSubscriptionRequest{
				Subscriptionid: r.id.String(),
				Params: &scdpb.PutSubscriptionParameters{
					UssBaseUrl:           r.ussBaseURL,
					Extents:              r.extents,
					NotifyForOperations:  true,
					NotifyForConstraints: true,
				},
			})
			require.Equal(t, r.wantErr, err)
			if r.wantErr == nil {
				require.NotNil(t, resp.Subscription)
				require.True(t, store.AssertExpectations(t))
			}
		})
	}
}

func TestCreateOperation(t *testing.T) {
	ctx := auth.ContextWithOwner(context.Background(), "foo")

	for _, r := range []struct {
		name          string
		id            scdmodels.ID
		extents       *scdpb.Volume4D
		url           string
		wantOperation *scdmodels.Operation
		wantErr       error
	}{
		{
			name:    "success",
			id:      scdmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			url:     "https://example.com",
			wantOperation: &scdmodels.Operation{
				ID:            "4348c8e5-0b1c-43cf-9114-2e67a4532765",
				USSBaseURL:    "https://example.com",
				Owner:         "foo",
				Cells:         mustPolygonToCellIDs(testdata.LoopPolygon),
				StartTime:     mustTimestamp(testdata.LoopVolume4D.GetTimeStart().GetValue()),
				EndTime:       mustTimestamp(testdata.LoopVolume4D.GetTimeEnd().GetValue()),
				AltitudeUpper: float32p(float32(testdata.LoopVolume3D.AltitudeUpper.Value)),
				AltitudeLower: float32p(float32(testdata.LoopVolume3D.AltitudeLower.Value)),
			},
		},
		{
			name: "missing-extents-spatial-volume",
			id:   scdmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &scdpb.Volume4D{
				TimeStart: testdata.LoopVolume4D.GetTimeStart(),
				TimeEnd:   testdata.LoopVolume4D.GetTimeEnd(),
			},
			url:     "https://example.com",
			wantErr: dsserr.BadRequest("bad area: missing footprint"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   scdmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &scdpb.Volume4D{
				Volume:    &scdpb.Volume3D{},
				TimeStart: testdata.LoopVolume4D.GetTimeStart(),
				TimeEnd:   testdata.LoopVolume4D.GetTimeEnd(),
			},
			url:     "https://example.com",
			wantErr: dsserr.BadRequest("bad area: missing footprint"),
		},
		{
			name: "missing-spatial-volume-footprint",
			id:   scdmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: &scdpb.Volume4D{
				Volume: &scdpb.Volume3D{
					OutlinePolygon: &scdpb.Polygon{},
				},
				TimeStart: testdata.LoopVolume4D.GetTimeStart(),
				TimeEnd:   testdata.LoopVolume4D.GetTimeEnd(),
			},
			url:     "https://example.com",
			wantErr: dsserr.BadRequest("failed to union extents: not enough points in polygon"),
		},
		{
			name:    "missing-flights-url",
			id:      scdmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
			extents: testdata.LoopVolume4D,
			wantErr: dsserr.BadRequest("missing required UssBaseUrl"),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			store := &mockStore{}
			store.On("UpsertSubscription", mock.Anything, mock.Anything).Return(
				&scdmodels.Subscription{
					ID: scdmodels.ID(uuid.New().String()),
				}, []*scdmodels.Operation(nil), error(nil),
			).Maybe()

			if r.wantOperation != nil {
				store.On("UpsertOperation", mock.Anything, mock.Anything, mock.Anything).Return(
					r.wantOperation, []*scdmodels.Subscription(nil), nil)
			}
			s := &Server{
				Store: store,
			}

			_, err := s.PutOperationReference(ctx, &scdpb.PutOperationReferenceRequest{
				Entityuuid: r.id.String(),
				Params: &scdpb.PutOperationReferenceParameters{
					Extents: []*scdpb.Volume4D{
						r.extents,
					},
					UssBaseUrl: r.url,
					NewSubscription: &scdpb.ImplicitSubscriptionParameters{
						UssBaseUrl: r.url,
					},
				},
			})
			require.Equal(t, r.wantErr, err)
			require.True(t, store.AssertExpectations(t))
		})
	}
}

func TestDeleteOperationRequiresOwnerInContext(t *testing.T) {
	var (
		id = uuid.New().String()
		ms = &mockStore{}
		s  = &Server{
			Store: ms,
		}
	)

	_, err := s.DeleteOperationReference(context.Background(), &scdpb.DeleteOperationReferenceRequest{
		Entityuuid: id,
	})

	require.Error(t, err)
	require.True(t, ms.AssertExpectations(t))
}

func TestDeleteOperation(t *testing.T) {
	var (
		owner   = dssmodels.Owner("foo")
		id      = scdmodels.ID(uuid.New().String())
		version = scdmodels.Version(1)
		ctx     = auth.ContextWithOwner(context.Background(), owner)
		ms      = &mockStore{}
		s       = &Server{
			Store: ms,
		}
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ms.On("DeleteOperation", mock.Anything, id, owner).Return(
		&scdmodels.Operation{
			ID:         scdmodels.ID(id),
			Owner:      dssmodels.Owner("me-myself-and-i"),
			USSBaseURL: "https://no/place/like/home",
			Version:    version,
		},
		[]*scdmodels.Subscription{
			{
				NotificationIndex: 42,
				BaseURL:           "https://no/place/like/home",
			},
		}, error(nil),
	)
	resp, err := s.DeleteOperationReference(ctx, &scdpb.DeleteOperationReferenceRequest{
		Entityuuid: id.String(),
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Subscribers, 1)
	require.True(t, ms.AssertExpectations(t))
}

func TestSearchIdentificationServiceAreas(t *testing.T) {
	var (
		ctx = auth.ContextWithOwner(context.Background(), dssmodels.Owner("foo"))
		ms  = &mockStore{}
		s   = &Server{
			Store: ms,
		}
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ms.On("SearchOperations", mock.Anything, mock.Anything, mock.Anything).Return(
		[]*scdmodels.Operation{
			{
				ID:         scdmodels.ID(uuid.New().String()),
				Owner:      dssmodels.Owner("me-myself-and-i"),
				USSBaseURL: "https://no/place/like/home",
			},
		}, error(nil),
	)
	resp, err := s.SearchOperationReferences(ctx, &scdpb.SearchOperationReferencesRequest{
		Params: &scdpb.SearchOperationReferenceParameters{
			AreaOfInterest: testdata.LoopVolume4D,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.OperationReferences, 1)
	require.True(t, ms.AssertExpectations(t))
}
