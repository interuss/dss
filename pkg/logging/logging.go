package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// DefaultLevel is the default log level.
	DefaultLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	// DefaultFormat is the default log format.
	DefaultFormat = FormatJSON
	// FormatConsole marks the console log format.
	FormatConsole = "console"
	// FormatJSON marks the JSON log format.
	FormatJSON = "json"
	// Logger is the default, system-wide logger.
	Logger *zap.Logger
)

func init() {
	var (
		format = "json"
		level  = DefaultLevel.String()
	)
	if v := os.Getenv("DSS_LOG_LEVEL"); v != "" {
		level = v
	}

	if v := os.Getenv("DSS_LOG_FORMAT"); v != "" {
		format = v
	}

	if err := setUpLogger(level, format); err != nil {
		panic(err)
	}
}

func setUpLogger(level string, format string) error {
	lvl := DefaultLevel
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	options := []zap.Option{
		zap.AddCaller(), zap.AddStacktrace(zapcore.PanicLevel),
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.StacktraceKey = "stack"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.NewProductionConfig()
	config.Level = lvl
	config.Encoding = format
	config.EncoderConfig = encoderConfig

	l, err := config.Build(options...)
	if err != nil {
		return err
	}

	Logger = l

	return nil
}

// Configure configures the default log "level" and the log "format".
func Configure(level string, format string) error {
	return setUpLogger(level, format)
}

// WithValuesFromContext augments logger with relevant fields from ctx and returns
// the resulting logger.
func WithValuesFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	// Naive implementation for now, meant to evolve over time.
	return logger
}
