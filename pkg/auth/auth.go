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

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/models"

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
	ResolveKeys(context.Context) ([]interface{}, error)
}

type fromMemoryKeyResolver struct {
	Keys []interface{}
}

// ResolveKeys returns the set of keys provided to the fromMemoryKeyResolver.
func (r *fromMemoryKeyResolver) ResolveKeys(context.Context) ([]interface{}, error) {
	return r.Keys, nil
}

// FromFileKeyResolver resolves keys from 'KeyFile'.
type FromFileKeyResolver struct {
	KeyFiles []string
	keys     []interface{}
}

// ResolveKeys resolves an RSA public key from file for verifying JWTs.
func (r *FromFileKeyResolver) ResolveKeys(context.Context) ([]interface{}, error) {
	if r.keys != nil {
		return r.keys, nil
	}

	for _, f := range r.KeyFiles {
		bytes, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		pub, _ := pem.Decode(bytes)
		if pub == nil {
			return nil, errors.New("failed to decode keyFile")
		}
		parsedKey, err := x509.ParsePKIXPublicKey(pub.Bytes)
		if err != nil {
			return nil, err
		}
		key, ok := parsedKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("could not create rsa public key from %s", f)
		}
		r.keys = append(r.keys, key)
	}
	return r.keys, nil
}

// JWKSResolver resolves the key(s) with ID 'KeyID' from 'Endpoint' serving
// JWK sets.
type JWKSResolver struct {
	Endpoint *url.URL
	// If empty, will use all the keys provided by the jwks Endpoint.
	KeyIDs []string
}

// ResolveKeys resolves an RSA public key from file for verifying JWTs.
func (r *JWKSResolver) ResolveKeys(ctx context.Context) ([]interface{}, error) {
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

	var keys []interface{}
	var webKeys []jose.JSONWebKey
	if len(r.KeyIDs) == 0 {
		webKeys = jwks.Keys
	}
	for _, kid := range r.KeyIDs {
		// jwks.Key returns a slice of keys.
		jkeys := jwks.Key(kid)
		if len(jkeys) == 0 {
			return nil, fmt.Errorf("failed to resolve key(s) for ID: %s", kid)
		}
		webKeys = append(webKeys, jkeys...)
	}
	for _, w := range webKeys {
		keys = append(keys, w.Key)
	}
	return keys, nil
}

// KeyClaimedScopesValidator validates a set of scopes claimed by an incoming
// JWT.
type KeyClaimedScopesValidator interface {
	// ValidateKeyClaimedScopes returns an error if 'scopes' are not sufficient
	// to authorize an operation, nil otherwise.
	ValidateKeyClaimedScopes(ctx context.Context, scopes ScopeSet) error
}

type allScopesRequiredValidator struct {
	scopes []Scope
}

func (v *allScopesRequiredValidator) ValidateKeyClaimedScopes(ctx context.Context, scopes ScopeSet) error {
	var (
		missing []string
	)

	for _, scope := range v.scopes {
		if _, present := scopes[scope]; !present {
			missing = append(missing, scope.String())
		}
	}

	if len(missing) > 0 {
		return &missingScopesError{
			s: missing,
		}
	}

	return nil
}

// RequireAllScopes returns a KeyClaimedScopesValidator instance ensuring that
// every element in scopes is claimed by an incoming set of scopes.
func RequireAllScopes(scopes ...Scope) KeyClaimedScopesValidator {
	return &allScopesRequiredValidator{
		scopes: scopes,
	}
}

type anyScopesRequiredValidator struct {
	scopes []Scope
}

func (v *anyScopesRequiredValidator) ValidateKeyClaimedScopes(ctx context.Context, scopes ScopeSet) error {
	var (
		missing []string
	)

	for _, scope := range v.scopes {
		if _, present := scopes[scope]; present {
			return nil
		}
		missing = append(missing, scope.String())
	}

	return &missingScopesError{
		s: missing,
	}
}

// RequireAnyScope returns a KeyClaimedScopesValidator instance ensuring that
// at least one element in scopes is claimed by an incoming set of scopes.
func RequireAnyScope(scopes ...Scope) KeyClaimedScopesValidator {
	return &anyScopesRequiredValidator{
		scopes: scopes,
	}
}

// Authorizer authorizes incoming requests.
type Authorizer struct {
	logger            *zap.Logger
	keys              []interface{}
	keyGuard          sync.RWMutex
	scopesValidators  map[Operation]KeyClaimedScopesValidator
	acceptedAudiences map[string]bool
}

// Configuration bundles up creation-time parameters for an Authorizer instance.
type Configuration struct {
	KeyResolver       KeyResolver                             // Used to initialize and periodically refresh keys.
	KeyRefreshTimeout time.Duration                           // Keys are refreshed on this cadence.
	ScopesValidators  map[Operation]KeyClaimedScopesValidator // ScopesValidators are used to enforce authorization for operations.
	AcceptedAudiences []string                                // AcceptedAudiences enforces the aud keyClaim on the jwt. An empty string allows no aud keyClaim.
}

// NewRSAAuthorizer returns an Authorizer instance using values from configuration.
func NewRSAAuthorizer(ctx context.Context, configuration Configuration) (*Authorizer, error) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	keys, err := configuration.KeyResolver.ResolveKeys(ctx)
	if err != nil {
		return nil, err
	}

	auds := make(map[string]bool)
	for _, s := range configuration.AcceptedAudiences {
		auds[s] = true
	}

	authorizer := &Authorizer{
		scopesValidators:  configuration.ScopesValidators,
		acceptedAudiences: auds,
		logger:            logger,
		keys:              keys,
	}

	go func() {
		ticker := time.NewTicker(configuration.KeyRefreshTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				keys, err := configuration.KeyResolver.ResolveKeys(ctx)
				if err != nil {
					logger.Panic("failed to refresh key", zap.Error(err))
				}

				authorizer.setKeys(keys)
			case <-ctx.Done():
				logger.Warn("finalizing key refresh worker", zap.Error(ctx.Err()))
				return
			}
		}
	}()

	return authorizer, nil
}

func (a *Authorizer) setKeys(keys []interface{}) {
	a.keyGuard.Lock()
	a.keys = keys
	a.keyGuard.Unlock()
}

// AuthInterceptor intercepts incoming gRPC requests and extracts and verifies
// accompanying bearer tokens.
func (a *Authorizer) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	tknStr, ok := getToken(ctx)
	if !ok {
		return nil, dsserr.Unauthenticated("missing token")
	}

	a.keyGuard.RLock()
	keys := a.keys
	a.keyGuard.RUnlock()
	validated := false
	var err error
	var keyClaims claims

	for _, key := range keys {
		keyClaims = claims{}
		key := key
		_, err = jwt.ParseWithClaims(tknStr, &keyClaims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})
		if err == nil {
			validated = true
			break
		}
	}
	if !validated {
		a.logger.Error("token validation failed", zap.Error(err))
		return nil, dsserr.Unauthenticated(err.Error())
	}

	if !a.acceptedAudiences[keyClaims.Audience] {
		return nil, dsserr.Unauthenticated(
			fmt.Sprintf("invalid token audience: %v", keyClaims.Audience))
	}

	if err := a.validateKeyClaimedScopes(ctx, info, keyClaims.Scopes); err != nil {
		return nil, dsserr.PermissionDenied(fmt.Sprintf("missing scopes"))
	}

	return handler(ContextWithOwner(ctx, models.Owner(keyClaims.Subject)), req)
}

// Matches keyClaimedScopes against the required scopes and returns true if
// keyClaimedScopes contains at least one of the required scopes in a.
func (a *Authorizer) validateKeyClaimedScopes(ctx context.Context, info *grpc.UnaryServerInfo, keyClaimedScopes ScopeSet) error {
	if validator, known := a.scopesValidators[Operation(info.FullMethod)]; known {
		return validator.ValidateKeyClaimedScopes(ctx, keyClaimedScopes)
	}

	return nil
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
