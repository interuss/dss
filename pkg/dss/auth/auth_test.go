package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
		"exp":       100,
		"nbf":       20,
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	}))
}

func rsaTokenCtx(ctx context.Context, key *rsa.PrivateKey, exp, nbf int64) context.Context {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"foo":       "bar",
		"client_id": "me",
		"exp":       exp,
		"nbf":       nbf,
	})

	// Sign and get the complete encoded token as a string using the secret
	// Ignore the error, it will fail the test anyways if it is not nil.
	tokenString, _ := token.SignedString(key)
	return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	}))
}

func TestNewRSAAuthClient(t *testing.T) {
	tmpfile, err := ioutil.TempFile("/tmp", "bad.pem")
	require.NoError(t, tmpfile.Close())
	// Test catches previous segfault.
	_, err = NewRSAAuthClient(tmpfile.Name(), nil)
	require.Error(t, err)
	require.NoError(t, os.Remove(tmpfile.Name()))
}

func TestRSAAuthInterceptor(t *testing.T) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(42, 0)
	}
	defer func() { jwt.TimeFunc = time.Now }()

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
		{rsaTokenCtx(ctx, badKey, 100, 20), codes.Unauthenticated},
		{rsaTokenCtx(ctx, key, 100, 20), codes.OK},
		{rsaTokenCtx(ctx, key, 30, 20), codes.Unauthenticated},  // Expired
		{rsaTokenCtx(ctx, key, 100, 50), codes.Unauthenticated}, // Not valid yet
	}

	a := &authClient{key: &key.PublicKey, logger: zap.L()}

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
		scopes []string
		want   error
	}{
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			[]string{"required1", "required2"},
			nil,
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			[]string{"required2"},
			&missingScopesError{[]string{"required1"}},
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			[]string{"required1"},
			&missingScopesError{[]string{"required2"}},
		},
		{
			&grpc.UnaryServerInfo{FullMethod: "/dss/syncservice/PutFoo"},
			[]string{},
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
	claims := &claims{}
	require.Error(t, claims.Valid())
	claims.ClientID = "me"
	require.NoError(t, claims.Valid())
}

func TestContextWithOwner(t *testing.T) {
	expected := models.Owner("me")
	ctx := context.Background()
	owner, ok := OwnerFromContext(ctx)
	require.False(t, ok)
	ctx = ContextWithOwner(ctx, expected)
	owner, ok = OwnerFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, expected, owner)
}
