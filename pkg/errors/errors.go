package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AlreadyExists(msg string) error {
	return status.Error(codes.AlreadyExists, msg)
}

func VersionMismatch(msg string) error {
	return status.Error(codes.Aborted, msg)
}

func Conflict(msg string) error {
	return status.Error(codes.Aborted, msg)
}

func NotFound(msg string) error {
	return status.Error(codes.NotFound, msg)
}

func BadRequest(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

func Internal(msg string) error {
	// Log and obfuscate any errors.
	return status.Error(codes.Internal, "Internal Server Error")
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
