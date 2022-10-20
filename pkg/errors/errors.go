package errors

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
)

const (
	// AreaTooLarge is used when a user tries to create a resource in an area
	// larger than the max area allowed. See geo/s2.go.
	AreaTooLarge = stacktrace.ErrorCode(iota)

	// MissingOVNs is the error to signal that an AirspaceConflictResponse should
	// be returned rather than the standard error response.
	MissingOVNs

	// AlreadyExists is used when attempting to create a resource that already
	// exists.
	AlreadyExists

	// BadRequest is used when a user supplies bad request parameters.
	BadRequest

	// VersionMismatch is used when updating a resource with an old version.
	VersionMismatch

	// NotFound is used when looking up a resource that doesn't exist.
	NotFound

	// PermissionDenied is used to represent a bad OAuth token. It can occur when
	// a user attempts to modify a resource "owned" by a different USS.
	PermissionDenied

	// Exhausted is used when a USS creates too many resources in a given area.
	Exhausted

	// Unauthenticated is used when an OAuth token is invalid or not supplied.
	Unauthenticated
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

// Handle parses and handles an error that happen in a REST handler. The error
// is logged and a message appropriate for the requesting client is returned.
func Handle(ctx context.Context, err error) *string {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	errID := MakeErrID()

	if err == nil {
		err = stacktrace.NewError("Error to handle is nil")
	}

	// Separate the root cause and code from the stacktrace wrapping.
	rootErr := stacktrace.RootCause(err)
	code := stacktrace.GetCode(err)

	logger = logger.With(
		zap.String("error_id", errID),
		zap.String("stacktrace", err.Error()),
		zap.Error(rootErr))

	if code != stacktrace.NoCode {
		logger.Error("Error during unary server call", zap.Int("code", int(code)))
	} else {
		logger.Error("Uncoded error during unary server call")
	}

	errMsg := fmt.Sprintf("%s (%s)", rootErr.Error(), errID)
	return &errMsg
}
