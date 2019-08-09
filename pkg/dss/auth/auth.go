package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	// ContextKeyOwner is the key to an owner value.
	ContextKeyOwner ContextKey = "owner"
)

// ContextKey models auth-specific keys in a context.
type ContextKey string

type missingScopesError struct {
	s []string
}

func (m *missingScopesError) Error() string {
	return strings.Join(m.s, ", ")
}

// ContextWithOwner adds "owner" to "ctx".
func ContextWithOwner(ctx context.Context, owner models.Owner) context.Context {
	return context.WithValue(ctx, ContextKeyOwner, owner)
}

// OwnerFromContext returns the value for owner from "ctx" and a boolean
// indicating whether a valid value was present or not.
func OwnerFromContext(ctx context.Context) (models.Owner, bool) {
	owner, ok := ctx.Value(ContextKeyOwner).(models.Owner)
	return owner, ok
}

type authClient struct {
	key            interface{}
	requiredScopes map[string][]string
}

// NewSymmetricAuthClient returns a new authClient instance using symmetric keys.
func NewSymmetricAuthClient(keyFile string) (*authClient, error) {
	bytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return &authClient{key: bytes}, nil
}

// NewRSAAuthClient returns a new authClient instance which uses RSA.
func NewRSAAuthClient(keyFile string) (*authClient, error) {
	bytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	pub, _ := pem.Decode(bytes)
	parsedKey, err := x509.ParsePKIXPublicKey(pub.Bytes)
	key, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not create rsa public key from %s", keyFile)
	}
	return &authClient{key: key}, nil
}

func (a *authClient) RequireScopes(scopes map[string][]string) {
	a.requiredScopes = scopes
}

func (a *authClient) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	tknStr, ok := getToken(ctx)
	if !ok {
		return nil, dsserr.Unauthenticated("missing token")
	}

	claims := claims{}

	// TODO(steeling): Modify to ParseWithClaims and inspect claims.
	_, err := jwt.ParseWithClaims(tknStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return a.key, nil
	})
	if err != nil {
		return nil, dsserr.Unauthenticated("invalid token")
	}

	if err := a.missingScopes(info, claims); err != nil {
		return nil, dsserr.PermissionDenied(fmt.Sprintf("missing scopes: %v", err))
	}

	return handler(ContextWithOwner(ctx, models.Owner(claims.ClientID)), req)
}

// Returns all of the required scopes that are missing.
func (a *authClient) missingScopes(info *grpc.UnaryServerInfo, claims claims) error {
	var (
		parts      = strings.Split(info.FullMethod, "/")
		method     = parts[len(parts)-1]
		claimedMap = make(map[string]bool)
		err        = &missingScopesError{}
	)

	for _, s := range claims.Scopes {
		claimedMap[s] = true
	}
	for _, required := range a.requiredScopes[method] {
		if ok := claimedMap[required]; !ok {
			err.s = append(err.s, required)
		}
	}
	// Need to explicitly return nil
	if err.Error() == "" {
		return nil
	}
	return err
}

func getToken(ctx context.Context) (string, bool) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	authHeader := headers.Get("authorization")
	if len(authHeader) == 0 {
		return "", false
	}

	// Remove Bearer before returning.
	return strings.TrimPrefix(authHeader[0], "Bearer "), true
}
