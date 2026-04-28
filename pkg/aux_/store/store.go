package store

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	auxdatastore "github.com/interuss/dss/pkg/aux_/store/datastore"
	dssstore "github.com/interuss/dss/pkg/store"
	"go.uber.org/zap"
)

// aux_.store.Store is a generic means to obtain an aux Repository (repo containing auxiliary
// information not related to standardized services like RID or SCD specifically) to perform
// aux-specific operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]

// Init selects and initializes the aux store backend.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool) (Store, error) {
	return auxdatastore.Init(ctx, logger, withCheckCron)
}
