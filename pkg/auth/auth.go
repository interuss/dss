package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/interuss/dss/pkg/api"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// KeyResolver abstracts resolving keys.
type KeyResolver interface {
	// ResolveKeys returns a public or private key, most commonly an rsa.PublicKey.
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
		bytes, err := os.ReadFile(f)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error reading key file")
		}
		pub, _ := pem.Decode(bytes)
		if pub == nil {
			return nil, stacktrace.NewError("Failed to decode key file")
		}
		parsedKey, err := x509.ParsePKIXPublicKey(pub.Bytes)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error parsing key as x509 public key")
		}
		key, ok := parsedKey.(*rsa.PublicKey)
		if !ok {
			return nil, stacktrace.NewError("Could not create RSA public key from %s", f)
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
		return nil, stacktrace.Propagate(err, "Error retrieving JWKS at %s", req.URL)
	}
	defer resp.Body.Close()

	jwks := jose.JSONWebKeySet{}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, stacktrace.Propagate(err, "Error decoding JWKS")
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
			return nil, stacktrace.NewError("Failed to resolve key(s) for ID: %s", kid)
		}
		webKeys = append(webKeys, jkeys...)
	}
	for _, w := range webKeys {
		keys = append(keys, w.Key)
	}
	return keys, nil
}

// Authorizer authorizes incoming requests.
type Authorizer struct {
	logger   *zap.Logger
	keys     []interface{}
	keyGuard sync.RWMutex

	AcceptedAudiences map[string]bool
}

// Configuration bundles up creation-time parameters for an Authorizer instance.
type Configuration struct {
	KeyResolver       KeyResolver   // Used to initialize and periodically refresh keys.
	KeyRefreshTimeout time.Duration // Keys are refreshed on this cadence.
	AcceptedAudiences []string      // AcceptedAudiences enforces the aud keyClaim on the jwt. An empty string allows no aud keyClaim.
}

// NewRSAAuthorizer returns an Authorizer instance using values from configuration.
func NewRSAAuthorizer(ctx context.Context, configuration Configuration) (*Authorizer, error) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	keys, err := configuration.KeyResolver.ResolveKeys(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to resolve keys")
	}

	auds := make(map[string]bool)
	for _, s := range configuration.AcceptedAudiences {
		auds[s] = true
	}

	authorizer := &Authorizer{
		AcceptedAudiences: auds,
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

type CtxAuthKey struct{}
type CtxAuthValue struct {
	Claims Claims
	Error  error
}

// Authorize extracts and verifies bearer tokens from a http.Request.
func (a *Authorizer) Authorize(_ http.ResponseWriter, r *http.Request, authOptions []api.AuthorizationOption) api.AuthorizationResult {
	v := r.Context().Value(CtxAuthKey{}).(CtxAuthValue)
	if v.Error != nil {
		return api.AuthorizationResult{Error: stacktrace.PropagateWithCode(v.Error, dsserr.Unauthenticated, "Failed to extract claims from access token")}
	}

	if pass, missing := validateScopes(authOptions, v.Claims.Scopes); !pass {
		return api.AuthorizationResult{Error: stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"Access token missing scopes (%v) while expecting %v and got %v",
			missing, describeAuthorizationExpectations(authOptions), strings.Join(v.Claims.Scopes.ToStringSlice(), ", "))}
	}

	return api.AuthorizationResult{
		ClientID: &v.Claims.Subject,
		Scopes:   v.Claims.Scopes.ToStringSlice(),
	}
}

func (a *Authorizer) ExtractClaims(r *http.Request) (Claims, error) {
	tknStr, ok := getToken(r)
	if !ok {
		return Claims{}, stacktrace.NewErrorWithCode(dsserr.Unauthenticated, "Missing access token")
	}

	a.keyGuard.RLock()
	keys := a.keys
	a.keyGuard.RUnlock()
	validated := false
	var err error
	var keyClaims Claims

	for _, key := range keys {
		keyClaims = Claims{}
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
		if err == nil { // If we have no keys, errs may be nil
			err = stacktrace.NewErrorWithCode(dsserr.Unauthenticated, "No keys to validate against")
		}
		return Claims{}, stacktrace.PropagateWithCode(err, dsserr.Unauthenticated, "Access token validation failed")
	}

	return keyClaims, nil
}

func HasScope(scopes []string, requiredScope api.RequiredScope) bool {
	for _, scope := range scopes {
		if scope == string(requiredScope) {
			return true
		}
	}
	return false
}

// describeAuthorizationExpectations builds a human-readable string describing the expectations of the authorization options.
func describeAuthorizationExpectations(authOptions []api.AuthorizationOption) string {
	if len(authOptions) == 0 {
		return "no expectation"
	}

	var expectations []string
	for _, authOption := range authOptions {
		var authOptionExpectations []string
		for scheme, scopes := range authOption {

			scopesStr := make([]string, len(scopes))
			for scopeIdx, scope := range scopes {
				scopesStr[scopeIdx] = string(scope)
			}

			schemeExpectation := fmt.Sprintf("%s: (%s)", scheme, strings.Join(scopesStr, " AND "))
			authOptionExpectations = append(authOptionExpectations, schemeExpectation)
		}

		authOptionExpectation := fmt.Sprintf("[%s]", strings.Join(authOptionExpectations, " AND "))
		expectations = append(expectations, authOptionExpectation)
	}
	return strings.Join(expectations, " OR ")
}

// validateScopes matches scopes against a set of authorization options. Validation against a single one of those is
// enough for the validation to succeed. Returns true if it succeeds, or returns false and a string describing the
// missing scopes if it fails. Empty authorization options means that the validation passes.
func validateScopes(authOptions []api.AuthorizationOption, clientScopes ScopeSet) (bool, string) {
	if len(authOptions) == 0 {
		return true, ""
	}

	var validationFailures []string
	for authOptionIdx, authOption := range authOptions {
		var reqScopes []string
		for _, scopes := range authOption {
			scopesStr := make([]string, len(scopes))
			for scopeIdx, scope := range scopes {
				scopesStr[scopeIdx] = string(scope)
			}
			reqScopes = append(reqScopes, scopesStr...)
		}

		if pass, missing := clientScopes.ValidateRequiredScopes(reqScopes); !pass {
			validationFailures = append(validationFailures,
				fmt.Sprintf("AuthOption[%d]: %v", authOptionIdx, strings.Join(missing, ", ")))
		} else {
			return true, ""
		}
	}

	return false, strings.Join(validationFailures, " ; ")
}

func getToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("authorization")
	if len(authHeader) < 7 || strings.ToLower(authHeader[0:6]) != "bearer" {
		return "", false
	}
	return authHeader[7:], true
}
