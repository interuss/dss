package scd

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// ReceivedReportHandler takes care of handling a DSS report received through the MakeDssReport REST handler.
type ReceivedReportHandler interface {
	// Handle a DSS report request. Returns the error report passed in 'req' after having set its identifier.
	Handle(ctx context.Context, req *restapi.MakeDssReportRequest) (*restapi.ErrorReport, error)
}

// JSONLoggingReceivedReportHandler a DSSReportHandler that simply logs the received report as JSON.
type JSONLoggingReceivedReportHandler struct {
	// ReportLogger is the logger to which the received reports will be logged.
	ReportLogger *zap.Logger
}

// HandleDssReport logs the received report as a JSON string to a logger.
func (h *JSONLoggingReceivedReportHandler) Handle(ctx context.Context, req *restapi.MakeDssReportRequest) (*restapi.ErrorReport, error) {
	reportID, err := uuid.NewRandom()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to generate report ID")
	}
	rVal := req.Body
	reportIDStr := reportID.String()
	rVal.ReportId = &reportIDStr
	// Serialize the report to a JSON string:
	jsonReport, err := json.Marshal(req.Body)
	if err != nil {
		logging.WithValuesFromContext(ctx, logging.Logger).Error("Failed to serialize DSS Report", zap.Error(err))
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to serialize DSS Report")
	}
	h.ReportLogger.Info("Received DSS Report", zap.String("reportID", reportIDStr), zap.String("report", string(jsonReport)))
	return rVal, nil
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *restapi.MakeDssReportRequest,
) restapi.MakeDssReportResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.MakeDssReportResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.MakeDssReportResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	report, err := a.DSSReportHandler.Handle(ctx, req)

	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete operational intent")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.MakeDssReportResponseSet{Response400: errResp}
		case dsserr.PermissionDenied:
			return restapi.MakeDssReportResponseSet{Response403: errResp}
		default:
			return restapi.MakeDssReportResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, err)}}
		}
	}

	return restapi.MakeDssReportResponseSet{Response201: report}
}
