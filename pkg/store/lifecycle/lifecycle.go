package lifecycle

import (
	"context"

	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/store/params"
	"go.uber.org/zap"
)

// Start starts the resources of the selected store that live as long as the core-service process, independently of the status of the stores using them.
// It is called before any store is initialized.
// The resources are released by Stop.
// A failure of these resources after they started is reported to onError.
func Start(ctx context.Context, logger *zap.Logger, onError func(error)) error {
	switch params.GetStoreParameters().StoreType {
	case params.RaftStoreType:
		return raftstore.StartListener(ctx, logger, onError)
	default:
		return nil
	}
}

// Stop releases the resources started by Start.
func Stop(ctx context.Context) error {
	switch params.GetStoreParameters().StoreType {
	case params.RaftStoreType:
		return raftstore.StopListener(ctx)
	default:
		return nil
	}
}
