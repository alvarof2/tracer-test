package httpclient

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Client wraps the HTTP client with tracing
type Client struct {
	httpClient *http.Client
	logger     *zap.Logger
	tracer     trace.Tracer
}

// Config holds HTTP client configuration
type Config struct {
	Timeout time.Duration
}

// New creates a new HTTP client with tracing
func New(config Config, logger *zap.Logger, tracer trace.Tracer) *Client {
	// Create instrumented transport
	transport := &instrumentedTransport{
		base:   http.DefaultTransport,
		logger: logger,
		tracer: tracer,
	}

	// Create HTTP client with custom transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &Client{
		httpClient: httpClient,
		logger:     logger,
		tracer:     tracer,
	}
}

// Get makes a GET request with tracing
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	// Create span for HTTP request
	ctx, span := c.tracer.Start(ctx, "http.get",
		trace.WithAttributes(
			attribute.String("http.method", "GET"),
			attribute.String("http.url", url),
		))
	defer span.End()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		c.logger.Error("HTTP request failed",
			zap.String("url", url),
			zap.Error(err),
			zap.Duration("duration", time.Since(time.Now())))
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Set span attributes based on response
	span.SetAttributes(
		semconv.HTTPResponseStatusCode(resp.StatusCode),
		semconv.HTTPResponseSize(int(resp.ContentLength)),
	)

	// Set span status based on HTTP status code
	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		c.logger.Warn("HTTP request returned error status",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
			zap.Int64("response_size", resp.ContentLength))
	} else {
		span.SetStatus(codes.Ok, "")
		c.logger.Info("HTTP request completed successfully",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
			zap.Int64("response_size", resp.ContentLength))
	}

	return resp, nil
}

// instrumentedTransport wraps http.RoundTripper with detailed instrumentation
type instrumentedTransport struct {
	base   http.RoundTripper
	logger *zap.Logger
	tracer trace.Tracer
}

// RoundTrip implements http.RoundTripper interface
func (t *instrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create span for HTTP transport
	ctx, span := t.tracer.Start(req.Context(), "http.transport",
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
		))
	defer span.End()

	// Update request context
	req = req.WithContext(ctx)

	// Perform DNS resolution
	host := req.URL.Hostname()
	port := req.URL.Port()
	if port == "" {
		if req.URL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// DNS resolution span
	_, dnsSpan := t.tracer.Start(ctx, "dns.resolve",
		trace.WithAttributes(
			attribute.String("dns.hostname", host),
		))
	
	start := time.Now()
	ips, err := net.LookupIP(host)
	dnsDuration := time.Since(start)
	
	if err != nil {
		dnsSpan.RecordError(err)
		dnsSpan.SetStatus(codes.Error, err.Error())
	} else {
		dnsSpan.SetAttributes(
			attribute.StringSlice("dns.addresses", ipToStrings(ips)),
			attribute.Int64("dns.duration_ms", dnsDuration.Milliseconds()),
		)
		dnsSpan.SetStatus(codes.Ok, "")
	}
	dnsSpan.End()

	// TCP connection span
	_, tcpSpan := t.tracer.Start(ctx, "tcp.connect",
		trace.WithAttributes(
			attribute.String("net.peer.name", host),
			attribute.String("net.peer.port", port),
		))

	// Make the actual HTTP request
	start = time.Now()
	resp, err := t.base.RoundTrip(req)
	httpDuration := time.Since(start)

	if err != nil {
		tcpSpan.RecordError(err)
		tcpSpan.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		tcpSpan.SetAttributes(
			attribute.Int64("tcp.duration_ms", httpDuration.Milliseconds()),
		)
		tcpSpan.SetStatus(codes.Ok, "")
		
		span.SetAttributes(
			semconv.HTTPResponseStatusCode(resp.StatusCode),
			semconv.HTTPResponseSize(int(resp.ContentLength)),
			attribute.Int64("http.duration_ms", httpDuration.Milliseconds()),
		)
		
		if resp.StatusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
	tcpSpan.End()

	return resp, err
}

// ipToStrings converts []net.IP to []string
func ipToStrings(ips []net.IP) []string {
	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}
	return result
}

// Close closes the HTTP client
func (c *Client) Close() {
	// Close any idle connections
	c.httpClient.CloseIdleConnections()
}
