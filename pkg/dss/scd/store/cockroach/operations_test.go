package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/dss/models"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	"github.com/stretchr/testify/require"
)

var (
	overflow  = -1
	operation = &scdmodels.Operation{
		ID:         scdmodels.ID(uuid.New().String()),
		Owner:      dssmodels.Owner(uuid.New().String()),
		USSBaseURL: "https://no/place/like/home/for/flights",
		StartTime:  &startTime,
		EndTime:    &endTime,
		Cells: s2.CellUnion{
			s2.CellID(uint64(overflow)),
			s2.CellID(42),
		},
		SubscriptionID: scdmodels.ID(uuid.New().String()),
	}
)

func TestStoreSearchOperations(t *testing.T) {
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

	sub, _, err := store.UpsertSubscription(ctx, subscriptionsPool[0].input)
	require.NoError(t, err)

	in := *operation
	in.Cells = cells
	in.SubscriptionID = sub.ID
	out, _, err := store.UpsertOperation(ctx, &in, nil)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, in.ID, out.ID)

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
			earliest, latest := r.timestampMutator(*out.StartTime, *out.EndTime)

			operations, err := store.SearchOperations(ctx, &dssmodels.Volume4D{
				StartTime: earliest,
				EndTime:   latest,
				SpatialVolume: &dssmodels.Volume3D{
					Footprint: models.GeometryFunc(func() (s2.CellUnion, error) {
						return r.cells, nil
					}),
				},
			}, in.Owner)
			require.NoError(t, err)
			require.Len(t, operations, r.expectedLen)
		})
	}
}

func TestStoreExpiredOperation(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	sub, _, err := store.UpsertSubscription(ctx, subscriptionsPool[0].input)
	require.NoError(t, err)

	in := *operation
	in.SubscriptionID = sub.ID

	out, _, err := store.UpsertOperation(ctx, &in, nil)
	require.NoError(t, err)
	require.NotNil(t, out)

	// The ISA's endTime is one hour from now.
	fakeClock.Advance(59 * time.Minute)

	// We should still be able to find the ISA by searching and by ID.
	operations, err := store.SearchOperations(ctx, &dssmodels.Volume4D{
		SpatialVolume: &dssmodels.Volume3D{
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return operation.Cells, nil
			}),
		},
	}, operation.Owner)
	require.NoError(t, err)
	require.Len(t, operations, 1)

	ret, err := store.GetOperation(ctx, operation.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)

	// But now the ISA has expired.
	fakeClock.Advance(2 * time.Minute)

	operations, err = store.SearchOperations(ctx, &dssmodels.Volume4D{
		SpatialVolume: &dssmodels.Volume3D{
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return operation.Cells, nil
			}),
		},
	}, operation.Owner)
	require.NoError(t, err)
	require.Len(t, operations, 0)

	ret, err = store.GetOperation(ctx, operation.ID)
	require.Error(t, err)
	require.Nil(t, ret)
}

func TestStoreDeleteOperations(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	insertedSubscriptions := []*scdmodels.Subscription{}

	for _, r := range subscriptionsPool {
		copy := *r.input
		copy.Cells = s2.CellUnion{s2.CellID(42)}
		s1, _, err := store.UpsertSubscription(ctx, &copy)
		require.NoError(t, err)
		require.NotNil(t, s1)
		require.Equal(t, 42, s1.NotificationIndex)
		insertedSubscriptions = append(insertedSubscriptions, s1)
	}

	// Insert the Operation.
	copy := *operation
	copy.SubscriptionID = insertedSubscriptions[0].ID
	tx, _ := store.Begin()
	operation, subscriptionsOut, err := store.pushOperation(ctx, tx, &copy)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NotNil(t, operation)

	for i := range insertedSubscriptions {
		require.Equal(t, 43, subscriptionsOut[i].NotificationIndex)
	}
	// Can't delete with different owner.
	_, _, err = store.DeleteOperation(ctx, operation.ID, "bad-owner")
	require.Error(t, err)

	// Delete the Operation.
	operationOut, subscriptionsOut, err := store.DeleteOperation(ctx, operation.ID, operation.Owner)
	require.NoError(t, err)
	require.Equal(t, operation, operationOut)
	require.NotNil(t, subscriptionsOut)
	require.Len(t, subscriptionsOut, len(subscriptionsPool))
	for i, s := range subscriptionsPool {
		require.Equal(t, s.input.BaseURL, subscriptionsOut[i].BaseURL)
	}

	for i := range insertedSubscriptions {
		require.Equal(t, 44, subscriptionsOut[i].NotificationIndex)
	}
}

func TestUpsertOperation(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	sub, _, err := store.UpsertSubscription(ctx, subscriptionsPool[0].input)
	require.NoError(t, err)

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
			name:    "missing-start-time",
			endTime: fakeClock.Now().Add(-6 * time.Minute),
			wantErr: "rpc error: code = InvalidArgument desc = Operation must have an time_start",
		},
		{
			name:      "missing-end-time",
			startTime: fakeClock.Now().Add(-6 * time.Minute),
			wantErr:   "rpc error: code = InvalidArgument desc = Operation must have an time_end",
		},
		{
			name:      "start-time-in-the-past",
			startTime: fakeClock.Now().Add(-6 * time.Minute),
			endTime:   fakeClock.Now().Add(time.Hour),
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
			wantErr:   "rpc error: code = InvalidArgument desc = Operation time_end must be after time_start",
		},
		{
			name:                "updating-keeps-old-times",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(-6 * time.Hour),
			endTime:             fakeClock.Now().Add(6 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-start-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(3 * time.Hour),
			endTime:             fakeClock.Now().Add(6 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(3 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-end-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(-6 * time.Hour),
			endTime:             fakeClock.Now().Add(3 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(3 * time.Hour),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			id := scdmodels.ID(uuid.New().String())
			owner := dssmodels.Owner(uuid.New().String())
			var (
				ovn     scdmodels.OVN
				version scdmodels.Version
			)

			// Insert a pre-existing ISA to simulate updating from something.
			if !r.updateFromStartTime.IsZero() {
				tx, err := store.Begin()
				require.NoError(t, err)
				existing, _, err := store.pushOperation(ctx, tx, &scdmodels.Operation{
					ID:             id,
					Owner:          owner,
					USSBaseURL:     r.name,
					StartTime:      &r.updateFromStartTime,
					EndTime:        &r.updateFromEndTime,
					SubscriptionID: sub.ID,
				})
				require.NoError(t, err)
				require.NoError(t, tx.Commit())
				ovn = existing.OVN
				version = existing.Version

				// Can't update if it has a different owner
				op := *existing
				op.Owner = "bad-owner"
				_, _, err = store.UpsertOperation(ctx, &op, nil)
				require.Error(t, err)
			}

			op := &scdmodels.Operation{
				ID:             id,
				Owner:          owner,
				Version:        version,
				OVN:            ovn,
				SubscriptionID: sub.ID,
			}
			if !r.startTime.IsZero() {
				op.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				op.EndTime = &r.endTime
			}
			op, _, err := store.UpsertOperation(ctx, op, nil)

			if r.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, r.wantErr)
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, op.StartTime)
				require.Equal(t, r.wantStartTime, *op.StartTime)
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, op.EndTime)
				require.Equal(t, r.wantEndTime, *op.EndTime)
			}
		})
	}
}
