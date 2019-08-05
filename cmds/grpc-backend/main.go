package main

import (
	"context"
	"flag"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/steeling/InterUSS-Platform/pkg/dss"
	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	"github.com/steeling/InterUSS-Platform/pkg/dssproto"
	"github.com/steeling/InterUSS-Platform/pkg/logging"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

var (
	address = flag.String("addr", "127.0.0.1:8080", "address")
	pkFile  = flag.String("public_key_file", "", "Path to public Key to use for JWT decoding.")
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

	ac, err := auth.NewRSAAuthClient(*pkFile)
	if err != nil {
		return err
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(ac.AuthInterceptor)))
	if err != nil {
		return err
	}
	dssproto.RegisterDiscoveryAndSynchronizationServiceServer(s, &dss.Server{
		Store:   dss.NewNilStore(),
		Coverer: geo.DefaultRegionCoverer,
	})

	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()
	return s.Serve(l)
}

func main() {
	flag.Parse()
	var (
		ctx    = context.Background()
		logger = logging.WithValuesFromContext(ctx, logging.Logger)
	)

	if err := RunGRPCServer(ctx, *address); err != nil {
		logger.Panic("Failed to execute service", zap.Error(err))
	}
	logger.Info("Shutting down gracefully")
}
