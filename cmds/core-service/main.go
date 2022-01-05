package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/interuss/dss/pkg/api/v1/auxpb"
	"github.com/interuss/dss/pkg/api/v1/ridpb"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	aux "github.com/interuss/dss/pkg/aux_"
	"github.com/interuss/dss/pkg/build"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/cockroach/flags" // Force command line flag registration
	uss_errors "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	application "github.com/interuss/dss/pkg/rid/application"
	rid "github.com/interuss/dss/pkg/rid/server"
	ridc "github.com/interuss/dss/pkg/rid/store/cockroach"
	"github.com/interuss/dss/pkg/scd"
	scdc "github.com/interuss/dss/pkg/scd/store/cockroach"
	"github.com/interuss/dss/pkg/validations"
	"github.com/interuss/stacktrace"
	"github.com/robfig/cron/v3"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	enableHTTP        = flag.Bool("enable_http", false, "Enables http scheme for Strategic Conflict Detection API")
	locality          = flag.String("locality", "", "self-identification string used as CRDB table writer column")

	jwtAudiences = flag.String("accepted_jwt_audiences", "", "comma-separated acceptable JWT `aud` claims")
)

func pingDB(ctx context.Context, db *cockroach.DB, databaseName string) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	if err := db.PingContext(ctx); err != nil {
		logger.Panic("Failed periodic DB Ping, panic to force restart", zap.String("Database", databaseName))
	} else {
		logger.Info("Successful periodic DB Ping ", zap.String("Database", databaseName))
	}
}

func createKeyResolver() (auth.KeyResolver, error) {
	switch {
	case *pkFile != "":
		return &auth.FromFileKeyResolver{
			KeyFiles: strings.Split(*pkFile, ","),
		}, nil
	case *jwksEndpoint != "" && *jwksKeyIDs != "":
		u, err := url.Parse(*jwksEndpoint)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error parsing JWKS URL")
		}

		return &auth.JWKSResolver{
			Endpoint: u,
			KeyIDs:   strings.Split(*jwksKeyIDs, ","),
		}, nil
	default:
		return nil, nil
	}
}

func createRIDServer(ctx context.Context, locality string, logger *zap.Logger) (*rid.Server, error) {
	connectParameters := flags.ConnectParameters()
	connectParameters.DBName = ridc.DatabaseName
	ridCrdb, err := cockroach.ConnectTo(connectParameters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to connect to remote ID database; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
	}

	ridStore, err := ridc.NewStore(ctx, ridCrdb, logger)
	if err != nil {
		// try DatabaseName with defaultdb for older versions.
		ridc.DatabaseName = "defaultdb"
		connectParameters.DBName = ridc.DatabaseName
		ridCrdb, err := cockroach.ConnectTo(connectParameters)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to connect to remote ID database for older version <defaultdb>; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
		}
		ridStore, err = ridc.NewStore(ctx, ridCrdb, logger)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to create remote ID store")
		}
	}

	repo, err := ridStore.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}
	gc := ridc.NewGarbageCollector(repo, locality)

	// schedule period tasks for RID Server
	ridCron := cron.New()
	// schedule pinging every minute for the underlying storage for RID Server
	if _, err := ridCron.AddFunc("@every 1m", func() { pingDB(ctx, ridCrdb, ridc.DatabaseName) }); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to schedule periodic ping to %s", ridc.DatabaseName)
	}

	cronLogger := cron.VerbosePrintfLogger(log.New(os.Stdout, "RIDGarbageCollectorJob: ", log.LstdFlags))
	// TODO(supicha): make the 30m configurable
	if _, err = ridCron.AddJob("@every 30m", cron.NewChain(cron.SkipIfStillRunning(cronLogger)).Then(RIDGarbageCollectorJob{"delete rid expired records", *gc, ctx})); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to schedule periodic delete rid expired records to %s", ridc.DatabaseName)
	}
	ridCron.Start()

	return &rid.Server{
		App:        application.NewFromTransactor(ridStore, logger),
		Timeout:    *timeout,
		Locality:   locality,
		EnableHTTP: *enableHTTP,
	}, nil
}

func createSCDServer(ctx context.Context, logger *zap.Logger) (*scd.Server, error) {
	connectParameters := flags.ConnectParameters()
	connectParameters.DBName = scdc.DatabaseName
	scdCrdb, err := cockroach.ConnectTo(connectParameters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to connect to strategic conflict detection database; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
	}
	// schedule period tasks for SCD Server
	scdCron := cron.New()
	// schedule pinging every minute for the underlying storage for SCD Server
	if _, err := scdCron.AddFunc("@every 1m", func() { pingDB(ctx, scdCrdb, scdc.DatabaseName) }); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to schedule periodic ping to %s", scdc.DatabaseName)
	}

	scdCron.Start()

	scdStore, err := scdc.NewStore(ctx, scdCrdb, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create strategic conflict detection store")
	}

	return &scd.Server{
		Store:      scdStore,
		Timeout:    *timeout,
		EnableHTTP: *enableHTTP,
	}, nil
}

// RunGRPCServer starts the example gRPC service.
// "network" and "address" are passed to net.Listen.
func RunGRPCServer(ctx context.Context, ctxCanceler func(), address string, locality string) error {
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
	// l does not need to be closed manually. Instead, the grpc Server instance owning
	// l will close it on a graceful stop.

	var (
		ridServer *rid.Server
		scdServer *scd.Server
		auxServer = &aux.Server{}
	)

	// Initialize remote ID
	server, err := createRIDServer(ctx, locality, logger)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create remote ID server")
	}
	ridServer = server

	scopesValidators := auth.MergeOperationsAndScopesValidators(
		ridServer.AuthScopes(), auxServer.AuthScopes(),
	)

	// Initialize strategic conflict detection

	if *enableSCD {
		server, err := createSCDServer(ctx, logger)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create strategic conflict detection server")
		}
		scdServer = server

		scopesValidators = auth.MergeOperationsAndScopesValidators(
			scopesValidators, scdServer.AuthScopes(),
		)
	}

	// Initialize access token validation
	keyResolver, err := createKeyResolver()
	switch {
	case err != nil:
		return stacktrace.Propagate(err, "Error creating RSA authorizer")
	case keyResolver == nil:
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

	logger.Info("build", zap.Any("description", build.Describe()))

	ridpb.RegisterDiscoveryAndSynchronizationServiceServer(s, ridServer)
	auxpb.RegisterDSSAuxServiceServer(s, auxServer)
	if *enableSCD {
		logger.Info("config", zap.Any("scd", "enabled"))
		scdpb.RegisterUTMAPIUSSDSSAndUSSUSSServiceServer(s, scdServer)
	} else {
		logger.Info("config", zap.Any("scd", "disabled"))
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	go func() {
		defer s.GracefulStop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("stopping server due to context having been canceled")
				return
			case s := <-signals:
				logger.Info("received OS signal", zap.Stringer("signal", s))
				ctxCanceler()
			}
		}
	}()
	return s.Serve(l)
}

type RIDGarbageCollectorJob struct {
	name string
	gc   ridc.GarbageCollector
	ctx  context.Context
}

func (gcj RIDGarbageCollectorJob) Run() {
	logger := logging.WithValuesFromContext(gcj.ctx, logging.Logger)
	err := gcj.gc.DeleteRIDExpiredRecords(gcj.ctx)
	if err != nil {
		logger.Warn("Fail to delete expired records", zap.Error(err))
	} else {
		logger.Info("Successful delete expired records")
	}
}

func main() {
	flag.Parse()

	if err := logging.Configure(*logLevel, *logFormat); err != nil {
		panic(fmt.Sprintf("Failed to configure logging: %s", err.Error()))
	}

	var (
		ctx, cancel = context.WithCancel(context.Background())
		logger      = logging.WithValuesFromContext(ctx, logging.Logger)
	)
	defer cancel()

	if *profServiceName != "" {
		if err := profiler.Start(profiler.Config{
			Service: *profServiceName,
		}); err != nil {
			logger.Panic("Failed to start the profiler ", zap.Error(err))
		}
	}

	if err := RunGRPCServer(ctx, cancel, *address, *locality); err != nil {
		logger.Panic("Failed to execute service", zap.Error(err))
	}

	logger.Info("Shutting down gracefully")
}
