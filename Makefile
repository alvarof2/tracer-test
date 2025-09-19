# Makefile for tracer-test

# Variables
BINARY_NAME=tracer-test
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
BUILD_FLAGS=-a -installsuffix cgo

.PHONY: all build clean test coverage deps help version

# Default target
all: test build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o $(BINARY_NAME) .

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	@echo "Building Linux AMD64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	@echo "Building Linux ARM64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	@echo "Building macOS AMD64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	@echo "Building macOS ARM64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	@echo "Building Windows AMD64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Building Windows ARM64..."
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe .
	@echo "Build complete! Binaries are in the dist/ directory."

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
test-race:
	$(GOTEST) -race ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...

# Security scan
security:
	gosec ./...

# Run the application
run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) . && ./$(BINARY_NAME)

# Run with custom parameters
run-custom:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) . && ./$(BINARY_NAME) -url "https://httpbin.org/json" -interval 2s -log-level debug

# Run without OTLP
run-no-otlp:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) . && ./$(BINARY_NAME) -disable-otlp -log-format console

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

# Docker build
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker build -t $(BINARY_NAME):latest .

# Docker run
docker-run:
	docker run --rm -p 8080:8080 $(BINARY_NAME):latest

# Create checksums for dist files
checksums:
	@echo "Creating checksums for dist files..."
	@cd dist && for file in $(BINARY_NAME)-*; do \
		sha256sum "$$file" > "$$file.sha256"; \
	done
	@echo "Checksums created!"

# Install binary to GOPATH/bin
install:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage"
	@echo "  test-race    - Run tests with race detection"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  security     - Run security scan"
	@echo "  run          - Run the application"
	@echo "  run-custom   - Run with custom parameters"
	@echo "  run-no-otlp  - Run without OTLP"
	@echo "  version      - Show version information"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  checksums    - Create checksums for dist files"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  help         - Show this help"
