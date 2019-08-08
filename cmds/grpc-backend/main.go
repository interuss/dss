package main

import (
	"context"
	"flag"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/steeling/InterUSS-Platform/pkg/dss"
	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/cockroach"
	"github.com/steeling/InterUSS-Platform/pkg/dssproto"
	"github.com/steeling/InterUSS-Platform/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	address      = flag.String("addr", "127.0.0.1:8080", "address")
	cockroachURI = flag.String("cockroach_uri", "", "URI pointing to a cockroach instance")
	pkFile       = flag.String("public_key_file", "", "Path to public Key to use for JWT decoding.")
	reflectAPI   = flag.Bool("reflect_api", false, "Whether to reflect the API.")
	logFormat    = flag.String("log_format", logging.DefaultFormat, "The log format in {json, console}")
	logLevel     = flag.String("log_level", logging.DefaultLevel.String(), "The log level")
)

// RunGRPCServer starts the example gRPC service.
// "network" and "address" are passed to net.Listen.
func RunGRPCServer(ctx context.Context, address string) error {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer func() {
		if err := l.Close(); err != nil {
			logger.Error("Failed to close listener", zap.String("address", address), zap.Error(err))
		}
	}()

	store, err := cockroach.Dial(*cockroachURI)
	if err != nil {
		logger.Panic("Failed to open connection to CRDB", zap.String("uri", *cockroachURI), zap.Error(err))
	}

	if err := store.Bootstrap(ctx); err != nil {
		logger.Panic("Failed to bootstrap CRDB instance", zap.Error(err))
	}

	dssServer := &dss.Server{
		Store: store,
	}

	ac, err := auth.NewRSAAuthClient(*pkFile)
	if err != nil {
		return err
	}
	ac.RequireScopes(dssServer.AuthScopes())

	s := grpc.NewServer(grpc_middleware.WithUnaryServerChain(logging.Interceptor(), ac.AuthInterceptor))
	if err != nil {
		return err
	}
	if *reflectAPI {
		reflection.Register(s)
	}

	dssproto.RegisterDiscoveryAndSynchronizationServiceServer(s, dssServer)

	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()
	return s.Serve(l)
}

func main() {
	flag.Parse()

	if err := logging.Configure(*logLevel, *logFormat); err != nil {
		panic(err)
	}

	var (
		ctx    = context.Background()
		logger = logging.WithValuesFromContext(ctx, logging.Logger)
	)

	if err := RunGRPCServer(ctx, *address); err != nil {
		logger.Panic("Failed to execute service", zap.Error(err))
	}
	logger.Info("Shutting down gracefully")
}
