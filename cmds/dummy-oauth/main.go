package main

import (
	"context"
	"crypto/rsa"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/interuss/dss/cmds/dummy-oauth/api"
	"github.com/interuss/dss/cmds/dummy-oauth/api/dummyoauth"
)

var (
	address = flag.String("addr", ":8085", "address")
	keyFile = flag.String("private_key_file", "../../build/test-certs/auth2.key", "OAuth private key file")
	jwksURI = flag.String("jwks_uri", "http://host.docker.internal:8085/.well-known/jwks.json", "JWKS URI")
)

type DummyOAuthImplementation struct {
	PrivateKey *rsa.PrivateKey
}

func (s *DummyOAuthImplementation) GetToken(ctx context.Context, req *dummyoauth.GetTokenRequest) dummyoauth.GetTokenResponseSet {
	resp := dummyoauth.GetTokenResponseSet{}

	var intendedAudience string
	if req.IntendedAudience != nil {
		intendedAudience = *req.IntendedAudience
	} else {
		msg := "Missing `intended_audience` query parameter"
		resp.Response400 = &dummyoauth.BadRequestResponse{Message: &msg}
		return resp
	}

	var scope string
	if req.Scope != nil {
		scope = *req.Scope
	} else {
		msg := "Missing `scope` query parameter"
		resp.Response400 = &dummyoauth.BadRequestResponse{Message: &msg}
		return resp
	}

	var issuer string
	if req.Issuer != nil {
		issuer = *req.Issuer
	} else {
		issuer = "dummyoauth"
	}

	var expireTime int64
	if req.Expire == nil {
		expireTime = time.Now().Add(time.Hour).Unix()
	} else {
		expireTime = int64(*req.Expire)
	}

	var sub string
	if req.Sub != nil {
		sub = *req.Sub
	} else {
		sub = "fake_uss"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   intendedAudience,
		"scope": scope,
		"iss":   issuer,
		"exp":   expireTime,
		"sub":   sub,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(s.PrivateKey)
	if err != nil {
		resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: err.Error()}
		return resp
	}

	resp.Response200 = &dummyoauth.TokenResponse{AccessToken: tokenString}
	return resp
}

func (s *DummyOAuthImplementation) PostFimsToken(ctx context.Context, req *dummyoauth.PostFimsTokenRequest) dummyoauth.PostFimsTokenResponseSet {
	resp := dummyoauth.PostFimsTokenResponseSet{}

	// Note - validation for message signed headers in token request can be added here
	var body dummyoauth.TokenRequestForm
	if req.Body != nil {
		body = *req.Body
	} else {
		e := "Missing request `body`"
		eDisc := "Body is required with grant_type, client_id, scope, audience, current_timestamp"
		resp.Response400 = &dummyoauth.HTTPErrorResponse{Error: &e, ErrorDescription: &eDisc}
		return resp
	}

	var scope string = body.Scope
	if &scope == nil {
		e := "Missing scope in request `body`"
		eDisc := "Body is required with grant_type, client_id, scope, audience, current_timestamp"
		resp.Response400 = &dummyoauth.HTTPErrorResponse{Error: &e, ErrorDescription: &eDisc}
		return resp
	}

	var sub string = body.ClientID
	if &scope == nil {
		e := "Missing clientId in request `body`"
		eDisc := "Body is required with grant_type, client_id, scope, audience, current_timestamp"
		resp.Response400 = &dummyoauth.HTTPErrorResponse{Error: &e, ErrorDescription: &eDisc}
		return resp
	}

	var aud string = body.Audience
	if &aud == nil {
		log.Print("Missing audience in requst body, setting it to no-aud")
		aud = "no-aud"
	}

	var expireTime int64 = time.Now().Add(time.Hour).Unix()

	var nbf int64 = time.Now().Unix()

	var issuer string = "dummy.auth"
	var tokenType string = "bearer"
	jti := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"token_type": tokenType,
		"aud":        aud,
		"scope":      scope,
		"iss":        issuer,
		"expires_in": expireTime,
		"sub":        sub,
		"nbf":        nbf,
		"jti":        jti,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(s.PrivateKey)
	if err != nil {
		resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: err.Error()}
		return resp
	}

	resp.Response200 = &dummyoauth.HTTPTokenResponse{AccessToken: &tokenString, Scope: &scope,
		TokenType: &tokenType, ExpiresIn: &expireTime, Nbf: &nbf, Sub: &sub, Jti: &jti, Aud: &aud}
	return resp
}

func (s *DummyOAuthImplementation) GetFimsWellKnownOauthAuthorizationServer(ctx context.Context, req *dummyoauth.GetFimsWellKnownOauthAuthorizationServerRequest) dummyoauth.GetFimsWellKnownOauthAuthorizationServerResponseSet {
	response := dummyoauth.GetFimsWellKnownOauthAuthorizationServerResponseSet{}

	response.Response200 = &dummyoauth.Metadata{JwksURI: *jwksURI}
	return response
}

func (s *DummyOAuthImplementation) GetFimsWellKnownJwksJSON(ctx context.Context, req *dummyoauth.GetFimsWellKnownJwksJSONRequest) dummyoauth.GetFimsWellKnownJwksJSONResponseSet {
	response := dummyoauth.GetFimsWellKnownJwksJSONResponseSet{}

	var jwkey dummyoauth.JSONWebKey = *new(dummyoauth.JSONWebKey)
	e := "AQAB"
	n := "eQ22nLcYHRhMKXZUIJ3baLSsnAgYFJrMPhBEq8fqtyHQg_iKBv7Tavu3Rf_-26PRVvC0nPdwQgI_w4ZKqt1NIIaPljTc5raA-TH_RzRXwPR5JdL8JQLSqtgecAYuqSjt5bzsdbSuHueeXZsHgu75Hx86ZC3l-sInl5OTPArlhzM"
	kid := "cadd2909-8638-4b2d-8e47-2d9816fe360e"

	// JWK for auth2.pem
	jwkey.E = &e
	jwkey.N = &n
	jwkey.Kty = "RSA"
	jwkey.Kid = &kid

	// Read private key - Following not working. Need to try more
	// josejwk, errorjwk := jose.GenerateJWKFromPEM("../../build/test-certs/auth2.pem", false)
	// if errorjwk != nil {
	// 	log.Printf("Error while generating Jwk form PEM - %s", errorjwk)
	// }
	// jwkey.Alg = &josejwk.Algorithm

	var arr = []dummyoauth.JSONWebKey{jwkey}
	response.Response200 = &dummyoauth.JSONWebKeySet{Keys: &arr}
	return response
}

type PermissiveAuthorizer struct{}

func (*PermissiveAuthorizer) Authorize(w http.ResponseWriter, r *http.Request, schemes *map[string]api.SecurityScheme) api.AuthorizationResult {
	return api.AuthorizationResult{}
}

func main() {
	flag.Parse()

	// Read private key
	bytes, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		log.Panic(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		log.Panic(err)
	}

	// Define and start HTTP server
	impl := DummyOAuthImplementation{PrivateKey: privateKey}
	router := dummyoauth.MakeAPIRouter(&impl, &PermissiveAuthorizer{})
	multiRouter := api.MultiRouter{Routers: []api.PartialRouter{&router}}
	s := &http.Server{
		Addr:    *address,
		Handler: &multiRouter,
	}
	log.Fatal(s.ListenAndServe())
}
