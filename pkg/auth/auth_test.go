package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/api"
	"github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/auth/claims"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func rsaTokenReq(key *rsa.PrivateKey, exp, nbf int64) *http.Request {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"foo": "bar",
		"exp": exp,
		"nbf": nbf,
		"sub": "real_owner",
		"iss": "baz",
		"aud": "test-aud",
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	req := &http.Request{Header: make(http.Header)}
	req.Header.Set("Authorization", "Bearer "+tokenString)
	return req
}
func rsaTokenReqWithMissingIssuer(key *rsa.PrivateKey, exp, nbf int64) *http.Request {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"foo": "bar",
		"exp": exp,
		"nbf": nbf,
		"sub": "real_owner",
		"aud": "test-aud",
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	req := &http.Request{Header: make(http.Header)}
	req.Header.Set("Authorization", "Bearer "+tokenString)
	return req
}

func rsaTokenReqWithMissingAudience(key *rsa.PrivateKey, exp, nbf int64) *http.Request {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"foo": "bar",
		"exp": exp,
		"nbf": nbf,
		"sub": "real_owner",
		"iss": "baz",
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	req := &http.Request{Header: make(http.Header)}
	req.Header.Set("Authorization", "Bearer "+tokenString)
	return req
}

func rsaTokenReqWithMultipleAudience(key *rsa.PrivateKey, exp, nbf int64) *http.Request {
	return rsaTokenReqWithAudiences(key, exp, nbf, []string{"test-aud", "test-aud2"})
}

func rsaTokenReqWithAudiences(key *rsa.PrivateKey, exp, nbf int64, audiences interface{}) *http.Request {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"foo": "bar",
		"exp": exp,
		"nbf": nbf,
		"sub": "real_owner",
		"iss": "baz",
		"aud": audiences,
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	req := &http.Request{Header: make(http.Header)}
	req.Header.Set("Authorization", "Bearer "+tokenString)
	return req
}

func TestNewRSAAuthClient(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	tmpfile, err := os.CreateTemp(t.TempDir(), "bad.pem")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())
	// Test catches previous segfault.
	_, err = NewRSAAuthorizer(ctx, Configuration{
		KeyResolver: &FromFileKeyResolver{
			KeyFiles: []string{tmpfile.Name()},
		},
		KeyRefreshInterval: 1 * time.Millisecond,
	})
	require.Error(t, err)
	require.NoError(t, os.Remove(tmpfile.Name()))
}

func TestRSAAuthInterceptor(t *testing.T) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(42, 0)
	}

	defer func() {
		jwt.TimeFunc = time.Now
	}()

	noHeaderReq := &http.Request{Header: make(http.Header)}
	noTokenReq := &http.Request{Header: make(http.Header)}
	noTokenReq.Header.Set("Authorization", "Bearer ")

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	badKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	var authTests = []struct {
		req  *http.Request
		code stacktrace.ErrorCode
	}{
		{noHeaderReq, dsserr.Unauthenticated},
		{noTokenReq, dsserr.Unauthenticated},
		{rsaTokenReq(badKey, 100, 20), dsserr.Unauthenticated},
		{rsaTokenReq(key, 100, 20), stacktrace.NoCode},
		{rsaTokenReqWithMultipleAudience(key, 100, 20), stacktrace.NoCode},
		{rsaTokenReqWithMissingAudience(key, 100, 20), dsserr.Unauthenticated},
		{rsaTokenReqWithMissingIssuer(key, 100, 20), dsserr.Unauthenticated},
		{rsaTokenReq(key, 30, 20), dsserr.Unauthenticated},
		{rsaTokenReq(key, 100, 50), dsserr.Unauthenticated},
	}

	a, err := NewRSAAuthorizer(t.Context(), Configuration{
		KeyResolver: &fromMemoryKeyResolver{
			Keys: []interface{}{&key.PublicKey},
		},
		KeyRefreshInterval: 1 * time.Millisecond,
		AcceptedAudiences:  []string{"test-aud"},
	})

	require.NoError(t, err)

	for i, test := range authTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := t.Context()
			claimsValue, err := a.extractClaims(test.req)
			if err != nil {
				ctx = claims.NewContextFromError(ctx, err)
			} else {
				ctx = claims.NewContext(ctx, claimsValue)
			}

			res := a.Authorize(nil, test.req.WithContext(ctx), []api.AuthorizationOption{})
			if test.code != stacktrace.ErrorCode(0) && stacktrace.GetCode(res.Error) != test.code {
				t.Logf("%v", res.Error)
				t.Errorf("expected: %v, got: %v, with message %s", test.code, stacktrace.GetCode(res.Error), res.Error.Error())
			}
		})
	}
}

func TestRSAAuthAudiences(t *testing.T) {

	var tests = []struct {
		Accepted           []string
		Provided           interface{}
		ShouldBeAuthorized bool
	}{
		{
			[]string{"aud1", "aud2"},
			[]string{"aud1", "aud2", "aud3"},
			true,
		},
		{
			[]string{"aud1", "aud2"},
			[]string{"aud2", "aud3"},
			true,
		},
		{
			[]string{"aud1", "aud2"},
			[]string{"aud1", "aud3"},
			true,
		},
		{
			[]string{"aud1", "aud2"},
			[]string{"aud1", "aud2"},
			true,
		},
		{
			[]string{"aud1", "aud2"},
			[]string{"aud1"},
			true,
		},
		{
			[]string{"aud1", "aud2"},
			[]string{"aud2"},
			true,
		},
		{
			[]string{"aud1", "aud2"},
			[]string{"aud3"},
			false,
		},
		{
			[]string{"aud1", "aud2"},
			"aud1",
			true,
		},
		{
			[]string{"aud1", "aud2"},
			"aud2",
			true,
		},
		{
			[]string{"aud1", "aud2"},
			"aud3",
			false,
		},
		{
			[]string{"aud1", "aud2"},
			[]string{},
			false,
		},
		{
			[]string{"aud1", "aud2"},
			"",
			false,
		},
		{
			[]string{"aud1", "aud2"},
			nil,
			false,
		},
		{
			[]string{"aud1"},
			[]string{"aud1", "aud2", "aud3"},
			true,
		},
		{
			[]string{"aud1"},
			[]string{"aud2", "aud3"},
			false,
		},
		{
			[]string{"aud1"},
			[]string{"aud2", "aud1"},
			true,
		},
		{
			[]string{"aud1"},
			[]string{"aud1"},
			true,
		},
		{
			[]string{"aud1"},
			[]string{"aud2"},
			false,
		},
		{
			[]string{"aud1"},
			"aud1",
			true,
		},
		{
			[]string{"aud1"},
			"aud2",
			false,
		},
		{
			[]string{"aud1"},
			[]string{},
			false,
		},
		{
			[]string{"aud1"},
			"",
			false,
		},
		{
			[]string{"aud1"},
			nil,
			false,
		},
		{
			[]string{},
			[]string{"aud1"},
			false,
		},
		{
			[]string{},
			"aud1",
			false,
		},
		{
			[]string{},
			[]string{},
			false,
		},
		{
			[]string{},
			"",
			false,
		},
		{
			[]string{},
			nil,
			false,
		},
	}

	jwt.TimeFunc = func() time.Time {
		return time.Unix(42, 0)
	}

	defer func() {
		jwt.TimeFunc = time.Now
	}()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			a, err := NewRSAAuthorizer(t.Context(), Configuration{
				KeyResolver: &fromMemoryKeyResolver{
					Keys: []interface{}{&key.PublicKey},
				},
				KeyRefreshInterval: 1 * time.Millisecond,
				AcceptedAudiences:  test.Accepted,
			})
			require.NoError(t, err)

			req := rsaTokenReqWithAudiences(key, 100, 20, test.Provided)
			code := dsserr.Unauthenticated

			if test.ShouldBeAuthorized {
				code = stacktrace.NoCode
			}

			ctx := t.Context()
			claimsValue, err := a.extractClaims(req)
			if err != nil {
				ctx = claims.NewContextFromError(ctx, err)
			} else {
				ctx = claims.NewContext(ctx, claimsValue)
			}

			res := a.Authorize(nil, req.WithContext(ctx), []api.AuthorizationOption{})
			if code != stacktrace.ErrorCode(0) && stacktrace.GetCode(res.Error) != code {
				t.Logf("%v", res.Error)
				t.Errorf("expected: %v, got: %v, with message %s", code, stacktrace.GetCode(res.Error), res.Error.Error())
			}
		})
	}
}

func TestMissingScopes(t *testing.T) {
	authOptions := []api.AuthorizationOption{
		{"TestAuth1": {"required1"}},
		{"TestAuth2": {"required2"}},
		{"TestAuth3": {"required3", "required4"}},
	}

	var tests = []struct {
		scopes                map[string]struct{}
		matchesRequiredScopes bool
	}{
		{
			map[string]struct{}{
				"required1": {},
				"required2": {},
			},
			true,
		},
		{
			map[string]struct{}{
				"required2": {},
			},
			true,
		},
		{
			map[string]struct{}{
				"required1": {},
			},
			true,
		},
		{
			map[string]struct{}{},
			false,
		},
		{
			map[string]struct{}{
				"required3": {},
				"required4": {},
			},
			true,
		},
		{
			map[string]struct{}{
				"required4": {},
			},
			false,
		},
		{
			map[string]struct{}{
				"required3": {},
			},
			false,
		},
		{
			map[string]struct{}{
				"required1": {},
				"required3": {},
				"required4": {},
			},
			true,
		},
	}
	for _, tc := range tests {
		pass, _ := validateScopes(authOptions, tc.scopes)
		require.Equal(t, tc.matchesRequiredScopes, pass)
	}
}

func TestClaimsValidation(t *testing.T) {
	claims.Now = func() time.Time {
		return time.Unix(42, 0)
	}
	jwt.TimeFunc = claims.Now

	defer func() {
		jwt.TimeFunc = time.Now
		claims.Now = time.Now
	}()

	claims := &claims.Claims{}

	require.Error(t, claims.Valid())

	claims.Subject = "real_owner"
	claims.ExpiresAt = jwt.NewNumericDate(time.Unix(45, 0))
	claims.Issuer = "real_issuer"

	require.NoError(t, claims.Valid())

	// Test error out on expired token Now.Unix() = 42
	claims.ExpiresAt = jwt.NewNumericDate(time.Unix(41, 0))
	require.Error(t, claims.Valid())

	// Test error out on missing Issuer URI
	claims.Issuer = ""
	claims.ExpiresAt = jwt.NewNumericDate(time.Unix(45, 0))
	require.Error(t, claims.Valid())
}

func TestHasScope(t *testing.T) {
	scopes := []string{
		string(scdv1.UtmStrategicCoordinationScope),
		string(scdv1.UtmConformanceMonitoringSaScope),
	}

	require.True(t, HasScope(scopes, scdv1.UtmStrategicCoordinationScope))
	require.True(t, HasScope(scopes, scdv1.UtmConformanceMonitoringSaScope))
	require.False(t, HasScope(scopes, scdv1.UtmAvailabilityArbitrationScope))
}

type failingKeyResolver struct {
	mu   sync.Mutex
	keys []interface{}
	err  error
}

func (r *failingKeyResolver) ResolveKeys(context.Context) ([]interface{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	return r.keys, nil
}

func (r *failingKeyResolver) setErr(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.err = err
}

type ignorePanicHook struct{}

func (ignorePanicHook) OnWrite(*zapcore.CheckedEntry, []zapcore.Field) {}

func observeLogs(t *testing.T) *observer.ObservedLogs {
	core, logs := observer.New(zapcore.DebugLevel)
	previous := logging.Logger
	logging.Logger = zap.New(core, zap.WithPanicHook(ignorePanicHook{}))
	t.Cleanup(func() { logging.Logger = previous })
	return logs
}

func TestKeyRefreshFailureKeepsCachedKeys(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	logs := observeLogs(t)
	resolver := &failingKeyResolver{keys: []interface{}{&key.PublicKey}}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	a, err := NewRSAAuthorizer(ctx, Configuration{
		KeyResolver:        resolver,
		KeyRefreshInterval: 1 * time.Millisecond,
		KeyTTL:             1 * time.Hour,
	})
	require.NoError(t, err)

	resolver.setErr(stacktrace.NewErrorWithCode(dsserr.Unavailable, "JWKS endpoint is unreachable"))

	require.Eventually(t, func() bool {
		return logs.FilterLevelExact(zapcore.ErrorLevel).Len() > 0
	}, time.Second, time.Millisecond)

	require.Zero(t, logs.FilterLevelExact(zapcore.PanicLevel).Len())

	a.keyGuard.RLock()
	defer a.keyGuard.RUnlock()
	require.Equal(t, []interface{}{&key.PublicKey}, a.keys)
}

func TestKeyRefreshFailurePanicsAfterMaxCacheAge(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	logs := observeLogs(t)
	resolver := &failingKeyResolver{keys: []interface{}{&key.PublicKey}}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	_, err = NewRSAAuthorizer(ctx, Configuration{
		KeyResolver:        resolver,
		KeyRefreshInterval: 1 * time.Millisecond,
		KeyTTL:             200 * time.Millisecond,
	})
	require.NoError(t, err)

	resolver.setErr(stacktrace.NewErrorWithCode(dsserr.Unavailable, "JWKS endpoint is unreachable"))

	require.Eventually(t, func() bool {
		return logs.FilterLevelExact(zapcore.ErrorLevel).Len() > 0
	}, time.Second, time.Millisecond)
	require.Zero(t, logs.FilterLevelExact(zapcore.PanicLevel).Len())

	require.Eventually(t, func() bool {
		return logs.FilterLevelExact(zapcore.PanicLevel).Len() > 0
	}, time.Second, time.Millisecond)
}

func TestKeyRefreshFailurePanicsImmediatelyWithoutMaxCacheAge(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	logs := observeLogs(t)
	resolver := &failingKeyResolver{keys: []interface{}{&key.PublicKey}, err: nil}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	_, err = NewRSAAuthorizer(ctx, Configuration{
		KeyResolver:        resolver,
		KeyRefreshInterval: 1 * time.Millisecond,
	})
	require.NoError(t, err)

	resolver.setErr(stacktrace.NewErrorWithCode(dsserr.Unavailable, "JWKS endpoint is unreachable"))

	require.Eventually(t, func() bool {
		return logs.FilterLevelExact(zapcore.PanicLevel).Len() > 0
	}, time.Second, time.Millisecond)

	require.Zero(t, logs.FilterLevelExact(zapcore.ErrorLevel).Len())
}

func TestKeyRefreshFailurePanicsOnNonRetryableError(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	logs := observeLogs(t)
	resolver := &failingKeyResolver{keys: []interface{}{&key.PublicKey}}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	_, err = NewRSAAuthorizer(ctx, Configuration{
		KeyResolver:        resolver,
		KeyRefreshInterval: 1 * time.Millisecond,
		KeyTTL:             1 * time.Hour,
	})
	require.NoError(t, err)

	resolver.setErr(stacktrace.NewError("Failed to resolve key(s) for ID: unknown"))

	require.Eventually(t, func() bool {
		return logs.FilterLevelExact(zapcore.PanicLevel).Len() > 0
	}, time.Second, time.Millisecond)

	require.Zero(t, logs.FilterLevelExact(zapcore.ErrorLevel).Len())
}

func TestJWKSResolverErrorCodes(t *testing.T) {
	for _, c := range []struct {
		name   string
		status int
		body   string
		keyIDs []string
		code   stacktrace.ErrorCode
	}{
		{"server error", http.StatusInternalServerError, "", nil, dsserr.Unavailable},
		{"client error", http.StatusNotFound, "", nil, dsserr.Unavailable},
		{"invalid body", http.StatusOK, "not json", nil, dsserr.Unavailable},
		{"unknown key id", http.StatusOK, `{"keys":[]}`, []string{"unknown"}, stacktrace.NoCode},
	} {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(c.status)
				_, _ = w.Write([]byte(c.body))
			}))
			defer server.Close()

			endpoint, err := url.Parse(server.URL)
			require.NoError(t, err)

			resolver := &JWKSResolver{Endpoint: endpoint, KeyIDs: c.keyIDs}
			_, err = resolver.ResolveKeys(t.Context())
			require.Error(t, err)
			require.Equal(t, c.code, stacktrace.GetCode(err))
		})
	}
}
