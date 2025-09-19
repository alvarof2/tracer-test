package tracer

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNew_Disabled(t *testing.T) {
	// Create a test logger with observer
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := Config{
		Endpoint:    "http://localhost:4318",
		ServiceName: "test-service",
		Disabled:    true,
	}

	tracer, err := New(config, logger)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	if tracer == nil {
		t.Error("New() returned nil tracer")
	}

	// Check that the disabled message was logged
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected at least one log entry")
	}

	// Check for the disabled message
	found := false
	for _, log := range logs {
		if log.Message == "OTLP tracing disabled - using no-op tracer" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'OTLP tracing disabled' log message")
	}
}

func TestNew_Enabled(t *testing.T) {
	// Create a test logger with observer
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := Config{
		Endpoint:    "http://localhost:4318",
		ServiceName: "test-service",
		Disabled:    false,
	}

	tracer, err := New(config, logger)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	if tracer == nil {
		t.Error("New() returned nil tracer")
	}

	// Check that initialization messages were logged
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected at least one log entry")
	}

	// Check for initialization messages
	foundInit := false
	foundSuccess := false
	for _, log := range logs {
		if log.Message == "Initializing OTLP tracer" {
			foundInit = true
		}
		if log.Message == "OTLP tracer initialized successfully" {
			foundSuccess = true
		}
	}
	if !foundInit {
		t.Error("Expected to find 'Initializing OTLP tracer' log message")
	}
	if !foundSuccess {
		t.Error("Expected to find 'OTLP tracer initialized successfully' log message")
	}
}

func TestShouldUseInsecure(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{
			name:     "https endpoint",
			endpoint: "https://example.com",
			want:     false,
		},
		{
			name:     "http endpoint",
			endpoint: "http://example.com",
			want:     true,
		},
		{
			name:     "localhost without protocol",
			endpoint: "localhost:4318",
			want:     true,
		},
		{
			name:     "127.0.0.1 without protocol",
			endpoint: "127.0.0.1:4318",
			want:     true,
		},
		{
			name:     "external domain without protocol",
			endpoint: "example.com:4318",
			want:     false,
		},
		{
			name:     "external domain with subdomain",
			endpoint: "api.example.com:4318",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUseInsecure(tt.endpoint); got != tt.want {
				t.Errorf("shouldUseInsecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanEndpointURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "http endpoint",
			endpoint: "http://localhost:4318",
			want:     "localhost:4318",
		},
		{
			name:     "https endpoint",
			endpoint: "https://example.com:4318",
			want:     "example.com:4318",
		},
		{
			name:     "endpoint without protocol",
			endpoint: "localhost:4318",
			want:     "localhost:4318",
		},
		{
			name:     "external domain",
			endpoint: "api.example.com:4318",
			want:     "api.example.com:4318",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanEndpointURL(tt.endpoint); got != tt.want {
				t.Errorf("cleanEndpointURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracer_GetTracer(t *testing.T) {
	// Create a test logger
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := Config{
		Endpoint:    "http://localhost:4318",
		ServiceName: "test-service",
		Disabled:    true, // Use no-op tracer for testing
	}

	tracer, err := New(config, logger)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	// Test GetTracer
	otelTracer := tracer.GetTracer()
	if otelTracer == nil {
		t.Error("GetTracer() returned nil tracer")
	}
}

func TestTracer_Shutdown(t *testing.T) {
	// Create a test logger
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := Config{
		Endpoint:    "http://localhost:4318",
		ServiceName: "test-service",
		Disabled:    true, // Use no-op tracer for testing
	}

	tracer, err := New(config, logger)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	// Test Shutdown
	ctx := context.Background()
	err = tracer.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}
