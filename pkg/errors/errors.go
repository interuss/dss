package errors

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api/v1/auxpb"
	"github.com/palantir/stacktrace"
	"go.uber.org/zap"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	obfuscateInternalErrors = true
)

const (
	// AreaTooLarge is used when a user tries to create a resource in an area
	// larger than the max area allowed. See geo/s2.go.  We want to signal to the
	// http gateway that it should return 413 to client.
	AreaTooLarge stacktrace.ErrorCode = stacktrace.ErrorCode(18)

	// MissingOVNs is the error to signal that an AirspaceConflictResponse should
	// be returned rather than the standard error response.
	MissingOVNs codes.Code = 19

	AlreadyExists stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.AlreadyExists))
	BadRequest    stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.InvalidArgument))

	// VersionMismatch returns an error used when updating a resource with an old
	// version.
	VersionMismatch stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.Aborted))

	// NotFound returns an error used when looking up a resource that doesn't exist.
	NotFound stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.NotFound))

	// PermissionDenied returns an error representing a bad Oauth token. It can
	// occur when a user attempts to modify a resource "owned" by a different USS.
	PermissionDenied stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.PermissionDenied))

	// Exhausted is used when a USS creates too many resources in a given area.
	Exhausted stacktrace.ErrorCode = stacktrace.ErrorCode(uint16(codes.ResourceExhausted))
)

func init() {
	if s, ok := os.LookupEnv("DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS"); ok {
		if b, err := strconv.ParseBool(s); err == nil {
			obfuscateInternalErrors = b
		}
	}
}

func makeErrID() string {
	errUUID, err := uuid.NewRandom()
	if err == nil {
		return errUUID.String()
	}
	return fmt.Sprintf("<error ID could not be constructed: %s>", err)
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
			return resp, err
		}

		// Separate the root cause from the stacktrace wrapping.
		trace := err.Error()
		rootErr := stacktrace.RootCause(err)

		errID := makeErrID()

		statusErr, ok := status.FromError(rootErr)
		if !ok {
			stacktraceCode := stacktrace.GetCode(err)
			if stacktraceCode != stacktrace.NoCode {
				logger.Error(
					fmt.Sprintf("stacktrace error %s during unary server call", errID),
					zap.String("method", info.FullMethod),
					zap.String("stacktrace", trace),
					zap.Int("code", int(stacktraceCode)),
					zap.Error(rootErr))
				p, constructionErr := MakeStatusProto(codes.Code(uint16(stacktraceCode)), rootErr.Error(), &auxpb.StandardErrorResponse{
					Error:   rootErr.Error(),
					Code:    int32(stacktraceCode),
					Message: rootErr.Error(),
					ErrorId: errID,
				})
				if constructionErr == nil {
					err = status.ErrorProto(p)
					logger.Error("Constructed Status from proto", zap.Int("detail_count", len(p.Details)), zap.Error(err))
				} else {
					logger.Error(
						fmt.Sprintf("Error constructing StandardErrorResponse with %s", errID),
						zap.Error(constructionErr))
					err = status.Error(codes.Internal, fmt.Sprintf("Internal server error %s"))
				}
			} else {
				logger.Error(
					fmt.Sprintf("Non-Status error %s during unary server call", errID),
					zap.String("method", info.FullMethod),
					zap.String("stacktrace", trace),
					zap.Error(rootErr))
				if obfuscateInternalErrors {
					err = status.Error(codes.Internal, fmt.Sprintf("Internal server error %s", errID))
				} else {
					err = status.Error(codes.Internal, err.Error())
				}
			}
		} else {
			logger.Error(
				fmt.Sprintf("Status error %s during unary server call", errID),
				zap.String("method", info.FullMethod),
				zap.Stringer("code", statusErr.Code()),
				zap.String("message", statusErr.Message()),
				zap.Any("details", statusErr.Details()),
				zap.String("stacktrace", trace),
				zap.Error(rootErr))
			if obfuscateInternalErrors && (statusErr.Code() == codes.Internal || statusErr.Code() == codes.Unknown) {
				err = status.Error(codes.Internal, fmt.Sprintf("Internal server error %s", errID))
			}
		}

		logger.Info("Just before returning from interceptor", zap.String("Error()", err.Error()), zap.Error(err))
		return resp, err
	}
}

// Internal returns an error that represents an internal DSS error.
func Internal(msg string) error {
	return status.Error(codes.Internal, msg)
}

// Unauthenticated returns an error that is used when an Oauth token is invalid
// or not supplied.
func Unauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}
