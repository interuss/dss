package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"InterUSS-Platform/src/backend"
	"InterUSS-Platform/src/dssproto"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

var (
	address      = flag.String("addr", "127.0.0.1:8080", "address")
	grpc_backend = flag.String("grpc-backend", "", "Endpoint for grpc backend. Only to be set if run in proxy mode")
	mode         = flag.String("mode", "", "One of [backend, proxy].")
)

// Run starts the example gRPC service.
// "network" and "address" are passed to net.Listen.
func RunGRPCServer(ctx context.Context, address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer func() {
		if err := l.Close(); err != nil {
			glog.Errorf("Failed to close %s: %v", address, err)
		}
	}()

	s := grpc.NewServer()
	dss, err := backend.NewServer()
	if err != nil {
		return err
	}
	dssproto.RegisterDSSServiceServer(s, dss)

	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()
	return s.Serve(l)
}

func RunHTTPProxy(ctx context.Context, address, endpoint string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(10 * time.Second),
	}

	err := dssproto.RegisterDSSServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(address, mux)
}

func main() {
	flag.Parse()
	var err error

	switch *mode {
	case "backend":
		err = RunGRPCServer(context.Background(), *address)
	case "proxy":
		err = RunHTTPProxy(context.Background(), *address, *grpc_backend)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
	if err != nil {
		panic(err)
	}
	log.Print("Shutting down gracefully")
}
