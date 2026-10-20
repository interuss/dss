package main_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	main "github.com/interuss/dss/cmds/dummy-oauth"
	"github.com/interuss/dss/cmds/dummy-oauth/api"
	"github.com/interuss/dss/cmds/dummy-oauth/api/dummyoauth"
	"github.com/stretchr/testify/require"
)

func newDummyOAuth(t *testing.T) *main.DummyOAuthImplementation {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return &main.DummyOAuthImplementation{
		KeyID:      "the-key-id",
		PrivateKey: key,
	}
}

func newGetTokenRequest() *dummyoauth.GetTokenRequest {
	return &dummyoauth.GetTokenRequest{
		IntendedAudience: new("the-audience"),
		Scope:            new("the-scope"),
		Issuer:           new("the-issuer"),
		Expire:           new(time.Now().Add(time.Minute).Unix()),
		Sub:              new("the-subject"),
	}
}

type dummyClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope"`
}

func parseToken(
	t *testing.T,
	dummy *main.DummyOAuthImplementation,
	token string,
) (*jwt.Token, dummyClaims) {
	t.Helper()
	var claims dummyClaims
	keyFunc := func(*jwt.Token) (any, error) {
		return &dummy.PrivateKey.PublicKey, nil
	}
	parsed, err := jwt.ParseWithClaims(token, &claims, keyFunc, jwt.WithValidMethods([]string{
		jwt.SigningMethodRS256.Alg(),
	}))
	require.NoError(t, err)
	return parsed, claims
}

func testBadRequest(t *testing.T, expectedMessage string, request *dummyoauth.GetTokenRequest) {
	t.Helper()
	dummy := newDummyOAuth(t)
	response := dummy.GetToken(t.Context(), request)
	require.Nil(t, response.Response200)
	require.Nil(t, response.Response500)
	require.Equal(t, expectedMessage, *response.Response400.Message)
}

func TestGetTokenMissingIntendedAudience(t *testing.T) {
	request := newGetTokenRequest()
	request.IntendedAudience = nil
	testBadRequest(t, "Missing `intended_audience` query parameter", request)
}

func TestGetTokenMissingScope(t *testing.T) {
	request := newGetTokenRequest()
	request.Scope = nil
	testBadRequest(t, "Missing `scope` query parameter", request)
}

func TestGetTokenFailsToSign(t *testing.T) {
	request := newGetTokenRequest()
	dummy := newDummyOAuth(t)
	dummy.PrivateKey = &rsa.PrivateKey{}
	response := dummy.GetToken(t.Context(), request)
	require.Nil(t, response.Response200)
	require.Nil(t, response.Response400)

	_, err := jwt.SigningMethodRS256.Sign("", dummy.PrivateKey)
	require.Equal(t, err.Error(), response.Response500.ErrorMessage)
}

func TestGetTokenSuccessWithExplicitParameters(t *testing.T) {
	request := newGetTokenRequest()
	dummy := newDummyOAuth(t)
	response := dummy.GetToken(t.Context(), request)

	require.Nil(t, response.Response500)
	require.Nil(t, response.Response400)

	token, claims := parseToken(t, dummy, response.Response200.AccessToken)
	require.Equal(t, "the-key-id", token.Header["kid"])
	require.EqualValues(t, []string{"the-audience"}, claims.Audience)
	require.Equal(t, "the-issuer", claims.Issuer)
	require.Equal(t, "the-subject", claims.Subject)
	require.Equal(t, "the-scope", claims.Scope)
	require.Equal(t, *request.Expire, claims.ExpiresAt.Unix())
}

func TestGetTokenSuccessWithDefaults(t *testing.T) {
	request := newGetTokenRequest()
	request.Issuer = nil
	request.Expire = nil
	request.Sub = nil
	dummy := newDummyOAuth(t)

	timeBefore := time.Now().Add(time.Hour).Unix()
	response := dummy.GetToken(t.Context(), request)
	timeAfter := time.Now().Add(time.Hour).Unix()

	require.Nil(t, response.Response500)
	require.Nil(t, response.Response400)
	token, claims := parseToken(t, dummy, response.Response200.AccessToken)
	require.Equal(t, "the-key-id", token.Header["kid"])
	require.EqualValues(t, []string{"the-audience"}, claims.Audience)
	require.Equal(t, "dummyoauth", claims.Issuer)
	require.Equal(t, "fake_uss", claims.Subject)
	require.GreaterOrEqual(t, claims.ExpiresAt.Unix(), timeBefore)
	require.LessOrEqual(t, claims.ExpiresAt.Unix(), timeAfter)
}

func TestGetJWKS(t *testing.T) {
	dummy := newDummyOAuth(t)
	response := dummy.GetJWKS(t.Context(), nil)
	require.Nil(t, response.Response500)
	require.Len(t, response.Response200.Keys, 1)

	jwk := response.Response200.Keys[0]
	require.Equal(t, "RSA", jwk.Kty)
	require.Equal(t, "sig", jwk.Use)
	require.Equal(t, "RS256", jwk.Alg)
	require.Equal(t, "the-key-id", jwk.Kid)
	require.Equal(t, base64Encode(dummy.PrivateKey.N), jwk.N)
	require.Equal(t, base64Encode(big.NewInt(int64(dummy.PrivateKey.E))), jwk.E)
}

func base64Encode(data *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(data.Bytes())
}

func TestPermissiveAuthorizerAlwaysSucceeds(t *testing.T) {
	authorizer := main.PermissiveAuthorizer{}
	require.Equal(t, api.AuthorizationResult{}, authorizer.Authorize(nil, nil, nil))
}
