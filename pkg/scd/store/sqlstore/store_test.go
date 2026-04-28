package sqlstore

import (
	"context"
	"testing"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/dss/pkg/sqlstore"
	"github.com/interuss/dss/pkg/sqlstore/params"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

var (
	fakeClock = clockwork.NewFakeClock()
)

func setUpStore(ctx context.Context, t *testing.T) (*sqlstore.Store[repos.Repository], func()) {
	connectParameters := params.GetConnectParameters()
	if connectParameters.Host == "" || connectParameters.Port == 0 {
		t.Skip()
	}
	connectParameters.DBName = "scd"
	// Reset the clock for every test.
	fakeClock = clockwork.NewFakeClock()

	store, err := newTestStore(ctx, t, connectParameters)
	require.NoError(t, err)
	return store, func() {
		require.NoError(t, cleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

func newTestStore(ctx context.Context, t *testing.T, connectParameters params.ConnectParameters) (*sqlstore.Store[repos.Repository], error) {
	s, err := Init(ctx, logging.Logger, false, false)

	if err != nil {
		return nil, err
	}
	s.Clock = fakeClock

	return s, nil
}

// cleanUp drops all required tables from the store, useful for testing.
func cleanUp(ctx context.Context, s *sqlstore.Store[repos.Repository]) error {
	const query = `
	DELETE FROM scd_subscriptions WHERE id IS NOT NULL;
	DELETE FROM scd_operations WHERE id IS NOT NULL;
	DELETE FROM scd_constraints WHERE id IS NOT NULL;
	DELETE FROM scd_uss_availability WHERE id IS NOT NULL;`

	_, err := s.Pool.Exec(ctx, query)
	return err
}
