package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"
	"github.com/steeling/InterUSS-Platform/pkg/logging"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gopkg.in/square/go-jose.v2"
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

// KeyResolver abstracts resolving keys.
type KeyResolver interface {
	// ResolveKey returns a public or private key, most commonly an rsa.PublicKey.
	ResolveKey(context.Context) (interface{}, error)
}

type fromMemoryKeyResolver struct {
	Key interface{}
}

func (r *fromMemoryKeyResolver) ResolveKey(context.Context) (interface{}, error) {
	return r.Key, nil
}

// FromFileKeyResolver resolves keys from 'KeyFile'.
type FromFileKeyResolver struct {
	KeyFile string
	key     interface{}
}

// ResolveKey resolves an RSA public key from file for verifying JWTs.
func (r *FromFileKeyResolver) ResolveKey(context.Context) (interface{}, error) {
	if r.key != nil {
		return r.key, nil
	}

	bytes, err := ioutil.ReadFile(r.KeyFile)
	if err != nil {
		return nil, err
	}
	pub, _ := pem.Decode(bytes)
	if pub == nil {
		return nil, errors.New("failed to decode keyFile")
	}
	parsedKey, err := x509.ParsePKIXPublicKey(pub.Bytes)
	key, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not create rsa public key from %s", r.KeyFile)
	}

	r.key = key
	return r.key, nil
}

// JWKSResolver resolves the key with ID 'KeyID' from 'Endpoint' serving JWK sets.
type JWKSResolver struct {
	Endpoint *url.URL
	KeyID    string
}

// ResolveKey resolves an RSA public key from file for verifying JWTs.
func (r *JWKSResolver) ResolveKey(ctx context.Context) (interface{}, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL:    r.Endpoint,
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jwks := jose.JSONWebKeySet{}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	keys := jwks.Key(r.KeyID)
	if len(keys) == 0 {
		return nil, fmt.Errorf("failed to resolve key for ID: %s", r.KeyID)
	}

	key := keys[0].Key.(*rsa.PublicKey)
	if key == nil {
		return nil, fmt.Errorf("failed to resolve key for ID: %s", r.KeyID)
	}

	return key, nil
}

// Authorizer authorizes incoming requests.
type Authorizer struct {
	logger           *zap.Logger
	keyGuard         sync.RWMutex
	key              interface{}
	requiredScopes   map[string][]string
	requiredAudience string
}

// Configuration bundles up creation-time parameters for an Authorizer instance.
type Configuration struct {
	KeyResolver       KeyResolver         // Used to initialize and periodically refresh keys.
	KeyRefreshTimeout time.Duration       // Keys are refreshed on this cadence.
	RequiredScopes    map[string][]string // RequiredScopes are enforced if not nil.
	RequiredAudience  string              // RequiredAudience is enforced if not empty.
}

// NewRSAAuthorizer returns an Authorizer instance using values from configuration.
func NewRSAAuthorizer(ctx context.Context, configuration Configuration) (*Authorizer, error) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	key, err := configuration.KeyResolver.ResolveKey(ctx)
	if err != nil {
		return nil, err
	}

	authorizer := &Authorizer{
		requiredScopes:   configuration.RequiredScopes,
		requiredAudience: configuration.RequiredAudience,
		logger:           logger,
		key:              key,
	}

	go func() {
		ticker := time.NewTicker(configuration.KeyRefreshTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				key, err := configuration.KeyResolver.ResolveKey(ctx)
				if err != nil {
					logger.Panic("failed to refresh key", zap.Error(err))
				}

				authorizer.setKey(key)
			case <-ctx.Done():
				logger.Warn("finalizing key refresh worker", zap.Error(ctx.Err()))
				return
			}
		}
	}()

	return authorizer, nil
}

func (a *Authorizer) setKey(key interface{}) {
	a.keyGuard.Lock()
	defer a.keyGuard.Unlock()

	a.key = key
}

// AuthInterceptor intercepts incoming gRPC requests and extracts and verifies
// accompanying bearer tokens.
func (a *Authorizer) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	tknStr, ok := getToken(ctx)
	if !ok {
		return nil, dsserr.Unauthenticated("missing token")
	}

	claims := claims{}

	a.keyGuard.RLock()
	_, err := jwt.ParseWithClaims(tknStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return a.key, nil
	})
	a.keyGuard.RUnlock()

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
func (a *Authorizer) missingScopes(info *grpc.UnaryServerInfo, scopes []string) error {
	var (
		parts      = strings.Split(info.FullMethod, "/")
		method     = parts[len(parts)-1]
		claimedMap = make(map[string]bool)
		err        = &missingScopesError{}
	)

	for _, s := range scopes {
		claimedMap[s] = true
	}

	if a.requiredScopes != nil {
		for _, required := range a.requiredScopes[method] {
			if ok := claimedMap[required]; !ok {
				err.s = append(err.s, required)
			}
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
