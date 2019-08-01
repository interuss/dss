package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the default, system-wide logger.
	Logger *zap.Logger
)

func init() {
	options := []zap.Option{
		zap.AddCaller(), zap.AddStacktrace(zapcore.PanicLevel),
	}

	config := zap.NewProductionConfig()
	if v := os.Getenv("DSS_LOG_LEVEL"); v != "" {
		lvl := zapcore.InfoLevel
		if err := lvl.UnmarshalText([]byte(v)); err != nil {
			panic(err)
		}
		config.Level = zap.NewAtomicLevelAt(lvl)
	}

	l, err := config.Build(options...)
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
