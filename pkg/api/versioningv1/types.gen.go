// This file is auto-generated; do not change as any changes will be overwritten
package versioning

// Identifier of a system boundary, known to both the client and the USS separate from this API, for which this interface can provide a version.  While the format is not prescribed by this API, any value must be URL-safe.  It is recommended to use an approach similar to reverse-order Internet domain names and Java packages where the global scope is described with increasingly-precise identifiers joined by periods.  For instance, the system boundary containing the mandatory Network Identification U-space service might be identified with `gov.eu.uspace.v1.netid` because the authority defining this system boundary is a governmental organization (specifically, the European Union) with requirements imposed on the system under test by the U-space regulation (first version) -- specifically, the Network Identification Service section.
type SystemBoundaryIdentifier string

// Identifier of a particular version of a system (defined by a known system boundary).  While the format is not prescribed by this API, a semantic version (https://semver.org/) prefixed with a `v` is recommended.
type VersionIdentifier string

type GetVersionResponse struct {
	// The requested system identity/boundary.
	SystemIdentity *SystemBoundaryIdentifier `json:"system_identity,omitempty"`

	// The version of the system with the specified system identity/boundary.
	SystemVersion *VersionIdentifier `json:"system_version,omitempty"`
}
