package tracer

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Config holds tracer configuration
type Config struct {
	Endpoint    string
	ServiceName string
	Disabled    bool
}

// Tracer wraps the OpenTelemetry tracer
type Tracer struct {
	tracer trace.Tracer
	logger *zap.Logger
}

// New creates a new tracer instance
func New(config Config, logger *zap.Logger) (*Tracer, error) {
	if config.Disabled {
		logger.Info("OTLP tracing disabled - using no-op tracer")
		// Return a no-op tracer
		noopTracer := otel.Tracer("noop")
		return &Tracer{
			tracer: noopTracer,
			logger: logger,
		}, nil
	}

	logger.Info("Initializing OTLP tracer",
		zap.String("otlp_endpoint", config.Endpoint),
		zap.String("service_name", config.ServiceName))

	// Parse the endpoint URL to determine if we should use insecure connection
	useInsecure := shouldUseInsecure(config.Endpoint)

	// Clean the endpoint URL (remove http:// or https:// prefix)
	cleanEndpoint := cleanEndpointURL(config.Endpoint)

	// Build exporter options
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cleanEndpoint),
		otlptracehttp.WithURLPath("/v1/traces"),
	}

	// Add insecure option if needed
	if useInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Create tracer
	tracer := tp.Tracer(config.ServiceName)

	logger.Info("OTLP tracer initialized successfully")

	return &Tracer{
		tracer: tracer,
		logger: logger,
	}, nil
}

// GetTracer returns the underlying tracer
func (t *Tracer) GetTracer() trace.Tracer {
	return t.tracer
}

// shouldUseInsecure determines if we should use insecure connection based on the endpoint
func shouldUseInsecure(endpoint string) bool {
	// If endpoint starts with https://, use secure connection
	if strings.HasPrefix(endpoint, "https://") {
		return false
	}

	// If endpoint starts with http://, use insecure connection
	if strings.HasPrefix(endpoint, "http://") {
		return true
	}

	// For endpoints without protocol prefix, determine based on hostname
	// If it's localhost or 127.0.0.1, use insecure (HTTP)
	// Otherwise, use secure (HTTPS)
	if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") {
		return true
	}

	// For external domains, use secure connection
	return false
}

// cleanEndpointURL removes the protocol prefix from the endpoint URL
func cleanEndpointURL(endpoint string) string {
	// Remove http:// or https:// prefix if present
	if strings.HasPrefix(endpoint, "http://") {
		return strings.TrimPrefix(endpoint, "http://")
	}
	if strings.HasPrefix(endpoint, "https://") {
		return strings.TrimPrefix(endpoint, "https://")
	}
	return endpoint
}

// Shutdown gracefully shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}
