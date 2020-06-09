package auth

// Operation models the name of an operation.
//
// In the case of gRPC, the operation should be fully scoped, i.e.:
//   /{package-qualified service name}/{handler name}
// For example:
//   /ridpb.DiscoveryAndSynchronizationService/CreateIdentificationServiceArea
type Operation string

// String returns the string representation of o.
func (o Operation) String() string {
	return string(o)
}

// Scope models an oauth scope.
type Scope string

// String returns the string representation of s.
func (s Scope) String() string {
	return string(s)
}

// MergeOperationsAndScopesValidators merges scopesValidators.
func MergeOperationsAndScopesValidators(scopesValidators ...map[Operation]KeyClaimedScopesValidator) map[Operation]KeyClaimedScopesValidator {
	result := map[Operation]KeyClaimedScopesValidator{}

	if len(scopesValidators) == 0 {
		return result
	}

	for _, rs := range scopesValidators {
		for k, v := range rs {
			result[k] = v
		}
	}

	return result
}
