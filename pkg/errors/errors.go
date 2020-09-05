package errors

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api/v1/auxpb"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	// AreaTooLarge is used when a user tries to create a resource in an area
	// larger than the max area allowed. See geo/s2.go.  We want to signal to the
	// http gateway that it should return 413 to client.
	AreaTooLarge stacktrace.ErrorCode = stacktrace.ErrorCode(18)

	// MissingOVNs is the error to signal that an AirspaceConflictResponse should
	// be returned rather than the standard error response.
	MissingOVNs stacktrace.ErrorCode = stacktrace.ErrorCode(19)

	// AlreadyExists is used when attempting to create a resource that already
	// exists.
	AlreadyExists stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.AlreadyExists))

	// BadRequest is used when a user supplies bad request parameters.
	BadRequest stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.InvalidArgument))

	// VersionMismatch is used when updating a resource with an old version.
	VersionMismatch stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.Aborted))

	// NotFound is used when looking up a resource that doesn't exist.
	NotFound stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.NotFound))

	// PermissionDenied is used to represent a bad OAuth token. It can occur when
	// PermissionDenied is used to represent a bad OAuth token. It can occur when
	// a user attempts to modify a resource "owned" by a different USS.
	PermissionDenied stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.PermissionDenied))

	// Exhausted is used when a USS creates too many resources in a given area.
	Exhausted stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.ResourceExhausted))

	// Unauthenticated is used when an OAuth token is invalid or not supplied.
	Unauthenticated stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.Unauthenticated))
)

func init() {
	if _, ok := os.LookupEnv("DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS"); ok {
		logging.Logger.Warn("DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS has been deprecated and will be removed in a future version")
	}
}

func MakeErrID() string {
	errUUID, err := uuid.NewRandom()
	if err == nil {
		return fmt.Sprintf("E:%s", errUUID.String())
	}
	return fmt.Sprintf("E:<error ID could not be constructed: %s>", err)
}

// MakeStatusProto adds the content of a proto as a detail to a Status proto
// consisting of the provided code and message.
func MakeStatusProto(code codes.Code, message string, detail proto.Message) (*spb.Status, error) {
	serialized, err := proto.MarshalOptions{Deterministic: true}.Marshal(detail)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error serializing detail proto")
	}

	p := &spb.Status{
		Code:    int32(code),
		Message: message,
		Details: []*any.Any{
			{
				TypeUrl: "github.com/interuss/dss/" + string(detail.ProtoReflect().Descriptor().FullName()),
				Value:   serialized,
			},
		},
	}
	return p, nil
}

// Interceptor returns a grpc.UnaryServerInterceptor that inspects outgoing
// errors and logs (to "logger") and replaces errors that are not *status.Status
// instances or status instances that indicate an internal/unknown error.
func Interceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		if err == nil {
			return resp, nil
		}

		errID := MakeErrID()

		// Separate the root cause and code from the stacktrace wrapping.
		trace := err.Error()
		rootErr := stacktrace.RootCause(err)
		code := stacktrace.GetCode(err)

		statusErr, ok := status.FromError(rootErr)
		if ok {
			// The root cause is a Status error; return it exactly as-is.
			logger.Error(
				fmt.Sprintf("Status error %s during unary server call", errID),
				zap.String("method", info.FullMethod),
				zap.String("stacktrace", trace),
				zap.String("grpc_code", statusErr.Code().String()),
				zap.Error(rootErr))
			return resp, rootErr
		}

		if code != stacktrace.NoCode {
			logger.Error(
				fmt.Sprintf("Error %s during unary server call", errID),
				zap.String("method", info.FullMethod),
				zap.String("stacktrace", trace),
				zap.String("grpc_code", codes.Code(uint16(code)).String()),
				zap.Int("code", int(code)),
				zap.Error(rootErr))
			p, constructionErr := MakeStatusProto(codes.Code(uint16(code)), rootErr.Error(), &auxpb.StandardErrorResponse{
				Error:   rootErr.Error(),
				Code:    int32(code),
				Message: rootErr.Error(),
				ErrorId: errID,
			})
			if constructionErr == nil {
				err = status.ErrorProto(p)
			} else {
				constructionErrID := MakeErrID()
				logger.Error(
					fmt.Sprintf("Error %s constructing StandardErrorResponse from %s", constructionErrID, errID),
					zap.Error(constructionErr))
				err = status.Error(codes.Internal, fmt.Sprintf("Internal server error %s", constructionErrID))
			}
		} else {
			logger.Error(
				fmt.Sprintf("Uncoded error %s during unary server call", errID),
				zap.String("method", info.FullMethod),
				zap.String("stacktrace", trace),
				zap.Error(rootErr))
			err = status.Error(codes.Internal, fmt.Sprintf("Internal server error %s", errID))
		}

		return resp, err
	}
}
