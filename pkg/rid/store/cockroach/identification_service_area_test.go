package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/stretchr/testify/require"
)

var (
	// Ensure the struct conforms to the interface
	_           repos.ISA = &repo{}
	overflow              = uint64(17106221850767130624) // face 5 L13 overflows
	serviceArea           = &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Writer:    writer,
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

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	isa := *serviceArea
	isa.Cells = cells
	saOut, err := repo.InsertISA(ctx, &isa)
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

			serviceAreas, err := repo.SearchISAs(ctx, r.cells, earliest, latest)
			require.NoError(t, err)
			require.Len(t, serviceAreas, r.expectedLen)
		})
	}
}

func TestBadVersion(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	saOut1, err := repo.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut1)

	// Rewriting service area should fail
	saOut2, err := repo.UpdateISA(ctx, serviceArea)
	require.NoError(t, err)
	require.Nil(t, saOut2)

	// Rewriting, but with the correct version should work.
	newEndTime := saOut1.EndTime.Add(time.Minute)
	saOut1.EndTime = &newEndTime
	saOut3, err := repo.UpdateISA(ctx, saOut1)
	require.NoError(t, err)
	require.NotNil(t, saOut3)
}

func TestStoreExpiredISA(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	saOut, err := repo.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut)

	// The ISA's endTime is one hour from now.
	fakeClock.Advance(59 * time.Minute)

	// We should still be able to find the ISA by searching and by ID.
	now := fakeClock.Now()
	serviceAreas, err := repo.SearchISAs(ctx, serviceArea.Cells, &now, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 1)

	ret, err := repo.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)

	// But now the ISA has expired.
	fakeClock.Advance(2 * time.Minute)
	now = fakeClock.Now()

	serviceAreas, err = repo.SearchISAs(ctx, serviceArea.Cells, &now, nil)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 0)

	// A get should work even if it is expired.
	ret, err = repo.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)
}

func TestStoreDeleteISAs(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	// Insert the ISA.
	copy := *serviceArea
	isa, err := repo.InsertISA(ctx, &copy)
	require.NoError(t, err)
	require.NotNil(t, isa)

	// Delete the ISA.
	// Ensure a fresh Get, then delete still updates the sub indexes
	isa, err = repo.GetISA(ctx, isa.ID)
	require.NoError(t, err)

	serviceAreaOut, err := repo.DeleteISA(ctx, isa)
	require.NoError(t, err)
	require.Equal(t, isa, serviceAreaOut)
}

func TestStoreISAWithNoGeoData(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.IdentificationServiceArea{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		EndTime: &endTime,
	}
	_, err = repo.InsertISA(ctx, sub)
	require.Error(t, err)
}

func TestListExpiredISAs(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	fakeClock := clockwork.NewFakeClockAt(time.Now())

	// Insert ISA with endtime 1 day from now
	isa1 := *serviceArea
	startTime := fakeClock.Now()
	isa1.StartTime = &startTime
	endTime := fakeClock.Now().Add(24 * time.Hour)
	isa1.EndTime = &endTime
	saOut1, err := repo.InsertISA(ctx, &isa1)
	require.NoError(t, err)
	require.NotNil(t, saOut1)

	// Insert ISA with endtime to 30 minutes ago
	isa2 := *serviceArea
	startTime = fakeClock.Now().Add(-1 * time.Hour)
	isa2.StartTime = &startTime
	endTime = fakeClock.Now().Add(-30 * time.Minute)
	isa2.EndTime = &endTime
	isa2.ID = dssmodels.ID(uuid.New().String())
	saOut2, err := repo.InsertISA(ctx, &isa2)
	require.NoError(t, err)
	require.NotNil(t, saOut2)

	serviceAreas, err := repo.ListExpiredISAs(ctx, &writer)
	require.NoError(t, err)
	require.Len(t, serviceAreas, 1)
}
