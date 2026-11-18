package repos

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
)

// Eviction is the repository for the eviction of a Raft group.
type Eviction interface {
	// DeleteExpired deletes the expired entities of every entity type according to params and returns their number per entity type.
	DeleteExpired(ctx context.Context, params dssmodels.EvictionParameters) (map[string]int, error)

	// GetEvictionState returns the eviction settings, or the default ones if they were never set, and the last cleanup job.
	GetEvictionState(ctx context.Context) (dssmodels.EvictionState, error)

	// SetEvictionConfig replaces the eviction settings with config and returns the stored settings.
	SetEvictionConfig(ctx context.Context, config dssmodels.EvictionConfig) (dssmodels.EvictionConfig, error)

	// SetEvictionDisabled disables or enables the scheduled cleanup jobs, keeping the eviction settings, and returns the eviction state.
	SetEvictionDisabled(ctx context.Context, disabled bool) (dssmodels.EvictionState, error)
}
