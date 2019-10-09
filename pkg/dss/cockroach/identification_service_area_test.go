package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/dss/models"
	"github.com/stretchr/testify/require"
)

var (
	overflow    = -1
	serviceArea = &models.IdentificationServiceArea{
		ID:        models.ID(uuid.New().String()),
		Owner:     models.Owner(uuid.New().String()),
		Url:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells: s2.CellUnion{
			s2.CellID(uint64(overflow)),
			s2.CellID(42),
		},
	}
)

func TestStoreSearchISAs(t *testing.T) {
	var (
		ctx   = context.Background()
		cells = s2.CellUnion{
			s2.CellID(42),
			s2.CellID(84),
			s2.CellID(126),
			s2.CellID(168),
			s2.CellID(uint64(overflow)),
		}
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	isa := *serviceArea
	isa.Cells = cells
	saOut, _, err := store.InsertISA(ctx, isa)
	require.NoError(t, err)
	require.NotNil(t, saOut)
	require.Equal(t, isa.ID, saOut.ID)

	for _, r := range []struct {
		name             string
		cells            s2.CellUnion
		timestampMutator func(time.Time, time.Time) (*time.Time, *time.Time)
		expectedLen      int
	}{
		{
			name:  "search for empty cell",
			cells: s2.CellUnion{s2.CellID(210)},
			timestampMutator: func(time.Time, time.Time) (*time.Time, *time.Time) {
				return nil, nil
			},
			expectedLen: 0,
		},
		{
			name:  "search for only one cell",
			cells: s2.CellUnion{s2.CellID(42)},
			timestampMutator: func(time.Time, time.Time) (*time.Time, *time.Time) {
				return nil, nil
			},
			expectedLen: 1,
		},
		{
			name:  "search for only one cell with high bit set",
			cells: s2.CellUnion{s2.CellID(uint64(overflow))},
			timestampMutator: func(time.Time, time.Time) (*time.Time, *time.Time) {
				return nil, nil
			},
			expectedLen: 1,
		},
		{
			name:  "search with nil timestamps",
			cells: cells,
			timestampMutator: func(time.Time, time.Time) (*time.Time, *time.Time) {
				return nil, nil
			},
			expectedLen: 1,
		},
		{
			name:  "search with exact timestamps",
			cells: cells,
			timestampMutator: func(start time.Time, end time.Time) (*time.Time, *time.Time) {
				return &start, &end
			},
			expectedLen: 1,
		},
		{
			name:  "search with non-matching time span",
			cells: cells,
			timestampMutator: func(start time.Time, end time.Time) (*time.Time, *time.Time) {
				var (
					offset   = time.Duration(100 * time.Second)
					earliest = end.Add(offset)
					latest   = end.Add(offset * 2)
				)

				return &earliest, &latest
			},
			expectedLen: 0,
		},
		{
			name:  "search with expanded time span",
			cells: cells,
			timestampMutator: func(start time.Time, end time.Time) (*time.Time, *time.Time) {
				var (
					offset   = time.Duration(100 * time.Second)
					earliest = start.Add(-offset)
					latest   = end.Add(offset)
				)

				return &earliest, &latest
			},
			expectedLen: 1,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			earliest, latest := r.timestampMutator(*saOut.StartTime, *saOut.EndTime)

			serviceAreas, err := store.SearchISAs(ctx, r.cells, earliest, latest)
			require.NoError(t, err)
			require.Len(t, serviceAreas, r.expectedLen)
		})
	}
}

func TestStoreExpiredISA(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	saOut, _, err := store.InsertISA(ctx, *serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut)

	// The ISA's endTime is one hour from now.
	fakeClock.Advance(59 * time.Minute)

	// We should still be able to find the ISA by searching and by ID.
	serviceAreas, err := store.SearchISAs(ctx, serviceArea.Cells, nil, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 1)

	ret, err := store.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)

	// But now the ISA has expired.
	fakeClock.Advance(2 * time.Minute)

	serviceAreas, err = store.SearchISAs(ctx, serviceArea.Cells, nil, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 0)

	ret, err = store.GetISA(ctx, serviceArea.ID)
	require.Error(t, err)
	require.Nil(t, ret)
}

func TestStoreDeleteISAs(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	insertedSubscriptions := []*models.Subscription{}

	for _, r := range subscriptionsPool {
		copy := *r.input
		copy.Cells = s2.CellUnion{s2.CellID(42)}
		s1, err := store.InsertSubscription(ctx, copy)
		require.NoError(t, err)
		require.NotNil(t, s1)
		require.Equal(t, 42, s1.NotificationIndex)
		insertedSubscriptions = append(insertedSubscriptions, s1)
	}

	// Insert the ISA.
	copy := *serviceArea
	tx, _ := store.Begin()
	isa, subscriptionsOut, err := store.pushISA(ctx, tx, &copy)
	tx.Commit()
	require.NoError(t, err)
	require.NotNil(t, isa)

	for i, _ := range insertedSubscriptions {
		require.Equal(t, 43, subscriptionsOut[i].NotificationIndex)
	}

	// Delete the ISA.
	serviceAreaOut, subscriptionsOut, err := store.DeleteISA(ctx, isa.ID, isa.Owner, isa.Version)
	require.NoError(t, err)
	require.Equal(t, isa, serviceAreaOut)
	require.NotNil(t, subscriptionsOut)
	require.Len(t, subscriptionsOut, len(subscriptionsPool))
	for i, s := range subscriptionsPool {
		require.Equal(t, s.input.Url, subscriptionsOut[i].Url)
	}

	for i, _ := range insertedSubscriptions {
		require.Equal(t, 44, subscriptionsOut[i].NotificationIndex)
	}
}

func TestInsertISA(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

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
			name:          "start-time-slighty-in-the-past",
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
			id := models.ID(uuid.New().String())
			owner := models.Owner(uuid.New().String())
			var version *models.Version

			// Insert a pre-existing ISA to simulate updating from something.
			if !r.updateFromStartTime.IsZero() {
				tx, err := store.Begin()
				require.NoError(t, err)
				existing, _, err := store.pushISA(ctx, tx, &models.IdentificationServiceArea{
					ID:        id,
					Owner:     owner,
					StartTime: &r.updateFromStartTime,
					EndTime:   &r.updateFromEndTime,
				})
				require.NoError(t, err)
				require.NoError(t, tx.Commit())
				version = existing.Version
			}

			sa := models.IdentificationServiceArea{
				ID:      id,
				Owner:   owner,
				Version: version,
			}
			if !r.startTime.IsZero() {
				sa.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				sa.EndTime = &r.endTime
			}
			isa, _, err := store.InsertISA(ctx, sa)

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
