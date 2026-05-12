package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/interuss/dss/pkg/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, metricsListeningAddress string) (func(context.Context) error, error) {

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tracerProvider)

	// Set up metrics exporter
	meterProvider, err := newMeterProvider(ctx, metricsListeningAddress)
	if err != nil {
		return nil, err
	}
	otel.SetMeterProvider(meterProvider)

	shutdown := func(ctx context.Context) error {
		return errors.Join(tracerProvider.Shutdown(ctx), meterProvider.Shutdown(ctx))
	}
	return shutdown, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context) (*trace.TracerProvider, error) {

	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("dss"),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context, listeningAddress string) (*metric.MeterProvider, error) {

	exporter, err := prometheus.New()

	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))

	// Start the prometheus HTTP server
	go serveMetrics(ctx, listeningAddress)

	return provider, nil

}

func serveMetrics(ctx context.Context, listeningAddress string) {

	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(listeningAddress, nil)
	if err != nil {
		logger.Panic("error serving http", zap.Error(err))
		return
	}
	logger.Info("Prometheus endpoint started", zap.String("listeningAddress", listeningAddress))
}
