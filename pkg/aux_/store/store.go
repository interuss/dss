package store

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	auxsqlstore "github.com/interuss/dss/pkg/aux_/store/sqlstore"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/store/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// aux_.store.Store is a generic means to obtain an aux Repository (repo containing auxiliary
// information not related to standardized services like RID or SCD specifically) to perform
// aux-specific operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]

// Init selects and initializes the aux store backend.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool) (Store, error) {
	switch storeType := params.GetStoreParameters().StoreType; storeType {
	case "sql":
		return auxsqlstore.Init(ctx, logger, withCheckCron)
	default:
		return nil, stacktrace.NewError("Unsupported store type %q for aux", storeType)
	}
}
