package auth

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

var (
	errMissingOrEmptySubject = errors.New("Missing or empty subject")
)

type claims struct {
	jwt.StandardClaims
	ScopeString string `json:"scope"`
}

func (c *claims) Valid() error {
	if c.Subject == "" {
		return errMissingOrEmptySubject
	}

	return c.StandardClaims.Valid()
}
