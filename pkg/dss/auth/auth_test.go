package auth

import (
	"context"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var hmacSampleSecret = []byte("secret_key")

func tokenCtx(ctx context.Context, key []byte) context.Context {
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

func TestAuthInterceptor(t *testing.T) {
	ctx := context.Background()
	var authTests = []struct {
		ctx  context.Context
		code codes.Code
	}{
		{ctx, codes.Unauthenticated},
		{metadata.NewIncomingContext(ctx, metadata.New(nil)), codes.Unauthenticated},
		{tokenCtx(ctx, []byte("bad_signing_key")), codes.Unauthenticated},
		{tokenCtx(ctx, hmacSampleSecret), codes.OK},
	}

	a := &authClient{publicKey: hmacSampleSecret}

	for _, test := range authTests {
		_, err := a.AuthInterceptor(test.ctx, nil, &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
		if status.Code(err) != test.code {
			t.Errorf("expected: %v, got: %v", test.code, status.Code(err))
		}
	}
}
