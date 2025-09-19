package help

import (
	"fmt"
)

// PrintHelp prints the help message
func PrintHelp() {
	fmt.Printf(`HTTP Client with OTLP Tracing

A Go program that makes HTTP GET requests to a configurable URL and sends
distributed traces via OpenTelemetry Protocol (OTLP).

OPTIONS:
    -url string
        URL to make GET request to (default: "https://httpbin.org/get")
    
    -otlp-endpoint string
        OTLP endpoint for traces (default: "http://localhost:4318")
        Examples:
            - http://localhost:4318 (local OTLP collector)
            - https://your-otlp-endpoint.com (external OTLP collector)
            - alloy-test.cel2.celo-networks-dev.org (external domain, auto-detects HTTPS)
    
    -service-name string
        Service name for tracing (default: "http-client")
    
    -interval duration
        Interval between requests (default: "5s")
        Examples: "1s", "30s", "1m", "2h30m"
    
    -log-level string
        Log level (default: "info")
        Options: debug, info, warn, error
    
    -log-format string
        Log format (default: "json")
        Options: json, console
    
    -disable-otlp
        Disable OTLP tracing export (useful for testing without backend)
    
    -help
        Show this help message and exit
    
    -version
        Show version information and exit

EXAMPLES:
    # Basic usage with default settings
    tracer-test

    # Custom URL and interval
    tracer-test -url "https://api.example.com/data" -interval 10s

    # External OTLP endpoint with debug logging
    tracer-test -otlp-endpoint "https://your-otlp-endpoint.com" -log-level debug

    # Disable OTLP tracing for testing
    tracer-test -disable-otlp -log-format console

    # High-frequency requests with custom service name
    tracer-test -url "https://httpbin.org/json" -interval 1s -service-name "load-tester"

FEATURES:
    • HTTP GET requests with configurable intervals
    • OpenTelemetry (OTLP) distributed tracing
    • Structured JSON logging with configurable levels
    • Detailed network instrumentation (DNS, TCP, HTTP)
    • Automatic protocol detection (HTTP/HTTPS)
    • Health check endpoints (/health, /ready, /metrics)
    • Trace correlation in logs (trace ID and span ID)

TRACING:
    The program creates detailed hierarchical spans:
    • request.cycle - Root span for each request cycle
    • http.get - HTTP request span
    • http.transport - Transport layer span
    • dns.resolve - DNS resolution span
    • tcp.connect - TCP connection span

    Each span includes timing, status, and relevant attributes for
    comprehensive observability and debugging.

HEALTH CHECKS:
    The program exposes HTTP endpoints on port 8080:
    • GET /health - Basic health check
    • GET /ready - Readiness check
    • GET /metrics - Simple metrics endpoint

`)
}
