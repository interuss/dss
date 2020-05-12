package scd

import (
  "context"
  "errors"
  testdata "github.com/interuss/dss/pkg/dss/geo/testdata/scd"

  "github.com/interuss/dss/pkg/api/v1/scdpb"
  dsserr "github.com/interuss/dss/pkg/errors"
  "testing"
  "time"

  "github.com/interuss/dss/pkg/dss/auth"
  dssmodels "github.com/interuss/dss/pkg/dss/models"
  scdmodels "github.com/interuss/dss/pkg/dss/scd/models"

  "github.com/golang/geo/s2"
  "github.com/google/uuid"
  "github.com/stretchr/testify/mock"
  "github.com/stretchr/testify/require"
)

var timeout = time.Second * 10

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

func (ms *mockStore) UpsertSubscription(ctx context.Context, subscription *scdmodels.Subscription) (*scdmodels.Subscription, error) {
  ctx, cancel := context.WithTimeout(ctx, timeout)
  defer cancel()
  args := ms.Called(ctx, subscription)
  return args.Get(0).(*scdmodels.Subscription), args.Error(1)
}

func (ms *mockStore) InsertSubscription(ctx context.Context, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {
  ctx, cancel := context.WithTimeout(ctx, timeout)
  defer cancel()
  args := ms.Called(ctx, s)
  return args.Get(0).(*scdmodels.Subscription), args.Error(1)
}

func (ms *mockStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*scdmodels.Subscription, error) {
  ctx, cancel := context.WithTimeout(ctx, timeout)
  defer cancel()
  args := ms.Called(ctx, cells, owner)
  return args.Get(0).([]*scdmodels.Subscription), args.Error(1)
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
  validAoi, _ := scdpb.FromVolume4D(testdata.LoopVolume4D)
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
  extentsLoop, _ := scdpb.FromVolume4D(testdata.LoopVolume4D)
  for _, r := range []struct {
    name             string
    id               dssmodels.ID
    owner            dssmodels.Owner
    oldVersion       int32
    ussBaseUrl       string
    extents          *scdpb.Volume4D
    wantSubscription *scdmodels.Subscription
    wantErr          error
  }{
    {
      name:             "missing-owner",
      id:               dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
      oldVersion:       0,
      ussBaseUrl:       "https://example.com",
      extents:          extentsLoop,
      wantSubscription: validSubscription,
      wantErr:          dsserr.PermissionDenied("missing owner from context"),
    },
    {
      name:             "missing-id",
      owner:            dssmodels.Owner("foo"),
      oldVersion:       0,
      ussBaseUrl:       "https://example.com",
      extents:          extentsLoop,
      wantSubscription: validSubscription,
      wantErr:          dsserr.BadRequest("missing Subscription ID"),
    },
    {
      name:             "success",
      id:               dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
      owner:            dssmodels.Owner("foo"),
      oldVersion:       0,
      ussBaseUrl:       "https://example.com",
      extents:          extentsLoop,
      wantSubscription: validSubscription,
    },
    /*{
      			name: "missing-extents",
      			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
            owner: dssmodels.Owner("foo"),
      			ussBaseUrl: "https://example.com",
      			wantErr: dsserr.BadRequest("missing required extents"),
      		},
          {
      			name: "missing-extents-spatial-volume",
      			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
            owner: dssmodels.Owner("foo"),
            ussBaseUrl: "https://example.com",
      			extents: &scdpb.Volume4D{},
      			wantErr: dsserr.BadRequest("bad extents: missing required spatial_volume"),
      		},
      		{
      			name: "missing-spatial-volume-footprint",
      			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
            owner: dssmodels.Owner("foo"),
            ussBaseUrl: "https://example.com",
      			extents: &scdpb.Volume4D{
      				Volume: &scdpb.Volume3D{},
      			},
      			wantErr: dsserr.BadRequest("bad extents: spatial_volume missing required footprint"),
      		},
      		{
      			name: "missing-spatial-polygon",
      			id:   dssmodels.ID("4348c8e5-0b1c-43cf-9114-2e67a4532765"),
            owner: dssmodels.Owner("foo"),
            ussBaseUrl: "https://example.com",
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
      ctx := context.Background()
      if r.owner != "" {
        ctx = auth.ContextWithOwner(ctx, r.owner)
      }
      store := &mockStore{}
      if r.wantErr == nil {
        store.On("UpsertSubscription", mock.Anything, r.wantSubscription).Return(
          r.wantSubscription, nil,
        )
      }
      s := &Server{
        Store: store,
      }

      resp, err := s.PutSubscription(ctx, &scdpb.PutSubscriptionRequest{
        Subscriptionid: r.id.String(),
        Params: &scdpb.PutSubscriptionParameters{
          UssBaseUrl:           r.ussBaseUrl,
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
