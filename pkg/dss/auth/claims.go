package auth

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var (
	errMissingOrEmptySubject = errors.New("Missing or empty subject")
)

// scopes models a set of scopes.
type scopes map[string]struct{}

func (s *scopes) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*s = map[string]struct{}{}

	for _, scope := range strings.Split(str, " ") {
		(*s)[scope] = struct{}{}
	}

	return nil
}

type claims struct {
	jwt.StandardClaims
	Scopes scopes `json:"scope"`
}

func (c *claims) Valid() error {
	if c.Subject == "" {
		return errMissingOrEmptySubject
	}

	return c.StandardClaims.Valid()
}
