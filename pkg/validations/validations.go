package validations

import (
	"context"

	"github.com/google/uuid"

	dsserr "github.com/interuss/dss/pkg/errors"
	"google.golang.org/grpc"
)

// ReqWithID checks if the proto message contains an ID, so it can validate it
// is an appropriate UUID.
type ReqWithID interface {
	GetId() string
}

// ValidationInterceptor is a grpc Interceptor to validate incoming requests
// with UUID's are properly formatted.
func ValidationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := ValidateUUID(req); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// ValidateUUID contains the UUID validation check.
func ValidateUUID(req interface{}) error {
	r, ok := req.(ReqWithID)
	if !ok {
		// Request doesn't have an ID, nothing to validate here.
		return nil
	}
	if _, err := uuid.Parse(r.GetId()); err != nil {
		return dsserr.BadRequest("Invalid uuid")
	}
	return nil
}
