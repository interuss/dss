package scd

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// JSONLoggingDssReportHandler a DssReportHandler that simply logs the received report as JSON.
type JSONLoggingDssReportHandler struct {
	// ReportLogger is the logger to which the received reports will be logged.
	ReportLogger *zap.Logger
}

// HandleDssReport logs the received report as a JSON string to a logger.
func (h JSONLoggingDssReportHandler) HandleDssReport(ctx context.Context, req *restapi.MakeDssReportRequest) (*restapi.ErrorReport, error) {
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
		return nil, stacktrace.Propagate(err, "Failed to serialize DSS Report")
	}
	h.ReportLogger.Info("Received DSS Report", zap.String("report", string(jsonReport)))
	return rVal, nil
}
