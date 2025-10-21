package auth

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/interuss/stacktrace"
)

var (
	errMissingOrEmptySubject = errors.New("missing or empty subject")
	errTokenExpireTooFar     = errors.New("token expiration time is too far in the furture, Max token duration is 1 Hour")
	errMissingIssuer         = errors.New("missing Issuer URI")
	// Now allows test to override with specific time values
	Now = time.Now
)

// ScopeSet models a set of scopes.
type ScopeSet map[string]struct{}

func (s *ScopeSet) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return stacktrace.Propagate(err, "Unable to unmarshal JSON")
	}

	*s = map[string]struct{}{}

	for _, scope := range strings.Split(str, " ") {
		(*s)[scope] = struct{}{}
	}

	return nil
}

// ValidateRequiredScopes validates this ScopeSet against a set of required scopes.
// If some are missing it returns false and the missing scopes. If the validation passes it returns true.
func (s *ScopeSet) ValidateRequiredScopes(reqScopes []string) (bool, []string) {
	var missing []string
	for _, reqScope := range reqScopes {
		if _, present := (*s)[reqScope]; !present {
			missing = append(missing, reqScope)
		}
	}

	if len(missing) > 0 {
		return false, missing
	}

	return true, nil
}

func (s *ScopeSet) ToStringSlice() []string {
	scopes := make([]string, 0, len(*s))
	for scope := range *s {
		scopes = append(scopes, scope)
	}
	return scopes
}

type claims struct {
	jwt.StandardClaims
	Scopes ScopeSet `json:"scope"`
}

func (c *claims) Valid() error {
	if c.Subject == "" {
		return errMissingOrEmptySubject
	}
	now := Now()

	c.VerifyExpiresAt(now.Unix(), true)

	if c.ExpiresAt > now.Add(time.Hour).Unix() {
		return errTokenExpireTooFar
	}

	if c.Issuer == "" {
		return errMissingIssuer
	}

	return c.StandardClaims.Valid()
}
