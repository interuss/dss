package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var hmacSampleSecret = []byte("secret_key")

func symmetricTokenCtx(ctx context.Context, key []byte) context.Context {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"foo":       "bar",
		"client_id": "me",
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	}))
}

func rsaTokenCtx(ctx context.Context, key *rsa.PrivateKey) context.Context {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"foo":       "bar",
		"client_id": "me",
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	}))
}

func TestSymmetricAuthInterceptor(t *testing.T) {
	ctx := context.Background()
	var authTests = []struct {
		ctx  context.Context
		code codes.Code
	}{
		{ctx, codes.Unauthenticated},
		{metadata.NewIncomingContext(ctx, metadata.New(nil)), codes.Unauthenticated},
		{symmetricTokenCtx(ctx, []byte("bad_signing_key")), codes.Unauthenticated},
		{symmetricTokenCtx(ctx, hmacSampleSecret), codes.OK},
	}

	a := &authClient{key: hmacSampleSecret}

	for _, test := range authTests {
		_, err := a.AuthInterceptor(test.ctx, nil, &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
		if status.Code(err) != test.code {
			t.Errorf("expected: %v, got: %v", test.code, status.Code(err))
		}
	}
}

func TestRSAAuthInterceptor(t *testing.T) {
	ctx := context.Background()
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
		code codes.Code
	}{
		{ctx, codes.Unauthenticated},
		{metadata.NewIncomingContext(ctx, metadata.New(nil)), codes.Unauthenticated},
		{rsaTokenCtx(ctx, badKey), codes.Unauthenticated},
		{rsaTokenCtx(ctx, key), codes.OK},
	}

	a := &authClient{key: &key.PublicKey}

	for _, test := range authTests {
		_, err := a.AuthInterceptor(test.ctx, nil, &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
		if status.Code(err) != test.code {
			t.Errorf("expected: %v, got: %v", test.code, status.Code(err))
		}
	}
}

func TestMissingScopes(t *testing.T) {
	ac := &authClient{requiredScopes: map[string][]string{
		"PutFoo": []string{"required1", "required2"},
	}}

	var tests = []struct {
		info   *grpc.UnaryServerInfo
		claims claims
		want   error
	}{
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			claims{Scopes: []string{"required1", "required2"}},
			nil,
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			claims{Scopes: []string{"required2"}},
			&missingScopesError{[]string{"required1"}},
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			claims{Scopes: []string{"required1"}},
			&missingScopesError{[]string{"required2"}},
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			claims{Scopes: []string{}},
			&missingScopesError{[]string{"required1", "required2"}},
		},
	}
	for _, tc := range tests {
		got := ac.missingScopes(tc.info, tc.claims)
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
	claims := &claims{}
	require.Error(t, claims.Valid())
	claims.ClientID = "me"
	require.NoError(t, claims.Valid())
}

func TestContextWithOwner(t *testing.T) {
	ctx := context.Background()
	owner, ok := OwnerFromContext(ctx)
	require.False(t, ok)
	ctx = ContextWithOwner(ctx, "me")
	owner, ok = OwnerFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "me", owner)
}
