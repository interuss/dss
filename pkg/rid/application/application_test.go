package application

import (
	"context"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/datastore/flags"
	"github.com/interuss/dss/pkg/logging"
	ridc "github.com/interuss/dss/pkg/rid/store/datastore"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

var (
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

func setUpStore(ctx context.Context, t *testing.T) (*ridc.Store, func()) {

	DefaultClock = fakeClock

	connectParameters := flags.ConnectParameters()
	if connectParameters.Host == "" || connectParameters.Port == 0 {
		t.Skip()
	}
	connectParameters.DBName = "rid"

	store, err := newStore(ctx, t, connectParameters)
	require.NoError(t, err)
	return store, func() {
		require.NoError(t, CleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

func newStore(ctx context.Context, t *testing.T, connectParameters datastore.ConnectParameters) (*ridc.Store, error) {
	db, err := datastore.Dial(ctx, connectParameters)
	require.NoError(t, err)

	s, err := ridc.NewStore(ctx, db, logging.Logger)
	if err != nil {
		return nil, err
	}
	s.Clock = fakeClock

	return s, nil
}

// CleanUp drops all required tables from the store, useful for testing.
func CleanUp(ctx context.Context, s *ridc.Store) error {
	return s.CleanUp(ctx)
}
