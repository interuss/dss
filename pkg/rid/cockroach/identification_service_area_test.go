package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/stretchr/testify/require"
)

var (
	// Ensure the struct conforms to the interface
	_           repos.ISA = &ISAStore{}
	overflow              = -1
	serviceArea           = &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
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
	saOut, _, err := store.ISA.Insert(ctx, &isa)
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

			serviceAreas, err := store.ISA.Search(ctx, r.cells, earliest, latest)
			require.NoError(t, err)
			require.Len(t, serviceAreas, r.expectedLen)
		})
	}
}

func TestBadVersion(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	saOut1, _, err := store.ISA.Insert(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut1)

	// Rewriting service area should fail
	saOut2, _, err := store.ISA.Insert(ctx, serviceArea)
	require.Error(t, err)
	require.Nil(t, saOut2)

	// Rewriting, but with the correct version should work.
	newEndTime := saOut1.EndTime.Add(time.Minute)
	saOut1.EndTime = &newEndTime
	saOut3, _, err := store.ISA.Insert(ctx, saOut1)
	require.NoError(t, err)
	require.NotNil(t, saOut3)
}

func TestStoreExpiredISA(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	saOut, _, err := store.ISA.Insert(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut)

	// The ISA's endTime is one hour from now.
	fakeClock.Advance(59 * time.Minute)

	// We should still be able to find the ISA by searching and by ID.
	serviceAreas, err := store.ISA.Search(ctx, serviceArea.Cells, nil, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 1)

	ret, err := store.ISA.Get(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)

	// But now the ISA has expired.
	fakeClock.Advance(2 * time.Minute)

	serviceAreas, err = store.ISA.Search(ctx, serviceArea.Cells, nil, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 0)

	ret, err = store.ISA.Get(ctx, serviceArea.ID)
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

	insertedSubscriptions := []*ridmodels.Subscription{}
	for _, r := range subscriptionsPool {
		copy := *r.input
		copy.Cells = s2.CellUnion{s2.CellID(42)}
		s1, err := store.Subscription.Insert(ctx, &copy)
		require.NoError(t, err)
		require.NotNil(t, s1)
		require.Equal(t, 42, s1.NotificationIndex)
		insertedSubscriptions = append(insertedSubscriptions, s1)
	}

	// Insert the ISA.
	copy := *serviceArea
	isa, subscriptionsOut, err := store.ISA.Insert(ctx, &copy)
	require.NoError(t, err)
	require.NotNil(t, isa)

	for i := range insertedSubscriptions {
		require.Equal(t, 43, subscriptionsOut[i].NotificationIndex)
	}
	// Can't delete with different owner.
	_, _, err = store.ISA.Delete(ctx, isa.ID, "bad-owner", isa.Version)
	require.Error(t, err)

	// Delete the ISA.
	serviceAreaOut, subscriptionsOut, err := store.ISA.Delete(ctx, isa.ID, isa.Owner, isa.Version)
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
