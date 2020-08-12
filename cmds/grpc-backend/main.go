package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/interuss/dss/pkg/api/v1/auxpb"
	"github.com/interuss/dss/pkg/api/v1/ridpb"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	aux "github.com/interuss/dss/pkg/aux_"
	"github.com/interuss/dss/pkg/build"
	"github.com/interuss/dss/pkg/cockroach"
	uss_errors "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	application "github.com/interuss/dss/pkg/rid/application"
	rid "github.com/interuss/dss/pkg/rid/server"
	ridc "github.com/interuss/dss/pkg/rid/store/cockroach"
	"github.com/interuss/dss/pkg/scd"
	scdc "github.com/interuss/dss/pkg/scd/store/cockroach"
	"github.com/interuss/dss/pkg/validations"
	"github.com/palantir/stacktrace"
	"github.com/coreos/go-semver/semver"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	// The code at this version requires a major schema version equal to this
	// value.
	RidRequiredMajorSchemaVersion = "3.0.0"

	// The code at this version requires a major schema version equal to this
	// value.
	ScdRequiredMajorSchemaVersion = "1.0.0"
)

var (
	address           = flag.String("addr", ":8081", "address")
	pkFile            = flag.String("public_key_files", "", "Path to public Keys to use for JWT decoding, separated by commas.")
	jwksEndpoint      = flag.String("jwks_endpoint", "", "URL pointing to an endpoint serving JWKS")
	jwksKeyIDs        = flag.String("jwks_key_ids", "", "IDs of a set of key in a JWKS, separated by commas")
	keyRefreshTimeout = flag.Duration("key_refresh_timeout", 1*time.Minute, "Timeout for refreshing keys for JWT verification")
	timeout           = flag.Duration("server timeout", 10*time.Second, "Default timeout for server calls")
	reflectAPI        = flag.Bool("reflect_api", false, "Whether to reflect the API.")
	logFormat         = flag.String("log_format", logging.DefaultFormat, "The log format in {json, console}")
	logLevel          = flag.String("log_level", logging.DefaultLevel.String(), "The log level")
	dumpRequests      = flag.Bool("dump_requests", false, "Log request and response protos")
	profServiceName   = flag.String("gcp_prof_service_name", "", "Service name for the Go profiler")
	enableSCD         = flag.Bool("enable_scd", false, "Enables the Strategic Conflict Detection API")
	locality          = flag.String("locality", "", "self-identification string used as CRDB table writer column")

	cockroachParams = struct {
		host            *string
		port            *int
		sslMode         *string
		sslDir          *string
		user            *string
		applicationName *string
	}{
		host:            flag.String("cockroach_host", "", "cockroach host to connect to"),
		port:            flag.Int("cockroach_port", 26257, "cockroach port to connect to"),
		sslMode:         flag.String("cockroach_ssl_mode", "disable", "cockroach sslmode"),
		user:            flag.String("cockroach_user", "root", "cockroach user to authenticate as"),
		sslDir:          flag.String("cockroach_ssl_dir", "", "directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key"),
		applicationName: flag.String("cockroach_application_name", "dss", "application name for tagging the connection to cockroach"),
	}

	jwtAudiences = flag.String("accepted_jwt_audiences", "", "comma-separated acceptable JWT `aud` claims")
)

func MustSupportRidSchema(ctx context.Context, store *ridc.Store) error {
	vs, err := store.GetVersion(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get database schema version for remote ID")
	}
	if vs.Equal(*semver.New("0.0.0")) {
		return stacktrace.NewError("Remote ID database has not been bootstrapped with Schema Manager, Please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas")
	}

	if vs.LessThan(*semver.New(RidRequiredMajorSchemaVersion)) {
		return stacktrace.NewError("Unsupported schema version for remote ID! Got %s, requires major version of %s.", vs, RidRequiredMajorSchemaVersion)
	}

	return nil
}

func MustSupportScdSchema(ctx context.Context, store *scdc.Store) error {
	vs, err := store.GetVersion(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get database schema version for strategic conflict detection")
	}
	if vs.Equal(*semver.New("0.0.0")) {
		return stacktrace.NewError("Strategic conflict detection database has not been bootstrapped with Schema Manager, Please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas")
	}

	if vs.LessThan(*semver.New(ScdRequiredMajorSchemaVersion)) {
		return stacktrace.NewError("Unsupported schema version for strategic conflict detection! Got %s, requires major version of %s.", vs, ScdRequiredMajorSchemaVersion)
	}

	return nil
}

func ConnectTo(dbName string) (*cockroach.DB, error) {
	uriParams := map[string]string{
		"host":             *cockroachParams.host,
		"port":             strconv.Itoa(*cockroachParams.port),
		"user":             *cockroachParams.user,
		"ssl_mode":         *cockroachParams.sslMode,
		"ssl_dir":          *cockroachParams.sslDir,
		"application_name": *cockroachParams.applicationName,
		"db_name":          dbName,
	}
	uri, err := cockroach.BuildURI(uriParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building URI")
	}
	db, err := cockroach.Dial(uri)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error dialing CockroachDB database at %s", uri)
	}
	return db, nil
}

// RunGRPCServer starts the example gRPC service.
// "network" and "address" are passed to net.Listen.
func RunGRPCServer(ctx context.Context, address string, locality string) error {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	if len(*jwtAudiences) == 0 {
		// TODO: Make this flag required once all parties can set audiences
		// correctly.
		logger.Warn("missing required --accepted_jwt_audiences")
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		return stacktrace.Propagate(err, "Error attempting to listen at %s", address)
	}
	defer func() {
		if err := l.Close(); err != nil {
			logger.Error("Failed to close listener", zap.String("address", address), zap.Error(err))
		}
	}()

	var (
		ridServer *rid.Server
		scdServer *scd.Server
		auxServer = &aux.Server{}
	)

	// Initialize remote ID

	ridCrdb, err := ConnectTo(ridc.DatabaseName)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to connect to remote ID database; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
	}
	ridStore := ridc.NewStore(ridCrdb, logger)

	err = MustSupportRidSchema(ctx, ridStore)
	if err != nil {
		return stacktrace.Propagate(err, "Required remote ID schema not supported by database")
	}

	ridServer = &rid.Server{
		App:      application.NewFromTransactor(ridStore, logger),
		Timeout:  *timeout,
		Locality: locality,
	}
	scopesValidators := auth.MergeOperationsAndScopesValidators(
		rid.AuthScopes(), auxServer.AuthScopes(),
	)

	// Initialize strategic conflict detection

	if *enableSCD {
		scdCrdb, err := ConnectTo(scdc.DatabaseName)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to connect to strategic conflict detection database; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
		}
		scdStore := scdc.NewStore(scdCrdb, logger)

		err = MustSupportScdSchema(ctx, scdStore)
		if err != nil {
			return stacktrace.Propagate(err, "Required strategic conflict detection schema not supported by database")
		}

		scdServer = &scd.Server{
			Store:   scdStore,
			Timeout: *timeout,
		}
		scopesValidators = auth.MergeOperationsAndScopesValidators(
			scopesValidators, scdServer.AuthScopes(),
		)
	}

	// Initialize access token validation

	var keyResolver auth.KeyResolver
	switch {
	case *pkFile != "":
		keyResolver = &auth.FromFileKeyResolver{
			KeyFiles: strings.Split(*pkFile, ","),
		}
	case *jwksEndpoint != "" && *jwksKeyIDs != "":
		u, err := url.Parse(*jwksEndpoint)
		if err != nil {
			return stacktrace.Propagate(err, "Error parsing JWKS URL")
		}

		keyResolver = &auth.JWKSResolver{
			Endpoint: u,
			KeyIDs:   strings.Split(*jwksKeyIDs, ","),
		}
	default:
		logger.Warn("operating without authorizing interceptor")
	}

	authorizer, err := auth.NewRSAAuthorizer(
		ctx, auth.Configuration{
			KeyResolver:       keyResolver,
			KeyRefreshTimeout: *keyRefreshTimeout,
			ScopesValidators:  scopesValidators,
			AcceptedAudiences: strings.Split(*jwtAudiences, ","),
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating RSA authorizer")
	}

	// Set up server functionality

	interceptors := []grpc.UnaryServerInterceptor{
		uss_errors.Interceptor(logger),
		logging.Interceptor(logger),
		authorizer.AuthInterceptor,
		validations.ValidationInterceptor,
	}
	if *dumpRequests {
		interceptors = append(interceptors, logging.DumpRequestResponseInterceptor(logger))
	}

	s := grpc.NewServer(grpc_middleware.WithUnaryServerChain(interceptors...))
	if err != nil {
		return stacktrace.Propagate(err, "Error creating new gRPC server")
	}
	if *reflectAPI {
		reflection.Register(s)
	}

	ridpb.RegisterDiscoveryAndSynchronizationServiceServer(s, ridServer)
	auxpb.RegisterDSSAuxServiceServer(s, auxServer)
	if *enableSCD {
		logger.Info("config", zap.Any("scd", "enabled"))
		scdpb.RegisterUTMAPIUSSDSSAndUSSUSSServiceServer(s, scdServer)
	}
	logger.Info("build", zap.Any("description", build.Describe()))

	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()
	return s.Serve(l)
}

func main() {
	flag.Parse()

	if err := logging.Configure(*logLevel, *logFormat); err != nil {
		panic(fmt.Sprintf("Failed to configure logging: %s", err.Error()))
	}

	var (
		ctx    = context.Background()
		logger = logging.WithValuesFromContext(ctx, logging.Logger)
	)
	if *profServiceName != "" {
		err := profiler.Start(profiler.Config{
			Service: *profServiceName})
		if err != nil {
			logger.Panic("Failed to start the profiler ", zap.Error(err))
		}
	}

	if err := RunGRPCServer(ctx, *address, *locality); err != nil {
		logger.Panic("Failed to execute service", zap.Error(err))
	}

	logger.Info("Shutting down gracefully")
}
