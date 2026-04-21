package repos

import (
	"context"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
)

// aux_.repos.Misc abstracts misc data-backing helpers
type Misc interface {
	// GetDSSAirspaceRepresentationID gets the ID of the common DSS Airspace Representation the
	// data backing represents
	GetDSSAirspaceRepresentationID(ctx context.Context) (string, error)
}

// aux_.repos.DSSMetadata abstracts pool-information interactions with the DSS metadata repository.
type DSSMetadata interface {
	// SaveOwnMetadata store our metadata into the pool participants
	SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error
	// GetDSSMetadata returns all DSS metadata of pool participants
	GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error)
	// Record a new Timestamp
	RecordHeartbeat(ctx context.Context, hearthbeat auxmodels.Heartbeat) error
}

// aux_.repos.Repository aggregates all aux Repository (repo containing auxiliary information not
// related to standardized services like RID or SCD specifically) interfaces to perform aux
// operations on any data backing.  This is a repository type, generally intended to be
// obtained/used via a store.Store[Repository] interface.
type Repository interface {
	DSSMetadata
	Misc
}
