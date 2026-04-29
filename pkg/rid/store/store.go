package store

import (
	"context"

	"github.com/interuss/dss/pkg/rid/repos"
	ridsqlstore "github.com/interuss/dss/pkg/rid/store/sqlstore"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/store/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// rid.store.Store is a generic means to obtain an RID rid.repos.Repository to perform RID-specific
// operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]

// Init selects and initializes the rid store backend.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool) (Store, error) {
	storeType := params.GetStoreParameters().StoreType
	switch storeType {
	case "sql":
		return ridsqlstore.Init(ctx, logger, withCheckCron)
	case "raft":
		logger.Warn("Raft store is not implemented for RID yet")
		return nil, nil
	default:
		return nil, stacktrace.NewError("Unsupported store type %q for rid", storeType)
	}
}
