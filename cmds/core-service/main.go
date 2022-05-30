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
	"strconv"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/interuss/dss/pkg/api/v1/auxpb"
	"github.com/interuss/dss/pkg/api/v1/ridpbv1"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	aux "github.com/interuss/dss/pkg/aux_"
	"github.com/interuss/dss/pkg/build"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/cockroach/flags" // Force command line flag registration
	uss_errors "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/logging"
	application "github.com/interuss/dss/pkg/rid/application"
	ridserverv1 "github.com/interuss/dss/pkg/rid/server/v1"
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
	address              = flag.String("addr", ":8081", "address")
	pkFile               = flag.String("public_key_files", "", "Path to public Keys to use for JWT decoding, separated by commas.")
	jwksEndpoint         = flag.String("jwks_endpoint", "", "URL pointing to an endpoint serving JWKS")
	jwksKeyIDs           = flag.String("jwks_key_ids", "", "IDs of a set of key in a JWKS, separated by commas")
	keyRefreshTimeout    = flag.Duration("key_refresh_timeout", 1*time.Minute, "Timeout for refreshing keys for JWT verification")
	timeout              = flag.Duration("server timeout", 10*time.Second, "Default timeout for server calls")
	reflectAPI           = flag.Bool("reflect_api", false, "Whether to reflect the API.")
	logFormat            = flag.String("log_format", logging.DefaultFormat, "The log format in {json, console}")
	logLevel             = flag.String("log_level", logging.DefaultLevel.String(), "The log level")
	dumpRequests         = flag.Bool("dump_requests", false, "Log request and response protos")
	profServiceName      = flag.String("gcp_prof_service_name", "", "Service name for the Go profiler")
	enableSCD            = flag.Bool("enable_scd", false, "Enables the Strategic Conflict Detection API")
	enableHTTP           = flag.Bool("enable_http", false, "Enables http scheme for Strategic Conflict Detection API")
	locality             = flag.String("locality", "", "self-identification string used as CRDB table writer column")
	garbageCollectorSpec = flag.String("garbage_collector_spec", "@every 30m", "Garbage collector schedule. The value must follow robfig/cron format. See https://godoc.org/github.com/robfig/cron#hdr-Usage for more detail.")

	jwtAudiences = flag.String("accepted_jwt_audiences", "", "comma-separated acceptable JWT `aud` claims")
)

const (
	codeRetryable = stacktrace.ErrorCode(1)
)

func getDBStats(ctx context.Context, db *cockroach.DB, databaseName string) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	statsPtr := db.Pool.Stat()
	stats := make(map[string]string)
	stats["DBName"] = databaseName
	stats["AcquireCount"] = strconv.Itoa(int(statsPtr.AcquireCount()))
	stats["AcquiredConns"] = strconv.Itoa(int(statsPtr.AcquiredConns()))
	stats["CanceledAcquireCount"] = strconv.Itoa(int(statsPtr.CanceledAcquireCount()))
	stats["ConstructingConns"] = strconv.Itoa(int(statsPtr.ConstructingConns()))
	stats["EmptyAcquireCount"] = strconv.Itoa(int(statsPtr.EmptyAcquireCount()))
	stats["IdleConns"] = strconv.Itoa(int(statsPtr.IdleConns()))
	stats["MaxConns"] = strconv.Itoa(int(statsPtr.MaxConns()))
	stats["TotalConns"] = strconv.Itoa(int(statsPtr.TotalConns()))
	if stats["TotalConns"] == "0" {
		logger.Warn("Failed periodic DB Ping (TotalConns=0)", zap.String("Database", databaseName))
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

func createRIDServer(ctx context.Context, locality string, logger *zap.Logger) (*ridserverv1.Server, error) {
	connectParameters := flags.ConnectParameters()
	connectParameters.DBName = "rid"
	ridCrdb, err := cockroach.Dial(ctx, connectParameters)
	if err != nil {
		// TODO: More robustly detect failure to create RID server is due to a problem that may be temporary
		if strings.Contains(err.Error(), "connect: connection refused") {
			return nil, stacktrace.PropagateWithCode(err, codeRetryable, "Failed to connect to CRDB server for remote ID store")
		}
		return nil, stacktrace.Propagate(err, "Failed to connect to remote ID database; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
	}

	ridStore, err := ridc.NewStore(ctx, ridCrdb, connectParameters.DBName, logger)
	if err != nil {
		// try DBName of defaultdb for older versions.
		ridCrdb.Pool.Close()
		connectParameters.DBName = "defaultdb"
		ridCrdb, err := cockroach.Dial(ctx, connectParameters)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to connect to remote ID database for older version <defaultdb>; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
		}
		ridStore, err = ridc.NewStore(ctx, ridCrdb, connectParameters.DBName, logger)
		if err != nil {
			// TODO: More robustly detect failure to create RID server is due to a problem that may be temporary
			if strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "database has not been bootstrapped with Schema Manager") {
				ridCrdb.Pool.Close()
				return nil, stacktrace.PropagateWithCode(err, codeRetryable, "Failed to connect to CRDB server for remote ID store")
			}
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
	// schedule printing of DB connection stats every minute for the underlying storage for RID Server
	if _, err := ridCron.AddFunc("@every 1m", func() { getDBStats(ctx, ridCrdb, connectParameters.DBName) }); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to schedule periodic db stat check to %s", connectParameters.DBName)
	}

	cronLogger := cron.VerbosePrintfLogger(log.New(os.Stdout, "RIDGarbageCollectorJob: ", log.LstdFlags))
	if _, err = ridCron.AddJob(*garbageCollectorSpec, cron.NewChain(cron.SkipIfStillRunning(cronLogger)).Then(RIDGarbageCollectorJob{"delete rid expired records", *gc, ctx})); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to schedule periodic delete rid expired records to %s", connectParameters.DBName)
	}
	ridCron.Start()

	return &ridserverv1.Server{
		App:        application.NewFromTransactor(ridStore, logger),
		Timeout:    *timeout,
		Locality:   locality,
		EnableHTTP: *enableHTTP,
		Cron:       ridCron,
	}, nil
}

func createSCDServer(ctx context.Context, logger *zap.Logger) (*scd.Server, error) {
	connectParameters := flags.ConnectParameters()
	connectParameters.DBName = scdc.DatabaseName
	scdCrdb, err := cockroach.Dial(ctx, connectParameters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to connect to strategic conflict detection database; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
	}

	scdStore, err := scdc.NewStore(ctx, scdCrdb, logger)
	if err != nil {
		// TODO: More robustly detect failure to create SCD server is due to a problem that may be temporary
		if strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "database \"scd\" does not exist") {
			scdCrdb.Pool.Close()
			return nil, stacktrace.PropagateWithCode(err, codeRetryable, "Failed to connect to CRDB server for strategic conflict detection store")
		}
		return nil, stacktrace.Propagate(err, "Failed to create strategic conflict detection store")
	}

	// schedule period tasks for SCD Server
	scdCron := cron.New()
	// schedule printing of DB connection stats every minute for the underlying storage for RID Server
	if _, err := scdCron.AddFunc("@every 1m", func() { getDBStats(ctx, scdCrdb, scdc.DatabaseName) }); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to schedule periodic db stat check to %s", scdc.DatabaseName)
	}

	scdCron.Start()

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

	var (
		ridServerV1 *ridserverv1.Server
		scdServer   *scd.Server
		auxServer   = &aux.Server{}
	)

	// Initialize remote ID
	server, err := createRIDServer(ctx, locality, logger)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create remote ID server")
	}
	ridServerV1 = server

	scopesValidators := auth.MergeOperationsAndScopesValidators(
		ridServerV1.AuthScopes(), auxServer.AuthScopes(),
	)

	// Initialize strategic conflict detection

	if *enableSCD {
		server, err := createSCDServer(ctx, logger)
		if err != nil {
			ridServerV1.Cron.Stop()
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

	ridpbv1.RegisterDiscoveryAndSynchronizationServiceServer(s, ridServerV1)
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
	l, err := net.Listen("tcp", address)
	if err != nil {
		return stacktrace.Propagate(err, "Error attempting to listen at %s", address)
	}
	// l does not need to be closed manually. Instead, the grpc Server instance owning
	// l will close it on a graceful stop.

	// Indicate ready for container health checks
	readyFile, err := os.Create("service.ready")
	if err != nil {
		return stacktrace.Propagate(err, "Error touching file to indicate service ready")
	}
	readyFile.Close()

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

	backoffs := []time.Duration{
		5 * time.Second, 15 * time.Second, 1 * time.Minute, 1 * time.Minute,
		1 * time.Minute, 5 * time.Minute}
	backoff := 0
	for {
		if err := RunGRPCServer(ctx, cancel, *address, *locality); err != nil {
			if stacktrace.GetCode(err) == codeRetryable {
				logger.Info(fmt.Sprintf("Prerequisites not yet satisfied; waiting %ds to retry...", backoffs[backoff]/1000000000), zap.Error(err))
				time.Sleep(backoffs[backoff])
				if backoff < len(backoffs)-1 {
					backoff++
				}
				continue
			}
			logger.Panic("Failed to execute service", zap.Error(err))
		}
		break
	}

	logger.Info("Shutting down gracefully")
}
