package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the default, system-wide logger.
	Logger *zap.Logger
)

func init() {
	l, err := zap.NewProduction(
		zap.AddCaller(), zap.AddStacktrace(zapcore.PanicLevel),
	)
	if err != nil {
		panic(err)
	}
	Logger = l
}

// WithValuesFromContext augments logger with relevant fields from ctx and returns
// the the resulting logger.
func WithValuesFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	// Naive implementation for now, meant to evolve over time.
	return logger
}
