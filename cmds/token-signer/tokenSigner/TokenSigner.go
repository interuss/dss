package tokenSigner

type CreateSignedTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type CreateSignedTokenRequest struct {
	Aud   string
	Scope string
	Iss   string
	Sub   string
	Exp   int64
}

type TokenSigner interface {
	CreateSignedToken(request CreateSignedTokenRequest) (CreateSignedTokenResponse, error)
}
