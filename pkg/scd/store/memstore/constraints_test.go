package memstore

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestConstraintUpsertGetDelete(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	got, err := r.UpsertConstraint(ctx, sampleConstraint())
	require.NoError(t, err)
	require.Equal(t, constraintId, got.ID)
	require.Equal(t, manager, got.Manager)
	require.NotEmpty(t, got.OVN)

	fetched, err := r.GetConstraint(ctx, constraintId)
	require.NoError(t, err)
	require.Equal(t, got.OVN, fetched.OVN)
	require.Equal(t, cells, fetched.Cells)

	count, err := r.CountConstraints(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	require.NoError(t, r.DeleteConstraint(ctx, constraintId))

	_, err = r.GetConstraint(ctx, constraintId)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestConstraintGetMissingReturnsErrNoRows(t *testing.T) {
	r := setUpStore(t)
	_, err := r.GetConstraint(writeCtx(), constraintId)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestConstraintDeleteMissingReturnsErrNoRows(t *testing.T) {
	r := setUpStore(t)
	err := r.DeleteConstraint(writeCtx(), constraintId)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestSearchConstraints(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)
	_, err := r.UpsertConstraint(ctx, sampleConstraint())
	require.NoError(t, err)

	// Overlapping volume with no time bounds matches.
	res, err := r.SearchConstraints(ctx, volume4D(cells, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Len(t, res, 1)

	// Time window after the constraint's end excludes it.
	afterStart := endTime.Add(time.Hour)
	afterEnd := afterStart.Add(time.Hour)
	res, err = r.SearchConstraints(ctx, volume4D(cells, &afterStart, &afterEnd, nil, nil))
	require.NoError(t, err)
	require.Empty(t, res)

	// No covering cells returns an empty (non-nil) slice.
	res, err = r.SearchConstraints(ctx, volume4D(s2.CellUnion{}, nil, nil, nil, nil))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Empty(t, res)
}
