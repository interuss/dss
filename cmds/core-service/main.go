package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/interuss/dss/pkg/api"
	apiauxv1 "github.com/interuss/dss/pkg/api/auxv1"
	apiridv1 "github.com/interuss/dss/pkg/api/ridv1"
	apiridv2 "github.com/interuss/dss/pkg/api/ridv2"
	apiscdv1 "github.com/interuss/dss/pkg/api/scdv1"
	apiversioningv1 "github.com/interuss/dss/pkg/api/versioningv1"
	"github.com/interuss/dss/pkg/auth"
	aux "github.com/interuss/dss/pkg/aux_"
	"github.com/interuss/dss/pkg/build"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/cockroach/flags" // Force command line flag registration
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/application"
	rid_v1 "github.com/interuss/dss/pkg/rid/server/v1"
	rid_v2 "github.com/interuss/dss/pkg/rid/server/v2"
	ridc "github.com/interuss/dss/pkg/rid/store/cockroach"
	"github.com/interuss/dss/pkg/scd"
	scdc "github.com/interuss/dss/pkg/scd/store/cockroach"
	"github.com/interuss/dss/pkg/version"
	"github.com/interuss/dss/pkg/versioning"
	"github.com/interuss/stacktrace"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var (
	address           = flag.String("addr", ":8080", "Local address that the service binds to and listens on for incoming connections")
	enableSCD         = flag.Bool("enable_scd", false, "Enables the Strategic Conflict Detection API")
	allowHTTPBaseUrls = flag.Bool("allow_http_base_urls", false, "Enables http scheme for Strategic Conflict Detection API")
	timeout           = flag.Duration("server timeout", 10*time.Second, "Default timeout for server calls")
	locality          = flag.String("locality", "", "self-identification string used as CRDB table writer column")

	logFormat            = flag.String("log_format", logging.DefaultFormat, "The log format in {json, console}")
	logLevel             = flag.String("log_level", logging.DefaultLevel.String(), "The log level")
	dumpRequests         = flag.Bool("dump_requests", false, "Log HTTP request and response")
	profServiceName      = flag.String("gcp_prof_service_name", "", "Service name for the Go profiler")
	garbageCollectorSpec = flag.String("garbage_collector_spec", "@every 30m", "Garbage collector schedule. The value must follow robfig/cron format. See https://godoc.org/github.com/robfig/cron#hdr-Usage for more detail.")

	pkFile            = flag.String("public_key_files", "", "Path to public Keys to use for JWT decoding, separated by commas.")
	jwksEndpoint      = flag.String("jwks_endpoint", "", "URL pointing to an endpoint serving JWKS")
	jwksKeyIDs        = flag.String("jwks_key_ids", "", "IDs of a set of key in a JWKS, separated by commas")
	keyRefreshTimeout = flag.Duration("key_refresh_timeout", 1*time.Minute, "Timeout for refreshing keys for JWT verification")
	jwtAudiences      = flag.String("accepted_jwt_audiences", "", "comma-separated acceptable JWT `aud` claims")
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

func createRIDServers(ctx context.Context, locality string, logger *zap.Logger) (*rid_v1.Server, *rid_v2.Server, error) {
	connectParameters := flags.ConnectParameters()
	connectParameters.DBName = "rid"
	ridCrdb, err := cockroach.Dial(ctx, connectParameters)
	if err != nil {
		// TODO: More robustly detect failure to create RID server is due to a problem that may be temporary
		if strings.Contains(err.Error(), "connect: connection refused") {
			return nil, nil, stacktrace.PropagateWithCode(err, codeRetryable, "Failed to connect to CRDB server for remote ID store")
		}
		return nil, nil, stacktrace.Propagate(err, "Failed to connect to remote ID database; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
	}

	ridStore, err := ridc.NewStore(ctx, ridCrdb, connectParameters.DBName, logger)
	if err != nil {
		// try DBName of defaultdb for older versions.
		ridCrdb.Pool.Close()
		connectParameters.DBName = "defaultdb"
		ridCrdb, err := cockroach.Dial(ctx, connectParameters)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Failed to connect to remote ID database for older version <defaultdb>; verify your database configuration is current with https://github.com/interuss/dss/tree/master/build#upgrading-database-schemas")
		}
		ridStore, err = ridc.NewStore(ctx, ridCrdb, connectParameters.DBName, logger)
		if err != nil {
			// TODO: More robustly detect failure to create RID server is due to a problem that may be temporary
			if strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "database has not been bootstrapped with Schema Manager") {
				ridCrdb.Pool.Close()
				return nil, nil, stacktrace.PropagateWithCode(err, codeRetryable, "Failed to connect to CRDB server for remote ID store")
			}
			return nil, nil, stacktrace.Propagate(err, "Failed to create remote ID store")
		}
	}

	repo, err := ridStore.Interact(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unable to interact with store")
	}
	gc := ridc.NewGarbageCollector(repo, locality)

	// schedule period tasks for RID Server
	ridCron := cron.New()
	// schedule printing of DB connection stats every minute for the underlying storage for RID Server
	if _, err := ridCron.AddFunc("@every 1m", func() { getDBStats(ctx, ridCrdb, connectParameters.DBName) }); err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to schedule periodic db stat check to %s", connectParameters.DBName)
	}

	cronLogger := cron.VerbosePrintfLogger(log.New(os.Stdout, "RIDGarbageCollectorJob: ", log.LstdFlags))
	if _, err = ridCron.AddJob(*garbageCollectorSpec, cron.NewChain(cron.SkipIfStillRunning(cronLogger)).Then(RIDGarbageCollectorJob{"delete rid expired records", *gc, ctx})); err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to schedule periodic delete rid expired records to %s", connectParameters.DBName)
	}
	ridCron.Start()

	app := application.NewFromTransactor(ridStore, logger)
	return &rid_v1.Server{
			App:               app,
			Timeout:           *timeout,
			Locality:          locality,
			AllowHTTPBaseUrls: *allowHTTPBaseUrls,
			Cron:              ridCron,
		}, &rid_v2.Server{
			App:               app,
			Timeout:           *timeout,
			Locality:          locality,
			AllowHTTPBaseUrls: *allowHTTPBaseUrls,
			Cron:              ridCron,
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
		Store:             scdStore,
		DSSReportHandler:  &scd.JSONLoggingReceivedReportHandler{ReportLogger: logger},
		Timeout:           *timeout,
		AllowHTTPBaseUrls: *allowHTTPBaseUrls,
	}, nil
}

// RunHTTPServer starts the DSS HTTP server.
func RunHTTPServer(ctx context.Context, ctxCanceler func(), address, locality string) error {
	logger := logging.WithValuesFromContext(ctx, logging.Logger).With(zap.String("address", address))
	logger.Info("version", zap.Any("version", version.Current()))
	logger.Info("build", zap.Any("description", build.Describe()))
	logger.Info("config", zap.Bool("scd", *enableSCD))

	if len(*jwtAudiences) == 0 {
		// TODO: Make this flag required once all parties can set audiences
		// correctly.
		logger.Warn("missing required --accepted_jwt_audiences")
	}

	var (
		err                error
		ridV1Server        *rid_v1.Server
		ridV2Server        *rid_v2.Server
		scdV1Server        *scd.Server
		auxV1Server        = &aux.Server{}
		versioningV1Server = &versioning.Server{}
	)

	// Initialize remote ID
	ridV1Server, ridV2Server, err = createRIDServers(ctx, locality, logger)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create remote ID server")
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
			AcceptedAudiences: strings.Split(*jwtAudiences, ","),
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating RSA authorizer")
	}

	auxV1Router := apiauxv1.MakeAPIRouter(auxV1Server, authorizer)
	versioningV1Router := apiversioningv1.MakeAPIRouter(versioningV1Server, authorizer)
	ridV1Router := apiridv1.MakeAPIRouter(ridV1Server, authorizer)
	ridV2Router := apiridv2.MakeAPIRouter(ridV2Server, authorizer)
	multiRouter := api.MultiRouter{
		Routers: []api.PartialRouter{
			&auxV1Router,
			&versioningV1Router,
			&ridV1Router,
			&ridV2Router,
		}}

	// Initialize strategic conflict detection
	if *enableSCD {
		scdV1Server, err = createSCDServer(ctx, logger)
		if err != nil {
			ridV1Server.Cron.Stop()
			ridV2Server.Cron.Stop()
			return stacktrace.Propagate(err, "Failed to create strategic conflict detection server")
		}

		scdV1Router := apiscdv1.MakeAPIRouter(scdV1Server, authorizer)
		multiRouter.Routers = append(multiRouter.Routers, &scdV1Router)
	}

	handler := logging.HTTPMiddleware(logger, *dumpRequests,
		healthyEndpointMiddleware(logger,
			&multiRouter,
		))

	httpServer := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	go func() {
		defer func() {
			if err := httpServer.Shutdown(context.Background()); err != nil {
				logger.Warn("failed to shut down http server", zap.Error(err))
			}
		}()

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

	// Indicate ready for container health checks
	readyFile, err := os.Create("service.ready")
	if err != nil {
		return stacktrace.Propagate(err, "Error touching file to indicate service ready")
	}

	err = readyFile.Close()
	if err != nil {
		return stacktrace.Propagate(err, "Error closing touched file to indicate service ready")
	}

	logger.Info("Starting DSS HTTP server")
	return httpServer.ListenAndServe()
}

// healthyEndpointMiddleware intercepts a request and responds with an "ok" message at the endpoint "/healthy".
func healthyEndpointMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthy" {
			if _, err := w.Write([]byte("ok")); err != nil {
				logger.Error("Error writing to /healthy")
			}
		} else {
			next.ServeHTTP(w, r)
		}
	})
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
		if err := profiler.Start(profiler.Config{Service: *profServiceName}); err != nil {
			logger.Panic("Failed to start the profiler ", zap.Error(err))
		}
	}

	backoffs := []time.Duration{
		5 * time.Second, 15 * time.Second, 1 * time.Minute, 1 * time.Minute,
		1 * time.Minute, 5 * time.Minute}
	backoff := 0
	for {
		if err := RunHTTPServer(ctx, cancel, *address, *locality); err != nil {
			if stacktrace.GetCode(err) == codeRetryable {
				logger.Info(fmt.Sprintf("Prerequisites not yet satisfied; waiting %.fs to retry...", backoffs[backoff].Seconds()), zap.Error(err))
				time.Sleep(backoffs[backoff])
				if backoff < len(backoffs)-1 {
					backoff++
				}
				continue
			}

			rootCause := stacktrace.RootCause(err)
			if rootCause == nil || rootCause == context.Canceled || rootCause == http.ErrServerClosed {
				logger.Info("Shutting down gracefully")
				break
			}
			logger.Panic("Failed to execute service", zap.Error(err))
		}
		break
	}
}
