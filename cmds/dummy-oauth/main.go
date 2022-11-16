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
	"github.com/interuss/dss/cmds/dummy-oauth/api"
	"github.com/interuss/dss/cmds/dummy-oauth/api/dummyoauth"
)

var (
	address = flag.String("addr", ":8085", "address")
	keyFile = flag.String("private_key_file", "build/test-certs/auth2.key", "OAuth private key file")
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

type PermissiveAuthorizer struct{}

func (*PermissiveAuthorizer) Authorize(w http.ResponseWriter, r *http.Request, authOptions []api.AuthorizationOption) api.AuthorizationResult {
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
