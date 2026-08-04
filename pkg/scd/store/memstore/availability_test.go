package memstore

import (
	"testing"

	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestUssAvailabilityUpsertGet(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	got, err := r.UpsertUssAvailability(ctx, sampleAvailability())
	require.NoError(t, err)
	require.Equal(t, manager, got.Uss)
	require.Equal(t, scdmodels.UssAvailabilityStateNormal, got.Availability)
	require.Equal(t, scdmodels.OVN("HXjEfPAc0lkinCf0ejtSGiGPE4o2Qogm-iXAGPG-QNo_"), got.Version)

	fetched, err := r.GetUssAvailability(ctx, manager)
	require.NoError(t, err)
	require.Equal(t, got.Version, fetched.Version)
	require.Equal(t, scdmodels.UssAvailabilityStateNormal, fetched.Availability)
}

func TestGetUssAvailabilityMissingReturnsErrNoRows(t *testing.T) {
	r := setUpStore(t)
	_, err := r.GetUssAvailability(writeCtx(), manager)
	require.ErrorIs(t, err, pgx.ErrNoRows)
}
