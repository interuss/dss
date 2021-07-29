package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/models"

	"github.com/golang-jwt/jwt"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func rsaTokenCtx(ctx context.Context, key *rsa.PrivateKey, exp, nbf int64) context.Context {
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
	return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	}))
}
func rsaTokenCtxWithMissingIssuer(ctx context.Context, key *rsa.PrivateKey, exp, nbf int64) context.Context {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"foo": "bar",
		"exp": exp,
		"nbf": nbf,
		"sub": "real_owner",
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	}))
}

func TestNewRSAAuthClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpfile, err := ioutil.TempFile("/tmp", "bad.pem")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())
	// Test catches previous segfault.
	_, err = NewRSAAuthorizer(ctx, Configuration{
		KeyResolver: &FromFileKeyResolver{
			KeyFiles: []string{tmpfile.Name()},
		},
		KeyRefreshTimeout: 1 * time.Millisecond,
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal(err)
	}
	badKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal(err)
	}
	var authTests = []struct {
		ctx  context.Context
		code stacktrace.ErrorCode
	}{
		{ctx, dsserr.Unauthenticated},
		{metadata.NewIncomingContext(ctx, metadata.New(nil)), dsserr.Unauthenticated},
		{rsaTokenCtx(ctx, badKey, 100, 20), dsserr.Unauthenticated},
		{rsaTokenCtx(ctx, key, 100, 20), stacktrace.NoCode},
		{rsaTokenCtxWithMissingIssuer(ctx, key, 100, 20), dsserr.Unauthenticated},
		{rsaTokenCtx(ctx, key, 30, 20), dsserr.Unauthenticated},
		{rsaTokenCtx(ctx, key, 100, 50), dsserr.Unauthenticated},
	}

	a, err := NewRSAAuthorizer(ctx, Configuration{
		KeyResolver: &fromMemoryKeyResolver{
			Keys: []interface{}{&key.PublicKey},
		},
		KeyRefreshTimeout: 1 * time.Millisecond,
		AcceptedAudiences: []string{""},
	})

	require.NoError(t, err)

	for i, test := range authTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := a.AuthInterceptor(test.ctx, nil, &grpc.UnaryServerInfo{},
				func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
			if test.code != stacktrace.ErrorCode(0) && stacktrace.GetCode(err) != test.code {
				t.Errorf("expected: %v, got: %v, with message %s", test.code, status.Code(err), err.Error())
			}
		})
	}
}

func TestMissingScopes(t *testing.T) {
	ac := &Authorizer{scopesValidators: map[Operation]KeyClaimedScopesValidator{
		"/dss.SyncService/PutFoo": RequireAnyScope(("required1"), Scope("required2")),
	}}

	var tests = []struct {
		info                  *grpc.UnaryServerInfo
		scopes                map[Scope]struct{}
		matchesRequiredScopes bool
	}{
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss.SyncService/PutFoo"},
			map[Scope]struct{}{
				"required1": {},
				"required2": {},
			},
			true,
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss.SyncService/PutFoo"},
			map[Scope]struct{}{
				"required2": {},
			},
			true,
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss.SyncService/PutFoo"},
			map[Scope]struct{}{
				"required1": {},
			},
			true,
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss.SyncService/PutFoo"},
			map[Scope]struct{}{},
			false,
		},
	}
	for _, tc := range tests {
		require.Equal(t, tc.matchesRequiredScopes, ac.validateKeyClaimedScopes(context.Background(), tc.info, tc.scopes) == nil)
	}
}

func TestClaimsValidation(t *testing.T) {
	Now = func() time.Time {
		return time.Unix(42, 0)
	}
	jwt.TimeFunc = Now

	defer func() {
		jwt.TimeFunc = time.Now
		Now = time.Now
	}()

	claims := &claims{}

	require.Error(t, claims.Valid())

	claims.Subject = "real_owner"
	claims.ExpiresAt = 45
	claims.Issuer = "real_issuer"

	require.NoError(t, claims.Valid())

	// Test error out on expired token Now.Unix() = 42
	claims.ExpiresAt = 41
	require.Error(t, claims.Valid())

	// Test error out on missing Issuer URI
	claims.Issuer = ""
	claims.ExpiresAt = 45
	require.Error(t, claims.Valid())
}

func TestContextWithOwner(t *testing.T) {
	ctx := context.Background()
	_, ok := OwnerFromContext(ctx)
	require.False(t, ok)

	ctx = ContextWithOwner(ctx, "real_owner")
	owner, ok := OwnerFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, models.Owner("real_owner"), owner)
}
