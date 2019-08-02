package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var hmacSampleSecret = []byte("secret_key")

func symmetricTokenCtx(ctx context.Context, key []byte) context.Context {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"foo": "bar",
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
		"foo": "bar",
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
