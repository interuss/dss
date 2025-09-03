package v1

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/ridv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"github.com/robfig/cron/v3"

	"github.com/interuss/dss/pkg/rid/application"
)

// Server implements ridv1.Implementation.
type Server struct {
	App               application.App
	Timeout           time.Duration
	Locality          string
	AllowHTTPBaseUrls bool
	Cron              *cron.Cron
}

func setAuthError(ctx context.Context, authErr error, resp401, resp403 **restapi.ErrorResponse, resp500 **api.InternalServerErrorBody) {
	switch stacktrace.GetCode(authErr) {
	case dsserr.Unauthenticated:
		*resp401 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authentication failed"))}
	case dsserr.PermissionDenied:
		*resp403 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authorization failed"))}
	default:
		if authErr == nil {
			authErr = stacktrace.NewError("Unknown error")
		}
		*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}
	}
}
