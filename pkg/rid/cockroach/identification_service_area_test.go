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
	_           repos.ISA = &Store{}
	overflow              = uint64(17106221850767130624) // face 5 L13 overflows
	serviceArea           = &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells: s2.CellUnion{
			s2.CellID(uint64(overflow)),
			s2.CellID(17106221850767130624),
		},
	}
)

func TestStoreSearchISAs(t *testing.T) {
	var (
		ctx   = context.Background()
		cells = s2.CellUnion{
			s2.CellID(17106221850767130624),
			s2.CellID(17106221885126868992),
			s2.CellID(17106221919486607360),
			s2.CellID(uint64(overflow)),
		}
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer tearDownStore()

	isa := *serviceArea
	isa.Cells = cells
	saOut, err := store.InsertISA(ctx, &isa)
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
			cells: s2.CellUnion{s2.CellID(17106221953846345728)},
			timestampMutator: func(start time.Time, end time.Time) (*time.Time, *time.Time) {
				return &start, nil
			},
			expectedLen: 0,
		},
		{
			name:  "search for only one cell",
			cells: s2.CellUnion{s2.CellID(17106221850767130624)},
			timestampMutator: func(start time.Time, end time.Time) (*time.Time, *time.Time) {
				return &start, nil
			},
			expectedLen: 1,
		},
		{
			name:  "search for only one cell with high bit set",
			cells: s2.CellUnion{s2.CellID(uint64(overflow))},
			timestampMutator: func(start time.Time, end time.Time) (*time.Time, *time.Time) {
				return &start, nil
			},
			expectedLen: 1,
		},
		{
			name:  "search with nil ends_at",
			cells: cells,
			timestampMutator: func(start time.Time, end time.Time) (*time.Time, *time.Time) {
				return &start, nil
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

func TestBadVersion(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	saOut1, err := store.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut1)

	// Rewriting service area should fail
	saOut2, err := store.InsertISA(ctx, serviceArea)
	require.Error(t, err)
	require.Nil(t, saOut2)

	// Rewriting, but with the correct version should work.
	newEndTime := saOut1.EndTime.Add(time.Minute)
	saOut1.EndTime = &newEndTime
	saOut3, err := store.InsertISA(ctx, saOut1)
	require.NoError(t, err)
	require.NotNil(t, saOut3)
}

func TestStoreExpiredISA(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	saOut, err := store.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut)

	// The ISA's endTime is one hour from now.
	fakeClock.Advance(59 * time.Minute)

	// We should still be able to find the ISA by searching and by ID.
	now := fakeClock.Now()
	serviceAreas, err := store.SearchISAs(ctx, serviceArea.Cells, &now, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 1)

	ret, err := store.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)

	// But now the ISA has expired.
	fakeClock.Advance(2 * time.Minute)
	now = fakeClock.Now()

	serviceAreas, err = store.SearchISAs(ctx, serviceArea.Cells, &now, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 0)

	// A get should work even if it is expired.
	ret, err = store.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)
}

func TestStoreDeleteISAs(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer tearDownStore()

	// Insert the ISA.
	copy := *serviceArea
	isa, err := store.InsertISA(ctx, &copy)
	require.NoError(t, err)
	require.NotNil(t, isa)

	// Can't delete with different owner.
	iCopy := *isa
	iCopy.Owner = "bad-owner"
	_, err = store.DeleteISA(ctx, &iCopy)
	require.Error(t, err)

	// Delete the ISA.
	// Ensure a fresh Get, then delete still updates the sub indexes
	isa, err = store.GetISA(ctx, isa.ID)
	require.NoError(t, err)

	serviceAreaOut, err := store.DeleteISA(ctx, isa)
	require.NoError(t, err)
	require.Equal(t, isa, serviceAreaOut)
}

func TestStoreISAWithNoGeoData(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.IdentificationServiceArea{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		EndTime: &endTime,
	}
	_, err := store.InsertISA(ctx, sub)
	require.Error(t, err)
}
