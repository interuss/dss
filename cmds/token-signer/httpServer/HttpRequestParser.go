package httpServer

import (
	"errors"
	"net/http"
	"time"
	conf "token-signer/config"
	"token-signer/tokenSigner"
)

type HttpRequestParser struct {
}

func (v HttpRequestParser) ParseRequest(r *http.Request) (tokenSigner.CreateSignedTokenRequest, error) {

	config := conf.GetGlobalConfig()
	queryStrings := r.URL.Query()

	request := tokenSigner.CreateSignedTokenRequest{}

	if queryStrings.Has("aud") {
		request.Aud = queryStrings.Get("aud")
	} else {
		return request, errors.New("Missing `aud` query parameter")
	}

	if queryStrings.Has("scope") {
		request.Scope = queryStrings.Get("scope")
	} else {
		return request, errors.New("Missing `scope` query parameter")
	}

	if user := r.Header.Get("X-Consumer-Username"); user != "" {
		request.Sub = user
	} else {
		return request, errors.New("Missing authorization header")
	}

	request.Exp = time.Now().Add(time.Hour).Unix()
	request.Iss = config.IssuerName

	return request, nil
}
