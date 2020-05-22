package application

import (
	"context"
	"database/sql"
	"flag"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridcrdb "github.com/interuss/dss/pkg/rid/cockroach"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

var (
	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

func setUpRepo(ctx context.Context, t *testing.T, logger *zap.Logger) (*ridcrdb.Store, func() error) {
	DefaultClock = fakeClock

	if len(*storeURI) == 0 {
		logger.Info("using the stubbed in memory store.")
		return &ridcrdb.Store{
			ISA: &isaStore{
				isas: make(map[dssmodels.ID]*ridmodels.IdentificationServiceArea),
			},
			Subscription: &subscriptionStore{
				subs: make(map[dssmodels.ID]*ridmodels.Subscription),
			},
		}, func() error { return nil }
	}
	ridcrdb.DefaultClock = fakeClock
	logger.Info("using cockroachDB.")

	// Use a real store.
	db, err := sql.Open("postgres", *storeURI)
	require.NoError(t, err)

	cdb := &cockroach.DB{DB: db}

	store := ridcrdb.NewStore(cdb, logger)
	require.NoError(t, store.Bootstrap(ctx))
	return store, func() error { return cdb.Close() }
}
