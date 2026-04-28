package store

import (
	"context"

	"github.com/interuss/dss/pkg/scd/repos"
	scddatastore "github.com/interuss/dss/pkg/scd/store/datastore"
	dssstore "github.com/interuss/dss/pkg/store"
	"go.uber.org/zap"
)

// scd.store.Store is a generic means to obtain an SCD scd.repos.Repository to perform SCD-specific
// operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]

// Init selects and initializes the scd store backend.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool, globalLock bool) (Store, error) {
	return scddatastore.Init(ctx, logger, withCheckCron, globalLock)
}
