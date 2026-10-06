package store

import (
	"context"

	"github.com/interuss/dss/pkg/rid/repos"
	ridmemstore "github.com/interuss/dss/pkg/rid/store/memstore"
	ridraftstore "github.com/interuss/dss/pkg/rid/store/raftstore"
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
	case params.SQLStoreType:
		return ridsqlstore.Init(ctx, logger, withCheckCron)
	case params.RaftStoreType:
		return ridraftstore.Init(ctx, logger)
	case params.MemStoreType:
		return ridmemstore.Init(ctx, logger)
	default:
		return nil, stacktrace.NewError("Unsupported store type %q for rid", storeType)
	}
}
