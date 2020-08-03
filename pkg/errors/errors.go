package errors

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	obfuscateInternalErrors = true
)

const (
	// AreaTooLargeErr is the error that we want to signal to the http gateway
	// that it should return 413 to client
	AreaTooLargeErr codes.Code = 18

	// MissingOVNs is the error to signal that an AirspaceConflictResponse should
	// be returned rather than the standard error response.
	MissingOVNs codes.Code = 19
)

func init() {
	if s, ok := os.LookupEnv("DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS"); ok {
		if b, err := strconv.ParseBool(s); err == nil {
			obfuscateInternalErrors = b
		}
	}
}

func MakeErrID() string {
	errUUID, err := uuid.NewRandom()
	if err == nil {
		return errUUID.String()
	}
	return fmt.Sprintf("<error ID could not be constructed: %s>", err)
}

// Interceptor returns a grpc.UnaryServerInterceptor that inspects outgoing
// errors and logs (to "logger") and replaces errors that are not *status.Status
// instances or status instances that indicate an internal/unknown error.
func Interceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)

		if err != nil {
			// Separate the root cause from the stacktrace wrapping.
			trace := err.Error()
			rootErr := stacktrace.RootCause(err)

			statusErr, ok := status.FromError(rootErr)
			switch {
			case !ok:
				errID := MakeErrID()
				logger.Error(
					fmt.Sprintf("encountered non-Status error %s during unary server call", errID),
					zap.String("method", info.FullMethod),
					zap.String("stacktrace", trace),
					zap.Error(rootErr))
				if obfuscateInternalErrors {
					err = status.Error(codes.Internal, fmt.Sprintf("Internal server error %s", errID))
				} else {
					err = status.Error(codes.Internal, err.Error())
				}
			case statusErr.Code() == codes.Internal, statusErr.Code() == codes.Unknown:
				errID := MakeErrID()
				logger.Error(
					fmt.Sprintf("encountered internal Status error %s during unary server call", errID),
					zap.String("method", info.FullMethod),
					zap.Stringer("code", statusErr.Code()),
					zap.String("message", statusErr.Message()),
					zap.Any("details", statusErr.Details()),
					zap.String("stacktrace", trace),
					zap.Error(rootErr))
				if obfuscateInternalErrors {
					err = status.Error(codes.Internal, fmt.Sprintf("Internal server error %s", errID))
				}
			}
		}
		return
	}
}

// AlreadyExists returns an error used when creating a resource that already
// exists.
func AlreadyExists(id string) error {
	return status.Error(codes.AlreadyExists, "resource already exists: "+id)
}

// VersionMismatch returns an error used when updating a resource with an old
// version.
func VersionMismatch(msg string) error {
	return status.Error(codes.Aborted, msg)
}

// NotFound returns an error used when looking up a resource that doesn't exist.
func NotFound(id string) error {
	return status.Error(codes.NotFound, "resource not found: "+id)
}

// BadRequest returns an error that is used when a user supplies bad request
// parameters.
func BadRequest(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

// Internal returns an error that represents an internal DSS error.
func Internal(msg string) error {
	return status.Error(codes.Internal, msg)
}

// Exhausted is used when a USS creates too many resources in a given area.
func Exhausted(msg string) error {
	return status.Error(codes.ResourceExhausted, msg)
}

// PermissionDenied returns an error representing a bad Oauth token. It can
// occur when a user attempts to modify a resource "owned" by a different USS.
func PermissionDenied(msg string) error {
	return status.Error(codes.PermissionDenied, msg)
}

// Unauthenticated returns an error that is used when an Oauth token is invalid
// or not supplied.
func Unauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}

// AreaTooLarge is used when a user tries to create a resource in an area larger
// than the max area allowed. See geo/s2.go.
func AreaTooLarge(msg string) error {
	return status.Error(AreaTooLargeErr, msg)
}
