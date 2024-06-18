package tokenSigner

import (
	"crypto/rsa"
	"os"
	"time"
	conf "token-signer/config"

	"github.com/golang-jwt/jwt"
)

type JwtTokenSigner struct {
	privateKey *rsa.PrivateKey
}

func New() JwtTokenSigner {

	config := conf.GetGlobalConfig()

	bytes, err := os.ReadFile(config.RsaPrivateKeyFileName)
	if err != nil {
		panic("No Private Key found")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		panic("Failed to parse Parse Key")
	}

	return JwtTokenSigner{
		privateKey: privateKey,
	}
}

func (j JwtTokenSigner) CreateSignedToken(request CreateSignedTokenRequest) (CreateSignedTokenResponse, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   request.Aud,
		"scope": request.Scope,
		"iss":   request.Iss,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"sub":   request.Sub,
	})

	tokenString, err := token.SignedString(j.privateKey)
	if err != nil {
		return CreateSignedTokenResponse{}, err
	}

	return CreateSignedTokenResponse{
		AccessToken: tokenString,
	}, nil

}
