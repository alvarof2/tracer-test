# Multi-stage build for Go application
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates (needed for go mod download and HTTPS requests)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tracer-test .

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/tracer-test .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (optional, for health checks or metrics)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set the executable as entrypoint
ENTRYPOINT ["./tracer-test"]

# Default command arguments with sensible defaults
CMD ["-url", "https://httpbin.org/json", \
     "-otlp-endpoint", "http://localhost:4318", \
     "-service-name", "http-client", \
     "-interval", "5s", \
     "-log-level", "info", \
     "-log-format", "json"]
