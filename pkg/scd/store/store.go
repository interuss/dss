package store

import (
	"context"

	"github.com/interuss/dss/pkg/scd/repos"
	scdsqlstore "github.com/interuss/dss/pkg/scd/store/sqlstore"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/store/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// scd.store.Store is a generic means to obtain an SCD scd.repos.Repository to perform SCD-specific
// operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]

// Init selects and initializes the scd store backend.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool, globalLock bool) (Store, error) {
	storeType := params.GetStoreParameters().StoreType
	switch storeType {
	case "sql":
		return scdsqlstore.Init(ctx, logger, withCheckCron, globalLock)
	default:
		return nil, stacktrace.NewError("Unsupported store type %q for scd", storeType)
	}
}
