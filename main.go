package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tracer-test/pkg/health"
	"tracer-test/pkg/help"
	"tracer-test/pkg/httpclient"
	"tracer-test/pkg/logger"
	"tracer-test/pkg/tracer"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	// Version information set during build
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	targetURL     = flag.String("url", "https://httpbin.org/get", "URL to make GET request to")
	otlpEndpoint  = flag.String("otlp-endpoint", "http://localhost:4318", "OTLP endpoint for traces")
	serviceName   = flag.String("service-name", "http-client", "Service name for tracing")
	interval      = flag.Duration("interval", 5*time.Second, "Interval between requests")
	logLevel      = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFormat     = flag.String("log-format", "json", "Log format (json, console)")
	disableOTLP   = flag.Bool("disable-otlp", false, "Disable OTLP tracing export")
	showHelp      = flag.Bool("help", false, "Show help message")
	showVersion   = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Check for help flag
	if *showHelp {
		help.PrintHelp()
		os.Exit(0)
	}

	// Check for version flag
	if *showVersion {
		fmt.Printf("tracer-test version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		os.Exit(0)
	}

	// Initialize logger
	log, err := logger.New(logger.Config{
		Level:  *logLevel,
		Format: *logFormat,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Initialize tracer
	t, err := tracer.New(tracer.Config{
		Endpoint:    *otlpEndpoint,
		ServiceName: *serviceName,
		Disabled:    *disableOTLP,
	}, log.Logger)
	if err != nil {
		log.Error("Failed to initialize tracer", zap.Error(err))
		os.Exit(1)
	}
	defer t.Shutdown(context.Background())

	// Initialize HTTP client
	client := httpclient.New(httpclient.Config{
		Timeout: 10 * time.Second,
	}, log.Logger, t.GetTracer())
	defer client.Close()

	// Initialize health server
	healthServer := health.New(8080)
	healthServer.SetReady(true)

	// Start health server in background
	go func() {
		if err := healthServer.Start(); err != nil {
			log.Error("Health server failed", zap.Error(err))
		}
	}()

	// Log startup information
	log.Info("Starting HTTP client with OTLP tracing",
		zap.String("target_url", *targetURL),
		zap.String("otlp_endpoint", *otlpEndpoint),
		zap.String("service_name", *serviceName),
		zap.Duration("request_interval", *interval),
		zap.String("log_level", *logLevel),
		zap.String("log_format", *logFormat))

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal")
		cancel()
	}()

	// Start request loop
	log.Info("Starting request loop")
	
	requestCount := 0
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Shutting down")
			return
		case <-ticker.C:
			requestCount++
			makeRequest(ctx, client, log, t.GetTracer(), *targetURL, requestCount)
			healthServer.IncrementRequests()
		}
	}
}

func makeRequest(ctx context.Context, client *httpclient.Client, log *logger.Logger, tracer trace.Tracer, url string, requestCount int) {
	// Create root span for the entire request cycle
	ctx, span := tracer.Start(ctx, "request.cycle",
		trace.WithAttributes(
			attribute.String("service.name", *serviceName),
			attribute.String("request.target_url", url),
			attribute.Int64("request.interval_ms", interval.Milliseconds()),
			attribute.Int("request.count", requestCount),
		))
	defer span.End()

	start := time.Now()

	// Make HTTP request
	resp, err := client.Get(ctx, url)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		
		// Log with trace context
		traceCtx := log.WithTraceContext(
			span.SpanContext().TraceID().String(),
			span.SpanContext().SpanID().String(),
		)
		traceCtx.Error("Request failed", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		
		traceCtx := log.WithTraceContext(
			span.SpanContext().TraceID().String(),
			span.SpanContext().SpanID().String(),
		)
		traceCtx.Error("Failed to read response body", zap.Error(err))
		return
	}

	duration := time.Since(start)

	// Set span attributes and status
	span.SetAttributes(
		attribute.Int64("request.cycle.duration_ms", duration.Milliseconds()),
		attribute.Bool("request.success", resp.StatusCode < 400),
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.Int("response.size", len(body)),
	)

	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		span.SetAttributes(attribute.String("request.error", fmt.Sprintf("HTTP %d", resp.StatusCode)))
	} else {
		span.SetStatus(codes.Ok, "")
	}

	// Log with trace context
	traceCtx := log.WithTraceContext(
		span.SpanContext().TraceID().String(),
		span.SpanContext().SpanID().String(),
	)

	if resp.StatusCode >= 400 {
		traceCtx.Warn("HTTP request returned error status",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("response_size", len(body)),
			zap.Duration("duration", duration))
	} else {
		traceCtx.Info("HTTP request completed successfully",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("response_size", len(body)),
			zap.Duration("duration", duration))
	}
}