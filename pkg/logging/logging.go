package logging

import (
	"context"
	"os"

	"github.com/gogo/protobuf/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
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

	config := zap.NewProductionConfig()
	config.Level = lvl
	config.Encoding = format

	l, err := config.Build(options...)
	if err != nil {
		return err
	}

	Logger = l
	// Make sure that log statements internal to gRPC library are logged using the Logger as well.
	grpcReplaceLogger(Logger)

	return nil
}

// Configure configures the default log "level" and the log "format".
func Configure(level string, format string) error {
	return setUpLogger(level, format)
}

// Interceptor returns a grpc.UnaryServerInterceptor that logs incoming requests
// and associated tags to "logger".
func Interceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	opts := []grpc_zap.Option{
		grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel),
	}
	return grpc_middleware.ChainUnaryServer(
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(logger, opts...),
	)
}

// WithValuesFromContext augments logger with relevant fields from ctx and returns
// the the resulting logger.
func WithValuesFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	// Naive implementation for now, meant to evolve over time.
	return logger
}

func DumpRequestResponseInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger.Sugar().Infof("Request (%s):\n%s",
			info.FullMethod,
			proto.MarshalTextString(req.(proto.Message)))

		resp, err = handler(ctx, req)

		if resp != nil && err == nil {
			logger.Sugar().Infof("Response (%s):\n%s",
				info.FullMethod,
				proto.MarshalTextString(resp.(proto.Message)))
		}
		return
	}
}
