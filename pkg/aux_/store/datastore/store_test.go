package datastore

import (
	"context"
	"testing"

	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/datastore/flags"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

var (
	fakeClock = clockwork.NewFakeClock()
)

func setUpStore(ctx context.Context, t *testing.T) (*Store, func()) {
	connectParameters := flags.ConnectParameters()
	if connectParameters.Host == "" || connectParameters.Port == 0 {
		t.Skip()
	}
	// Reset the clock for every test.
	fakeClock = clockwork.NewFakeClock()

	store, err := newStore(ctx, t, connectParameters)
	require.NoError(t, err)
	return store, func() {
		require.NoError(t, CleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

func newStore(ctx context.Context, t *testing.T, connectParameters datastore.ConnectParameters) (*Store, error) {
	db, err := datastore.Dial(ctx, connectParameters)
	require.NoError(t, err)

	return &Store{
		db:    db,
		clock: fakeClock,
	}, nil
}

// CleanUp drops all required tables from the store, useful for testing.
func CleanUp(ctx context.Context, s *Store) error {
	const query = `
	DELETE FROM dss_metadata WHERE locality IS NOT NULL;
    `

	_, err := s.db.Pool.Exec(ctx, query)
	return err
}

// TestVersionFileIsRead checks that the schema version file is read and parsed.
func TestVersionIsParsed(t *testing.T) {
	_, err := getCurrentMajorSchemaVersion(currentCrdbSchemaVersionFile)
	require.NoError(t, err)

	_, err = getCurrentMajorSchemaVersion(currentYugabyteSchemaVersionFile)
	require.NoError(t, err)
}
