package auth

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

var (
	errMissingOrEmptyClientID = errors.New("Missing or empty client_id")
)

type claims struct {
	jwt.StandardClaims
	ClientID    string `json:"client_id"`
	ScopeString string `json:"scope"`
}

func (c *claims) Valid() error {
	if c.ClientID == "" {
		return errMissingOrEmptyClientID
	}

	return c.StandardClaims.Valid()
}
