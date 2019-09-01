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
	"go.uber.org/zap"
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
	logger           *zap.Logger
	key              interface{}
	requiredScopes   map[string][]string
	requiredAudience string
}

// NewRSAAuthClient returns a new authClient instance which uses RSA.
func NewRSAAuthClient(keyFile string, logger *zap.Logger) (*authClient, error) {
	bytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	pub, _ := pem.Decode(bytes)
	if pub == nil {
		return nil, fmt.Errorf("error decoding keyFile")
	}
	parsedKey, err := x509.ParsePKIXPublicKey(pub.Bytes)
	key, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not create rsa public key from %s", keyFile)
	}
	return &authClient{logger: logger, key: key}, nil
}

func (a *authClient) RequireScopes(scopes map[string][]string) {
	a.requiredScopes = scopes
}

func (a *authClient) RequireAudience(audience string) {
	a.requiredAudience = audience
}

func (a *authClient) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	tknStr, ok := getToken(ctx)
	if !ok {
		return nil, dsserr.Unauthenticated("missing token")
	}

	claims := claims{}

	_, err := jwt.ParseWithClaims(tknStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return a.key, nil
	})
	if err != nil {
		a.logger.Error("token validation failed", zap.Error(err))
		return nil, dsserr.Unauthenticated("invalid token")
	}

	if a.requiredAudience != "" && claims.Audience != a.requiredAudience {
		return nil, dsserr.Unauthenticated(
			fmt.Sprintf("invalid token audience, expected %v, got %v", a.requiredAudience, claims.Audience))
	}

	if err := a.missingScopes(info, strings.Split(claims.ScopeString, " ")); err != nil {
		return nil, dsserr.PermissionDenied(fmt.Sprintf("missing scopes: %v", err))
	}

	return handler(ContextWithOwner(ctx, models.Owner(claims.Subject)), req)
}

// Returns all of the required scopes that are missing.
func (a *authClient) missingScopes(info *grpc.UnaryServerInfo, scopes []string) error {
	var (
		parts      = strings.Split(info.FullMethod, "/")
		method     = parts[len(parts)-1]
		claimedMap = make(map[string]bool)
		err        = &missingScopesError{}
	)

	for _, s := range scopes {
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
