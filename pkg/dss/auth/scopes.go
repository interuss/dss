package auth

// Operation models the name of an operation.
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

// MergeOperationsAndScopes merges a and be together.
func MergeOperationsAndScopes(requiredScopes ...map[Operation][]Scope) map[Operation][]Scope {
	if len(requiredScopes) == 0 {
		return nil
	}

	result := map[Operation][]Scope{}

	for _, rs := range requiredScopes {
		for k, v := range rs {
			result[k] = append(result[k], v...)
		}
	}

	return result
}
