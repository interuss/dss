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
//
// Implementations do not validate their arguments: callers are responsible for ensuring their correctness.
type DSSMetadata interface {
	// SaveOwnMetadata stores our metadata into the pool participants.
	// locality and publicEndpoint must both be non-empty.
	SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error
	// GetDSSMetadata returns all DSS metadata of pool participants
	GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error)
	// RecordHeartbeat records a new heartbeat.
	// hearthbeat.Locality and hearthbeat.Source must both be non-empty
	// hearthbeat.Timestamp must be set
	// if hearthbeat.NextHeartbeatExpectedBefore is set, it must not be before
	// hearthbeat.Timestamp.
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
