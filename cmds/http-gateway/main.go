package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dss/build"
	"github.com/steeling/InterUSS-Platform/pkg/dssv1"
	"github.com/steeling/InterUSS-Platform/pkg/logging"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

var (
	address     = flag.String("addr", ":8080", "address")
	grpcBackend = flag.String("grpc-backend", "", "Endpoint for grpc backend. Only to be set if run in proxy mode")
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

	err := dssv1.RegisterDiscoveryAndSynchronizationServiceHandlerFromEndpoint(ctx, grpcMux, endpoint, opts)
	if err != nil {
		return err
	}

	// Register a health check handler.
	m := mux.NewRouter()
	m.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Let grpcMux handle everything else.
	m.NotFoundHandler = grpcMux

	logger.Info("build", zap.Any("description", build.Describe()))

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(address, m)
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
