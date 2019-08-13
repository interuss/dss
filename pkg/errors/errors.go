package errors

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errInternal = status.Error(codes.Internal, "Internal Server Error")
)

// Interceptor returns a grpc.UnaryServerInterceptor that inspects outgoing
// errors and logs (to "logger") and replaces errors that are not *status.Status
// instances or status instances that indicate an internal/unknown error.
func Interceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		status, ok := status.FromError(err)

		switch {
		case !ok:
			logger.Error("encountered error during unary server call", zap.String("method", info.FullMethod), zap.Error(err))
			err = errInternal
		case status.Code() == codes.Internal, status.Code() == codes.Unknown:
			logger.Error("encountered internal error during unary server call",
				zap.String("method", info.FullMethod),
				zap.Stringer("code", status.Code()),
				zap.String("message", status.Message()),
				zap.Any("details", status.Details()),
				zap.Error(err))
			err = errInternal
		}
		return
	}
}

func AlreadyExists(id string) error {
	return status.Error(codes.AlreadyExists, "resource already exists: "+id)
}

func VersionMismatch(msg string) error {
	return status.Error(codes.Aborted, msg)
}

func Conflict(msg string) error {
	return status.Error(codes.Aborted, msg)
}

func NotFound(id string) error {
	return status.Error(codes.NotFound, "resource not found: "+id)
}

func BadRequest(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

func Internal(msg string) error {
	// Log and obfuscate any errors.
	return errInternal
}

func Exhausted(msg string) error {
	return status.Error(codes.ResourceExhausted, msg)
}

func PermissionDenied(msg string) error {
	return status.Error(codes.PermissionDenied, msg)
}

func Unauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}
