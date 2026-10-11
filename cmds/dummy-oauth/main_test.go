package main_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"

	main "github.com/interuss/dss/cmds/dummy-oauth"
	"github.com/stretchr/testify/require"
)

func TestGetJWKS(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	dummy := &main.DummyOAuthImplementation{
		KeyID:      "the-key-id",
		PrivateKey: key,
	}
	response := dummy.GetJWKS(t.Context(), nil)
	require.Nil(t, response.Response500)
	require.Len(t, response.Response200.Keys, 1)

	jwk := response.Response200.Keys[0]
	require.Equal(t, "RSA", jwk.Kty)
	require.Equal(t, "sig", jwk.Use)
	require.Equal(t, "RS256", jwk.Alg)
	require.Equal(t, "the-key-id", jwk.Kid)
	require.Equal(t, base64Encode(key.N), jwk.N)
	require.Equal(t, base64Encode(big.NewInt(int64(key.E))), jwk.E)
}

func base64Encode(data *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(data.Bytes())
}
