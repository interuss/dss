package repos

import (
	"context"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
)

// repos.Misc abstracts misc database helpers
type Misc interface {
	// GetDSSAirspaceRepresentationID gets the ID of the common DSS Airspace Representation the Datastore represents
	GetDSSAirspaceRepresentationID(ctx context.Context) (string, error)
}

// repos.DSSMetadata abstracts constraint-specific interactions with the dss metadata repository.
type DSSMetadata interface {
	// SaveOwnMetadata store our metadata into the pool participants
	SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error
	// GetDSSMetadata returns all DSS metadata of pool participants
	GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error)
}

// Repository aggregates all SCD-specific repo interfaces.
type Repository interface {
	DSSMetadata
	Misc
}
