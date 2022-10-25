// This file is auto-generated; do not change as any changes will be overwritten
package auxv1

type VersionResponse struct {
	// The version of the DSS.
	Version string `json:"version"`
}

type ErrorResponse struct {
	// Human-readable message indicating what error occurred and/or why.
	Message *string `json:"message"`
}
