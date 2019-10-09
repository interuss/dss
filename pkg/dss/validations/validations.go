package validations

import (
	"context"

	"github.com/google/uuid"

	dsserr "github.com/interuss/dss/pkg/errors"
	"google.golang.org/grpc"
)

type ReqWithID interface {
	GetId() string
}

func ValidationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := ValidateUUID(req); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func ValidateUUID(req interface{}) error {
	r, ok := req.(ReqWithID)
	if !ok {
		// Request doesn't have an ID, nothing to validate here.
		return nil
	}
	if _, err := uuid.Parse(r.GetId()); err != nil {
		return dsserr.BadRequest("invalid uuid")
	}
	return nil
}
