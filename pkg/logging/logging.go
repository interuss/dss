package logging

import (
	"context"
	"os"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
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
	// Make sure that log statements internal to gRPC library are logged using the Logger as well.
	grpcReplaceLogger(Logger)
}

func Interceptor() grpc.UnaryServerInterceptor {
	opts := []grpc_zap.Option{
		grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel),
	}
	return grpc_middleware.ChainUnaryServer(
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(Logger, opts...),
	)
}

// WithValuesFromContext augments logger with relevant fields from ctx and returns
// the the resulting logger.
func WithValuesFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	// Naive implementation for now, meant to evolve over time.
	return logger
}
