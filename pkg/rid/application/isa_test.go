package application

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	_ ISAApp = &app{}
)

func setUpISAApp(ctx context.Context, t *testing.T) (*app, func()) {
	l := zap.L()
	transactor, cleanup := setUpStore(ctx, t, l)
	return NewFromTransactor(transactor, l).(*app), cleanup
}

// TODO:steeling add owner logic.
type isaStore struct {
	isas map[dssmodels.ID]*ridmodels.IdentificationServiceArea
}

func (store *isaStore) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	if isa, ok := store.isas[id]; ok {
		return isa, nil
	}
	return nil, nil
}

// Implements repos.ISA.DeleteISA
func (store *isaStore) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	isa, ok := store.isas[isa.ID]
	if !ok {
		return nil, nil
	}
	delete(store.isas, isa.ID)

	return isa, nil
}

// Implements repos.ISA.InsertISA
func (store *isaStore) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	storedCopy := *isa
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.isas[isa.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, nil
}

// Implements repos.ISA.UpdateISA
func (store *isaStore) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea, version string) (*ridmodels.IdentificationServiceArea, error) {
	storedCopy := *isa
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.isas[isa.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (store *isaStore) GetVersion(ctx context.Context) (string, error) {
	return "v3.1.0", nil
}

// Implements repos.ISA.SearchISA
func (store *isaStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	var isas []*ridmodels.IdentificationServiceArea

	for _, isa := range store.isas {
		if isa.Cells.Intersects(cells) {
			isas = append(isas, isa)
		}
	}
	return isas, nil
}

func TestISAUpdateIdxCells(t *testing.T) {
	ctx := context.Background()
	app, cleanup := setUpISAApp(ctx, t)

	defer cleanup()
	// ensure that when we do an update, nothing in the s2 library joins multiple
	// cells together at a lower level.

	// These 4 cells are fully encompassed by the parent cell, meaning the s2
	// library might try to Normalize (this is the name of the function) the Union
	// into a single cell. We don't support this currently, so let's make sure
	// this doesn't happen.
	isa, _, err := app.InsertISA(ctx, &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     "owner",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{17106221850767130624, 17106221885126868992, 17106221919486607360},
	})
	require.NoError(t, err)
	require.NotNil(t, isa)
	// Now insert 2 subs, one overlaps with the original isa, and the second, overlaps
	// with the soon to be new version of the isa. both should increase their
	// notification index.

	_, err = app.InsertSubscription(ctx, &ridmodels.Subscription{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     "owner",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{17106221850767130624, 17106221919486607360},
	})
	require.NoError(t, err)

	_, err = app.InsertSubscription(ctx, &ridmodels.Subscription{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     "owner",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{17106221953846345728},
	})
	require.NoError(t, err)

	isa.Cells = s2.CellUnion{17106221953846345728}

	isa, subs, err := app.UpdateISA(ctx, isa)
	require.NoError(t, err)
	require.NotNil(t, isa)
	require.Len(t, subs, 2)
	for _, sub := range subs {
		require.Equal(t, 1, sub.NotificationIndex)
	}

	isas, err := app.SearchISAs(ctx, isa.Cells, &startTime, nil)
	require.NoError(t, err)
	require.NotNil(t, isas)
	require.Len(t, isas, 1)
}

func TestInsertISA(t *testing.T) {
	ctx := context.Background()
	app, cleanup := setUpISAApp(ctx, t)

	defer cleanup()

	for _, r := range []struct {
		name          string
		startTime     time.Time
		endTime       time.Time
		wantErr       string
		wantStartTime time.Time
		wantEndTime   time.Time
	}{
		{
			name:    "missing-end-time",
			wantErr: "rpc error: code = InvalidArgument desc = IdentificationServiceArea must have an time_end",
		},
		{
			name:          "start-time-defaults-to-now",
			endTime:       fakeClock.Now().Add(time.Hour),
			wantStartTime: fakeClock.Now(),
		},
		{
			name:      "start-time-in-the-past",
			startTime: fakeClock.Now().Add(-6 * time.Minute),
			endTime:   fakeClock.Now().Add(time.Hour),
			wantErr:   "rpc error: code = InvalidArgument desc = IdentificationServiceArea time_start must not be in the past",
		},
		{
			name:          "start-time-slightly-in-the-past",
			startTime:     fakeClock.Now().Add(-4 * time.Minute),
			endTime:       fakeClock.Now().Add(time.Hour),
			wantStartTime: fakeClock.Now().Add(-4 * time.Minute),
		},
		{
			name:      "end-time-before-start-time",
			startTime: fakeClock.Now().Add(20 * time.Minute),
			endTime:   fakeClock.Now().Add(10 * time.Minute),
			wantErr:   "rpc error: code = InvalidArgument desc = IdentificationServiceArea time_end must be after time_start",
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			sa := &ridmodels.IdentificationServiceArea{
				ID:    dssmodels.ID(uuid.New().String()),
				Owner: dssmodels.Owner(uuid.New().String()),
				Cells: s2.CellUnion{12494535935418957824},
			}
			if !r.startTime.IsZero() {
				sa.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				sa.EndTime = &r.endTime
			}
			isa, _, err := app.InsertISA(ctx, sa)

			if r.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, r.wantErr)
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, isa.StartTime)
				require.Equal(t, r.wantStartTime, *isa.StartTime)
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, isa.EndTime)
				require.Equal(t, r.wantEndTime, *isa.EndTime)
			}
		})
	}
}

func TestUpdateISA(t *testing.T) {
	ctx := context.Background()
	app, cleanup := setUpISAApp(ctx, t)

	defer cleanup()

	for _, r := range []struct {
		name                string
		updateFromStartTime time.Time
		updateFromEndTime   time.Time
		startTime           time.Time
		endTime             time.Time
		wantErr             string
		wantStartTime       time.Time
		wantEndTime         time.Time
	}{
		{
			name:                "updating-keeps-old-times",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-start-time-to-past",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(-3 * time.Hour),
			wantErr:             "rpc error: code = InvalidArgument desc = IdentificationServiceArea time_start must not be in the past",
		},
		{
			name:                "changing-start-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(3 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(3 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-end-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			endTime:             fakeClock.Now().Add(3 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(3 * time.Hour),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			id := dssmodels.ID(uuid.New().String())
			owner := dssmodels.Owner(uuid.New().String())
			repo, err := app.Store.Interact(ctx)
			require.NoError(t, err)
			// Insert a pre-existing ISA to simulate updating from something.
			existing, err := repo.InsertISA(ctx, &ridmodels.IdentificationServiceArea{
				ID:        id,
				Owner:     owner,
				StartTime: &r.updateFromStartTime,
				EndTime:   &r.updateFromEndTime,
				Cells:     s2.CellUnion{12494535935418957824},
			})
			require.NoError(t, err)

			sa := &ridmodels.IdentificationServiceArea{
				ID:      id,
				Owner:   owner,
				Version: existing.Version,
				Cells:   s2.CellUnion{12494535935418957824},
			}
			if !r.startTime.IsZero() {
				sa.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				sa.EndTime = &r.endTime
			}
			isa, _, err := app.UpdateISA(ctx, sa)

			if r.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, r.wantErr)
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, isa.StartTime)
				require.Equal(t, r.wantStartTime, *isa.StartTime)
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, isa.EndTime)
				require.Equal(t, r.wantEndTime, *isa.EndTime)
			}
		})
	}
}

func TestAppDeleteISAs(t *testing.T) {
	var (
		ctx          = context.Background()
		app, cleanup = setUpISAApp(ctx, t)
	)
	defer cleanup()

	insertedSubscriptions := []*ridmodels.Subscription{}
	for _, r := range subscriptionsPool {
		copy := *r.input
		s1, err := app.InsertSubscription(ctx, &copy)
		require.NoError(t, err)
		require.NotNil(t, s1)
		require.Equal(t, 42, s1.NotificationIndex)
		insertedSubscriptions = append(insertedSubscriptions, s1)
	}
	serviceArea := &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells: s2.CellUnion{
			s2.CellID(12494535935418957824),
		},
	}

	// Insert the ISA.
	copy := *serviceArea
	isa, subscriptionsOut, err := app.InsertISA(ctx, &copy)
	require.NoError(t, err)
	require.NotNil(t, isa)
	require.Len(t, subscriptionsOut, len(insertedSubscriptions))

	for i := range insertedSubscriptions {
		require.Equal(t, 43, subscriptionsOut[i].NotificationIndex)
	}
	// Can't delete with different owner.
	_, _, err = app.DeleteISA(ctx, isa.ID, "bad-owner", isa.Version)
	require.Error(t, err)

	// Delete the ISA.
	// Ensure a fresh Get, then delete still updates the subscription indexes
	isa, err = app.GetISA(ctx, isa.ID)
	require.NoError(t, err)

	serviceAreaOut, subscriptionsOut, err := app.DeleteISA(ctx, isa.ID, isa.Owner, isa.Version)
	require.NoError(t, err)
	require.Equal(t, isa, serviceAreaOut)
	require.NotNil(t, subscriptionsOut)
	require.Len(t, subscriptionsOut, len(subscriptionsPool))
	for i, s := range subscriptionsPool {
		require.Equal(t, s.input.URL, subscriptionsOut[i].URL)
	}

	for i := range insertedSubscriptions {
		require.Equal(t, 44, subscriptionsOut[i].NotificationIndex)
	}
}
