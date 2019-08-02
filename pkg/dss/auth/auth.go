package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authClient struct {
	key interface{}
}

// NewSymmetricAuthClient returns a new authClient instance using symmetric keys.
func NewSymmetricAuthClient(keyFile string) (*authClient, error) {
	bytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return &authClient{key: bytes}, nil
}

// NewRSAAuthClient returns a new authClient instance which uses RSA.
func NewRSAAuthClient(keyFile string) (*authClient, error) {
	bytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	pub, _ := pem.Decode(bytes)
	parsedKey, err := x509.ParsePKIXPublicKey(pub.Bytes)
	key, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not create rsa public key from %s", keyFile)
	}
	return &authClient{key: key}, nil
}

func (a *authClient) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	tknStr, ok := getToken(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}
	// TODO(steeling): Modify to ParseWithClaims and inspect claims.
	_, err := jwt.Parse(tknStr, func(token *jwt.Token) (interface{}, error) {
		return a.key, nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return handler(ctx, req)
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
