package main

import (
	"context"
	"flag"
	"net"
	"strconv"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/steeling/InterUSS-Platform/pkg/dss"
	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/cockroach"
	"github.com/steeling/InterUSS-Platform/pkg/dss/validations"
	"github.com/steeling/InterUSS-Platform/pkg/dssproto"
	uss_errors "github.com/steeling/InterUSS-Platform/pkg/errors"
	"github.com/steeling/InterUSS-Platform/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	address      = flag.String("addr", ":8081", "address")
	pkFile       = flag.String("public_key_file", "", "Path to public Key to use for JWT decoding.")
	reflectAPI   = flag.Bool("reflect_api", false, "Whether to reflect the API.")
	logFormat    = flag.String("log_format", logging.DefaultFormat, "The log format in {json, console}")
	logLevel     = flag.String("log_level", logging.DefaultLevel.String(), "The log level")
	dumpRequests = flag.Bool("dump_requests", false, "Log request and response protos")

	cockroachHost    = flag.String("cockroach_host", "", "cockroach host to connect to")
	cockroachPort    = flag.Int("cockroach_port", 26257, "cockroach port to connect to")
	cockroachSSLMode = flag.String("cockroach_ssl_mode", "disable", "cockroach sslmode")
	cockroachUser    = flag.String("cockroach_user", "root", "cockroach user to authenticate as")
	cockroachSSLDir  = flag.String("cockroach_ssl_dir", "", "directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key")

	jwtAudience = flag.String("jwt_audience", "", "Require that JWTs contain this `aud` claim")
)

// RunGRPCServer starts the example gRPC service.
// "network" and "address" are passed to net.Listen.
func RunGRPCServer(ctx context.Context, address string) error {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	if *jwtAudience == "" {
		// TODO: Make this flag required once all parties can set audiences
		// correctly.
		logger.Warn("missing required --jwt_audience")
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer func() {
		if err := l.Close(); err != nil {
			logger.Error("Failed to close listener", zap.String("address", address), zap.Error(err))
		}
	}()

	uriParams := map[string]string{
		"host":     *cockroachHost,
		"port":     strconv.Itoa(*cockroachPort),
		"user":     *cockroachUser,
		"ssl_mode": *cockroachSSLMode,
		"ssl_dir":  *cockroachSSLDir,
	}
	uri, err := cockroach.BuildURI(uriParams)
	if err != nil {
		logger.Panic("Failed to build URI", zap.Error(err))
	}

	store, err := cockroach.Dial(uri)
	if err != nil {
		logger.Panic("Failed to open connection to CRDB", zap.String("uri", uri), zap.Error(err))
	}

	if err := store.Bootstrap(ctx); err != nil {
		logger.Panic("Failed to bootstrap CRDB instance", zap.Error(err))
	}

	dssServer := &dss.Server{
		Store: store,
	}

	ac, err := auth.NewRSAAuthClient(*pkFile, logger)
	if err != nil {
		return err
	}
	ac.RequireScopes(dssServer.AuthScopes())
	ac.RequireAudience(*jwtAudience)

	interceptors := []grpc.UnaryServerInterceptor{
		uss_errors.Interceptor(logger),
		logging.Interceptor(logger),
		ac.AuthInterceptor,
		validations.ValidationInterceptor,
	}
	if *dumpRequests {
		interceptors = append(interceptors, logging.DumpRequestResponseInterceptor(logger))
	}

	s := grpc.NewServer(grpc_middleware.WithUnaryServerChain(interceptors...))
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
