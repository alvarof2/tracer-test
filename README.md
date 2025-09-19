# HTTP Client with OTLP Tracing

A simple Golang program that makes HTTP GET requests to a configurable URL with OpenTelemetry (OTLP) tracing instrumentation.

## Features

- Configurable target URL for HTTP requests
- OTLP tracing with configurable endpoint
- Automatic trace context propagation
- Request timing and status tracking
- Continuous operation with configurable intervals
- Rich span attributes for observability
- Structured JSON logging with configurable log levels
- Trace correlation in logs (trace ID and span ID)

## Prerequisites

- Go 1.21 or later
- An OTLP-compatible tracing backend (e.g., Jaeger, Zipkin, or cloud providers)

## Installation

1. Clone or download this repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

### Basic Usage

```bash
go run main.go
```

This will make GET requests to `https://httpbin.org/get` every 5 seconds and send traces to `http://localhost:4318`.

### Command Line Options

- `-url`: Target URL for HTTP requests (default: `https://httpbin.org/get`)
- `-otlp-endpoint`: OTLP endpoint for traces (default: `http://localhost:4318`)
- `-service-name`: Service name for tracing (default: `http-client`)
- `-interval`: Interval between requests (default: `5s`)
- `-log-level`: Log level (debug, info, warn, error) (default: `info`)
- `-log-format`: Log format (json, console) (default: `json`)

### Examples

```bash
# Custom URL and interval
go run main.go -url "https://api.github.com/users/octocat" -interval 10s

# Custom OTLP endpoint
go run main.go -otlp-endpoint "http://jaeger:14268/api/traces"

# Custom service name
go run main.go -service-name "my-http-client"

# Debug logging with console format
go run main.go -log-level debug -log-format console

# All options combined
go run main.go \
  -url "https://httpbin.org/json" \
  -otlp-endpoint "http://localhost:4318" \
  -service-name "test-client" \
  -interval 3s \
  -log-level debug \
  -log-format json
```

## Building

To build the binary:

```bash
go build -o http-client main.go
```

Then run it:

```bash
./http-client -url "https://example.com"
```

## Logging

The program uses structured JSON logging with the following features:

### Log Levels

- `debug`: Detailed information including trace IDs and span IDs
- `info`: General information about program operation
- `warn`: Warning messages for non-critical issues
- `error`: Error messages for failures

### Log Formats

- `json`: Structured JSON format (default, production-ready)
- `console`: Human-readable console format (development-friendly)

### Log Fields

The logs include rich context information:

- `timestamp`: ISO8601 formatted timestamp
- `level`: Log level
- `message`: Log message
- `caller`: Source file and line number
- `trace_id`: OpenTelemetry trace ID (when available)
- `span_id`: OpenTelemetry span ID (when available)
- `url`: Target URL for HTTP requests
- `status_code`: HTTP response status code
- `duration`: Request duration
- `response_size`: Size of response body

### Example Log Output

**JSON Format:**
```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "info",
  "message": "HTTP request completed successfully",
  "caller": "main.go:215",
  "url": "https://httpbin.org/get",
  "status_code": 200,
  "duration": "150ms",
  "response_size": 1234,
  "trace_id": "1234567890abcdef",
  "span_id": "abcdef1234567890"
}
```

**Console Format:**
```
2024-01-15T10:30:45.123Z	INFO	HTTP request completed successfully	main.go:215	{"url": "https://httpbin.org/get", "status_code": 200, "duration": "150ms", "response_size": 1234}
```

## Tracing Details

The program creates the following hierarchical spans:

1. **request.cycle**: Root span for each request cycle
2. **http.get**: Child span for the actual HTTP request
3. **http.transport**: Child span for HTTP transport operations
4. **dns.resolve**: Child span for DNS resolution
5. **tcp.connect**: Child span for TCP connection establishment

### Span Attributes

#### Request Cycle Span (`request.cycle`)
- `service.name`: Service name for tracing
- `request.target_url`: Target URL for HTTP requests
- `request.interval`: Interval between requests
- `request.cycle.duration_ms`: Total cycle duration in milliseconds
- `request.success`: Boolean indicating if the request was successful
- `request.error`: Error message (only present if request failed)

#### HTTP Request Span (`http.get`)
- `http.method`: HTTP method (always "GET")
- `http.url`: Target URL
- `http.status_code`: HTTP response status code
- `http.response.size`: Size of response body in bytes
- `http.request.duration_ms`: Request duration in milliseconds

#### HTTP Transport Span (`http.transport`)
- `http.method`: HTTP method
- `http.url`: Full request URL
- `http.scheme`: URL scheme (http/https)
- `http.host`: Target host
- `http.path`: Request path
- `http.status_code`: HTTP response status code
- `http.status_text`: HTTP status text
- `http.transport.duration_ms`: Transport layer duration

#### DNS Resolution Span (`dns.resolve`)
- `dns.hostname`: Hostname being resolved
- `dns.port`: Target port
- `dns.duration_ms`: DNS resolution duration
- `dns.resolved_ips`: Array of resolved IP addresses

#### TCP Connection Span (`tcp.connect`)
- `tcp.host`: Target host
- `tcp.port`: Target port
- `tcp.duration_ms`: TCP connection duration
- `tcp.local_addr`: Local connection address
- `tcp.remote_addr`: Remote connection address

### Trace Context Propagation

The program automatically injects trace context into HTTP headers, allowing downstream services to continue the trace if they support OpenTelemetry.

## OTLP Backend Setup

### Using Jaeger

1. Start Jaeger with OTLP support:
   ```bash
   docker run -d --name jaeger \
     -p 16686:16686 \
     -p 4317:4317 \
     -p 4318:4318 \
     jaegertracing/all-in-one:latest \
     --collector.otlp.enabled=true
   ```

2. Run the client:
   ```bash
   go run main.go -otlp-endpoint "http://localhost:4318"
   ```

3. View traces at http://localhost:16686

### Using Grafana Tempo

1. Start Grafana Tempo:
   ```bash
   docker run -d --name tempo \
     -p 3200:3200 \
     -p 4317:4317 \
     -p 4318:4318 \
     grafana/tempo:latest
   ```

2. Run the client:
   ```bash
   go run main.go -otlp-endpoint "http://localhost:4318"
   ```

## Environment Variables

You can also configure the program using environment variables:

- `TARGET_URL`: Target URL for requests
- `OTLP_ENDPOINT`: OTLP endpoint
- `SERVICE_NAME`: Service name for tracing
- `REQUEST_INTERVAL`: Request interval (e.g., "5s", "1m")

## Error Handling

The program includes comprehensive error handling:

- Failed HTTP requests are recorded as errors in traces
- HTTP status codes >= 400 are marked as errors
- Tracer initialization failures cause the program to exit
- Individual request failures are logged but don't stop the program

## GitHub Actions

This project includes comprehensive GitHub Actions workflows for CI/CD:

### CI Workflow (`.github/workflows/ci.yml`)
- **Triggers**: Push to main/develop branches, pull requests
- **Features**:
  - Multi-version Go testing (1.21, 1.22, 1.23)
  - Test coverage reporting
  - Code linting with golangci-lint
  - Security scanning with Gosec
  - Multi-platform builds
  - Race condition detection

### Release Workflow (`.github/workflows/release.yml`)
- **Triggers**: Git tags (v*), manual workflow dispatch
- **Features**:
  - Multi-platform builds (Linux, macOS, Windows)
  - Architecture support (AMD64, ARM64)
  - Automatic GitHub releases
  - SHA256 checksums
  - Comprehensive release notes

### Creating Releases

#### Automatic Release (Recommended)
```bash
# Create and push a tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions will automatically create the release
```

#### Manual Release
```bash
# Use the release script
./scripts/release.sh v1.0.0

# Or use GitHub CLI
gh release create v1.0.0 dist/* --title "Release v1.0.0"
```

#### Manual Workflow Dispatch
1. Go to Actions tab in GitHub
2. Select "Build and Release" workflow
3. Click "Run workflow"
4. Enter version tag (e.g., v1.0.0)

## Local Development

### Makefile Commands
```bash
# Build the application
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run tests with coverage
make coverage

# Run linter
make lint

# Run security scan
make security

# Clean build artifacts
make clean

# Show help
make help
```

### Version Information
The application includes version information that can be set during build:
```bash
# Build with version info
make build VERSION=v1.0.0

# Check version
./tracer-test -version
```

## Dependencies

- `go.opentelemetry.io/otel`: OpenTelemetry core library
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`: OTLP HTTP exporter
- `go.opentelemetry.io/otel/sdk`: OpenTelemetry SDK
- `go.opentelemetry.io/otel/semconv/v1.37.0`: Semantic conventions
- `go.opentelemetry.io/otel/trace`: Trace API
- `go.uber.org/zap`: Structured logging library

## License

This project is provided as-is for educational and testing purposes.
