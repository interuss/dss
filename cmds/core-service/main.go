package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
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
	auxs "github.com/interuss/dss/pkg/aux_/store"
	"github.com/interuss/dss/pkg/build"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/application"
	rid_v1 "github.com/interuss/dss/pkg/rid/server/v1"
	rid_v2 "github.com/interuss/dss/pkg/rid/server/v2"
	rids "github.com/interuss/dss/pkg/rid/store"
	"github.com/interuss/dss/pkg/scd"
	scds "github.com/interuss/dss/pkg/scd/store"
	"github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/store/params"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/dss/pkg/version"
	"github.com/interuss/dss/pkg/versioning"
	"github.com/interuss/stacktrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var (
	address           = flag.String("addr", ":8080", "Address and port that the service binds to and listens on for incoming connections")
	enableSCD         = flag.Bool("enable_scd", false, "Enables the Strategic Conflict Detection API")
	allowHTTPBaseUrls = flag.Bool("allow_http_base_urls", false, "Enables http scheme for Strategic Conflict Detection API")
	enableHTTP        = flag.Bool("enable_http", false, "DEPRECATED (replaced by allow_http_base_urls): Enables http scheme for Strategic Conflict Detection API")
	legacyTimeout     = flag.Duration("server timeout", 10*time.Second, "DEPRECATED (replaced by server_timeout) Default timeout for server calls")
	timeout           = flag.Duration("server_timeout", 10*time.Second, "Default timeout for server calls")
	locality          = flag.String("locality", "", "self-identification string of this DSS instance")
	publicEndpoint    = flag.String("public_endpoint", "", "Public endpoint to access this DSS instance. Must be an absolute URI")

	logFormat               = flag.String("log_format", logging.DefaultFormat, "The log format in {json, console}")
	logLevel                = flag.String("log_level", logging.DefaultLevel.String(), "The log level")
	dumpRequests            = flag.Bool("dump_requests", false, "Log full HTTP request and response (note: will dump sensitive information to logs; intended only for debugging and/or development)")
	profServiceName         = flag.String("gcp_prof_service_name", "", "Service name for the Go profiler")
	enableOpenTelemetry     = flag.Bool("enable_opentelemetry", false, "DEPRECATED (replaced by enable_tracing) Enable tracing")
	enableMetrics           = flag.Bool("enable_metrics", false, "Enable metric endpoint")
	enableTracing           = flag.Bool("enable_tracing", false, "Enable tracing")
	metricsListeningAddress = flag.String("metrics_addr", ":8079", "Address and port that the for the prometheus-compatible metric service binds to and listens on for incoming connections")

	pkFile            = flag.String("public_key_files", "", "Path to public Keys to use for JWT decoding, separated by commas.")
	jwksEndpoint      = flag.String("jwks_endpoint", "", "URL pointing to an endpoint serving JWKS")
	jwksKeyIDs        = flag.String("jwks_key_ids", "", "IDs of a set of key in a JWKS, separated by commas")
	keyRefreshTimeout = flag.Duration("key_refresh_timeout", 1*time.Minute, "Timeout for refreshing keys for JWT verification")
	jwtAudiences      = flag.String("accepted_jwt_audiences", "", "comma-separated acceptable JWT `aud` claims")

	scdGlobalLock = flag.Bool("enable_scd_global_lock", false, "Experimental: Use a global lock when working with SCD subscriptions. Reduce global throughput but improve throughput with lot of subscriptions in the same areas.")
)

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

func createAuxServer(ctx context.Context, locality string, publicEndpoint string, scdGlobalLock bool, logger *zap.Logger) (*aux.Server, error) {
	auxStore, err := auxs.Init(ctx, logger, true)
	if err != nil {
		return nil, err
	}

	repo, err := auxStore.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}

	err = repo.SaveOwnMetadata(ctx, locality, publicEndpoint)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to store current metadata")
	}

	return &aux.Server{Store: auxStore, Locality: locality, ScdGlobalLock: scdGlobalLock}, nil
}

func createRIDServers(ctx context.Context, locality string, logger *zap.Logger) (*rid_v1.Server, *rid_v2.Server, error) {

	ridStore, err := rids.Init(ctx, logger, true)
	if err != nil {
		return nil, nil, err
	}

	_, err = ridStore.Interact(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unable to interact with store")
	}

	if *enableMetrics {
		err = registerRIDMetrics(ctx, ridStore)

		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Unable to setup metrics")
		}
	}

	app := application.NewFromTransactor(ridStore, logger)
	return &rid_v1.Server{
			App:               app,
			Locality:          locality,
			AllowHTTPBaseUrls: *allowHTTPBaseUrls,
		}, &rid_v2.Server{
			App:               app,
			Locality:          locality,
			AllowHTTPBaseUrls: *allowHTTPBaseUrls,
		}, nil
}

func createSCDServer(ctx context.Context, logger *zap.Logger) (*scd.Server, error) {

	scdStore, err := scds.Init(ctx, logger, true, *scdGlobalLock)
	if err != nil {
		return nil, err
	}

	if *enableMetrics {
		err = registerSCDMetrics(ctx, scdStore)

		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to setup metrics")
		}
	}

	return &scd.Server{
		Store:             scdStore,
		DSSReportHandler:  &scd.JSONLoggingReceivedReportHandler{ReportLogger: logger},
		AllowHTTPBaseUrls: *allowHTTPBaseUrls,
	}, nil
}

func registerRIDMetrics(ctx context.Context, store rids.Store) error {

	meter := otel.Meter("rid")

	_, err := meter.Int64ObservableUpDownCounter(
		"rid_subscriptions_total",
		metric.WithDescription("Number of rid subscriptions"),
		metric.WithInt64Callback(newCachedObservation(func(ctx context.Context) (int64, error) {
			repo, err := store.Interact(ctx)
			if err != nil {
				return 0, stacktrace.Propagate(err, "Unable to interact with store")
			}
			count, err := repo.CountSubscriptions(ctx)
			return count, err
		})),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"rid_identification_service_areas_total",
		metric.WithDescription("Number of rid ISAs"),
		metric.WithInt64Callback(newCachedObservation(func(ctx context.Context) (int64, error) {
			repo, err := store.Interact(ctx)
			if err != nil {
				return 0, stacktrace.Propagate(err, "Unable to interact with store")
			}
			count, err := repo.CountISAs(ctx)
			return count, err
		})),
	)

	return err
}

func registerSCDMetrics(ctx context.Context, store scds.Store) error {

	meter := otel.Meter("scd")

	_, err := meter.Int64ObservableUpDownCounter(
		"scd_subscriptions_total",
		metric.WithDescription("Number of scd subscriptions"),
		metric.WithInt64Callback(newCachedObservation(func(ctx context.Context) (int64, error) {
			repo, err := store.Interact(ctx)
			if err != nil {
				return 0, stacktrace.Propagate(err, "Unable to interact with store")
			}
			count, err := repo.CountSubscriptions(ctx)
			return count, err
		})),
	)
	if err != nil {
		return err
	}
	_, err = meter.Int64ObservableUpDownCounter(
		"scd_operational_intents_total",
		metric.WithDescription("Number of scd operational intents"),
		metric.WithInt64Callback(newCachedObservation(func(ctx context.Context) (int64, error) {
			repo, err := store.Interact(ctx)
			if err != nil {
				return 0, stacktrace.Propagate(err, "Unable to interact with store")
			}
			count, err := repo.CountOperationalIntents(ctx)
			return count, err
		})),
	)
	if err != nil {
		return err
	}
	_, err = meter.Int64ObservableUpDownCounter(
		"scd_constraints_total",
		metric.WithDescription("Number of scd constraints"),
		metric.WithInt64Callback(newCachedObservation(func(ctx context.Context) (int64, error) {
			repo, err := store.Interact(ctx)
			if err != nil {
				return 0, stacktrace.Propagate(err, "Unable to interact with store")
			}
			count, err := repo.CountConstraints(ctx)
			return count, err
		})),
	)

	return err
}

// RunHTTPServer starts the DSS HTTP server.
func RunHTTPServer(ctx context.Context, ctxCanceler func(), address, locality string) error {
	logger := logging.WithValuesFromContext(ctx, logging.Logger).With(zap.String("address", address))
	logger.Info("version", zap.Any("version", version.Current()))
	logger.Info("build", zap.Any("description", build.Describe()))
	logger.Info("config", zap.Bool("scd", *enableSCD))
	logger.Info("config", zap.Bool("scdGlobalLock", *scdGlobalLock))
	// params.StoreParameters should not be used directly in this file but this log warning is temporarily helpful and will be removed in the future.
	if params.GetStoreParameters().StoreType == params.RaftStoreType {
		logger.Warn("The raft datastore is experimental and its implementation is in progress. See issue for more details: https://github.com/interuss/dss/issues/1463")
	}

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
		auxV1Server        *aux.Server
		versioningV1Server = &versioning.Server{}
	)

	ctx, ctxCancel := context.WithCancel(ctx)
	defer ctxCancel()

	// Initialize aux
	auxV1Server, err = createAuxServer(ctx, locality, *publicEndpoint, *scdGlobalLock, logger)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create aux server")
	}

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
			return stacktrace.Propagate(err, "Failed to create strategic conflict detection server")
		}

		scdV1Router := apiscdv1.MakeAPIRouter(scdV1Server, authorizer)
		multiRouter.Routers = append(multiRouter.Routers, &scdV1Router)
	}

	// the middlewares are wrapped and, therefore, executed in the opposite order
	handler := healthyEndpointMiddleware(logger, &multiRouter)
	handler = authorizer.TokenMiddleware(handler)
	handler = http.TimeoutHandler(handler, *timeout, "request timeout")
	handler = logging.HTTPMiddleware(logger, *dumpRequests, handler)
	handler = timestamp.Middleware(handler)

	if *enableMetrics || *enableTracing {
		// We use the default settings; the APIRouter handler will override the span value accordingly, as it has more information.
		handler = otelhttp.NewHandler(handler, "http")
	}

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

func SetDeprecatingHttpFlag(logger *zap.Logger, newFlag **bool, deprecatedFlag **bool) {
	if **deprecatedFlag {
		logger.Warn("DEPRECATED: enable_http has been renamed to allow_http_base_urls.")
		if !**newFlag {
			*newFlag = *deprecatedFlag
		}
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

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "server timeout" {
			*timeout = *legacyTimeout
			logger.Warn("'server timeout' has been renamed to 'server_timeout'")
		}
	})

	SetDeprecatingHttpFlag(logger, &allowHTTPBaseUrls, &enableHTTP)

	if *enableOpenTelemetry {
		logger.Warn("'enable_opentelemetry' has been renamed to 'enable_tracing")
		*enableTracing = true
	}

	if *profServiceName != "" {
		if err := profiler.Start(profiler.Config{Service: *profServiceName}); err != nil {
			logger.Panic("Failed to start the profiler ", zap.Error(err))
		}
	}

	// Set up OpenTelemetry.
	if *enableMetrics || *enableTracing {
		otelShutdown, err := setupOTelSDK(ctx, *enableMetrics, *enableTracing, *metricsListeningAddress)
		if err != nil {
			logger.Panic("Failed to initialize OpenTelemetry", zap.Error(err))
		}
		// Handle shutdown properly so nothing leaks.
		defer otelShutdown(context.Background())
	}

	backoffs := []time.Duration{
		5 * time.Second, 15 * time.Second, 1 * time.Minute, 1 * time.Minute,
		1 * time.Minute, 5 * time.Minute}
	backoff := 0
	for {
		if err := RunHTTPServer(ctx, cancel, *address, *locality); err != nil {
			if stacktrace.GetCode(err) == store.CodeRetryable {
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
