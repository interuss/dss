package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/interuss/dss/pkg/dss/build"
	"github.com/interuss/dss/pkg/dssproto"
	"github.com/interuss/dss/pkg/logging"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	address       = flag.String("addr", ":8080", "Local address that the gateway binds to and listens on for incoming connections")
	traceRequests = flag.Bool("trace-requests", false, "Logs HTTP request/response pairs to stderr if true")
	grpcBackend   = flag.String("grpc-backend", "", "Endpoint for grpc backend. Only to be set if run in proxy mode")
)

// RunHTTPProxy starts the HTTP proxy for the DSS gRPC service on ctx, listening
// on address, proxying to endpoint.
func RunHTTPProxy(ctx context.Context, address, endpoint string) error {
	logger := logging.WithValuesFromContext(ctx, logging.Logger).With(
		zap.String("address", address), zap.String("endpoint", endpoint),
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true, // Include empty JSON arrays.
			Indent:       "  ",
		}),
	)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(10 * time.Second),
	}

	err := dssproto.RegisterDiscoveryAndSynchronizationServiceHandlerFromEndpoint(ctx, grpcMux, endpoint, opts)
	if err != nil {
		return err
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthy" {
			w.Write([]byte("ok"))
		}
		grpcMux.ServeHTTP(w, r)
	})

	if *traceRequests {
		handler = logging.HTTPMiddleware(logger, handler)
	}

	logger.Info("build", zap.Any("description", build.Describe()))

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(address, handler)
}

func main() {
	flag.Parse()
	var (
		ctx    = context.Background()
		logger = logging.WithValuesFromContext(ctx, logging.Logger)
	)

	if err := RunHTTPProxy(ctx, *address, *grpcBackend); err != nil {
		logger.Panic("Failed to execute service", zap.Error(err))
	}
	logger.Info("Shutting down gracefully")
}
