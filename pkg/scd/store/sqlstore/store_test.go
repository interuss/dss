package sqlstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/dss/pkg/sqlstore"
	"github.com/interuss/dss/pkg/sqlstore/params"
	"github.com/interuss/stacktrace"
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

	store, err := newTestStore(ctx)
	require.NoError(t, err)
	return store, func() {
		require.NoError(t, cleanUp(ctx, store))
		require.NoError(t, store.Shutdown(ctx))
	}
}

func newTestStore(ctx context.Context) (*sqlstore.Store[repos.Repository], error) {
	s, err := Init(ctx, logging.Logger, false)

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

func TestSearchMoreThanMaxResultLimit(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
		cell                 = s2.CellID(17106221850767130624)
		now                  = fakeClock.Now()
		end                  = now.Add(time.Hour)
		volume               = &dssmodels.CellsVolume4D{Cells: s2.CellUnion{cell}, StartTime: &now, EndTime: &end}
	)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	// Inserting this many rows through the repo one at a time takes minutes. A subscription must
	// notify for something.
	insert := func(count int) {
		for table, columns := range map[string]string{
			"scd_operations":    "",
			"scd_constraints":   "",
			"scd_subscriptions": ", notify_for_operations",
		} {
			values := ""
			if columns != "" {
				values = ", true"
			}
			_, err := store.Pool.Exec(ctx, fmt.Sprintf(`
				INSERT INTO %s (id, owner, url, cells, starts_at, ends_at, updated_at%s)
				SELECT gen_random_uuid(), 'me', 'https://uss', ARRAY[$1::INT8], $2, $3, $2%s
				FROM generate_series(1, $4)`, table, columns, values),
				int64(cell), now, end, count)
			require.NoError(t, err)
		}
	}

	// Exactly MaxResultLimit matches are returned in full.
	insert(dssmodels.MaxResultLimit)
	operationalIntents, err := repo.SearchOperationalIntents(ctx, volume)
	require.NoError(t, err)
	require.Len(t, operationalIntents, dssmodels.MaxResultLimit)
	constraints, err := repo.SearchConstraints(ctx, volume)
	require.NoError(t, err)
	require.Len(t, constraints, dssmodels.MaxResultLimit)
	subscriptions, err := repo.SearchSubscriptions(ctx, volume)
	require.NoError(t, err)
	require.Len(t, subscriptions, dssmodels.MaxResultLimit)

	// One more is an error rather than a truncated list.
	insert(1)
	_, err = repo.SearchOperationalIntents(ctx, volume)
	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(err))
	_, err = repo.SearchConstraints(ctx, volume)
	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(err))
	_, err = repo.SearchSubscriptions(ctx, volume)
	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(err))
}
