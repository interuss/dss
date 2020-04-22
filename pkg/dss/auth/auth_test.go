package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/dss/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
			KeyFile: tmpfile.Name(),
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
		ctx          context.Context
		code         codes.Code
		errorMessage string
	}{
		{ctx, codes.Unauthenticated, "missing token"},
		{metadata.NewIncomingContext(ctx, metadata.New(nil)), codes.Unauthenticated, "missing token"},
		{rsaTokenCtx(ctx, badKey, 100, 20), codes.Unauthenticated, "crypto/rsa: verification error"},
		{rsaTokenCtx(ctx, key, 100, 20), codes.OK, ""},
		{rsaTokenCtxWithMissingIssuer(ctx, key, 100, 20), codes.Unauthenticated, "missing Issuer"},
		{rsaTokenCtx(ctx, key, 30, 20), codes.Unauthenticated, "token is expired"},
		{rsaTokenCtx(ctx, key, 100, 50), codes.Unauthenticated, "token is not valid yet"},
	}

	a, err := NewRSAAuthorizer(ctx, Configuration{
		KeyResolver: &fromMemoryKeyResolver{
			Key: &key.PublicKey,
		},
		KeyRefreshTimeout: 1 * time.Millisecond,
	})

	require.NoError(t, err)

	for _, test := range authTests {
		_, err := a.AuthInterceptor(test.ctx, nil, &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
		if status.Code(err) != test.code {
			t.Errorf("expected: %v, got: %v", test.code, status.Code(err))
		}
		if err != nil && !strings.Contains(err.Error(), test.errorMessage) {
			t.Errorf("expected: %v, got: %v", test.errorMessage, err.Error())
		}
	}
}

func TestMissingScopes(t *testing.T) {
	ac := &Authorizer{requiredScopes: map[string][]string{
		"PutFoo": {"required1", "required2"},
	}}

	var tests = []struct {
		info   *grpc.UnaryServerInfo
		scopes map[string]struct{}
		want   error
	}{
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			map[string]struct{}{
				"required1": {},
				"required2": {},
			},
			nil,
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			map[string]struct{}{
				"required2": {},
			},
			&missingScopesError{[]string{"required1"}},
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			map[string]struct{}{
				"required1": {},
			},
			&missingScopesError{[]string{"required2"}},
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			map[string]struct{}{},
			&missingScopesError{[]string{"required1", "required2"}},
		},
	}
	for _, tc := range tests {
		got := ac.missingScopes(tc.info, tc.scopes)
		want := tc.want
		// both are nil, terminate early.
		if got == want {
			continue
		}
		// 1 is nil, and the other is not
		if (got == nil) != (want == nil) {
			t.Errorf("got: %s, want %s", got, want)
		}
		// Neither are nil, but maybe still don't equal each other
		if got.Error() != want.Error() {
			t.Errorf("got: %s, want %s", got, want)
		}
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
