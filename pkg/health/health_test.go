package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	server := New(8080)
	if server == nil {
		t.Error("New() returned nil server")
	}

	if server.server == nil {
		t.Error("New() returned server with nil http.Server")
	}

	if server.server.Addr != ":8080" {
		t.Errorf("New() server address = %s, expected :8080", server.server.Addr)
	}
}

func TestServer_SetReady(t *testing.T) {
	server := New(8080)

	// Test setting ready to true
	server.SetReady(true)
	if atomic.LoadInt32(&server.ready) != 1 {
		t.Error("SetReady(true) did not set ready to 1")
	}

	// Test setting ready to false
	server.SetReady(false)
	if atomic.LoadInt32(&server.ready) != 0 {
		t.Error("SetReady(false) did not set ready to 0")
	}
}

func TestServer_IncrementRequests(t *testing.T) {
	server := New(8080)

	// Test initial count
	initialCount := atomic.LoadInt64(&server.requests)
	if initialCount != 0 {
		t.Errorf("Initial request count = %d, expected 0", initialCount)
	}

	// Test incrementing
	server.IncrementRequests()
	count := atomic.LoadInt64(&server.requests)
	if count != 1 {
		t.Errorf("Request count after increment = %d, expected 1", count)
	}

	// Test multiple increments
	server.IncrementRequests()
	server.IncrementRequests()
	count = atomic.LoadInt64(&server.requests)
	if count != 3 {
		t.Errorf("Request count after multiple increments = %d, expected 3", count)
	}
}

func TestServer_healthHandler(t *testing.T) {
	server := New(8080)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.healthHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("healthHandler() status = %d, expected %d", w.Code, http.StatusOK)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("healthHandler() content type = %s, expected application/json", contentType)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, "healthy") {
		t.Errorf("healthHandler() body = %s, expected to contain 'healthy'", body)
	}
	if !strings.Contains(body, "timestamp") {
		t.Errorf("healthHandler() body = %s, expected to contain 'timestamp'", body)
	}
}

func TestServer_readyHandler_Ready(t *testing.T) {
	server := New(8080)
	server.SetReady(true)

	// Create test request
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.readyHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("readyHandler() status = %d, expected %d", w.Code, http.StatusOK)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("readyHandler() content type = %s, expected application/json", contentType)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, "ready") {
		t.Errorf("readyHandler() body = %s, expected to contain 'ready'", body)
	}
}

func TestServer_readyHandler_NotReady(t *testing.T) {
	server := New(8080)
	server.SetReady(false)

	// Create test request
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.readyHandler(w, req)

	// Check response
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("readyHandler() status = %d, expected %d", w.Code, http.StatusServiceUnavailable)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("readyHandler() content type = %s, expected application/json", contentType)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, "not_ready") {
		t.Errorf("readyHandler() body = %s, expected to contain 'not_ready'", body)
	}
}

func TestServer_metricsHandler(t *testing.T) {
	server := New(8080)
	
	// Set some test data
	server.SetReady(true)
	server.IncrementRequests()
	server.IncrementRequests()

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.metricsHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("metricsHandler() status = %d, expected %d", w.Code, http.StatusOK)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("metricsHandler() content type = %s, expected text/plain", contentType)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, "http_requests_total 2") {
		t.Errorf("metricsHandler() body = %s, expected to contain 'http_requests_total 2'", body)
	}
	if !strings.Contains(body, "service_ready 1") {
		t.Errorf("metricsHandler() body = %s, expected to contain 'service_ready 1'", body)
	}
}

func TestServer_Start(t *testing.T) {
	server := New(8081) // Use a specific port for testing

	// Start server in goroutine
	go func() {
		err := server.Start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Start() error = %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running by making a request
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("http://" + server.GetAddr() + "/health")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Server response status = %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestServer_Stop(t *testing.T) {
	server := New(0) // Use port 0 for testing

	// Start server in goroutine
	go func() {
		server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestServer_Integration(t *testing.T) {
	server := New(8082) // Use a specific port for testing
	server.SetReady(true) // Set ready to true

	// Start server in goroutine
	go func() {
		server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Timeout: 1 * time.Second}
	baseURL := "http://" + server.GetAddr()

	// Test health endpoint
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Errorf("Health endpoint error: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Health endpoint status = %d, expected %d", resp.StatusCode, http.StatusOK)
		}
	}

	// Test ready endpoint (should be ready by default)
	resp, err = client.Get(baseURL + "/ready")
	if err != nil {
		t.Errorf("Ready endpoint error: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Ready endpoint status = %d, expected %d", resp.StatusCode, http.StatusOK)
		}
	}

	// Test metrics endpoint
	resp, err = client.Get(baseURL + "/metrics")
	if err != nil {
		t.Errorf("Metrics endpoint error: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Metrics endpoint status = %d, expected %d", resp.StatusCode, http.StatusOK)
		}
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	server.Stop(ctx)
}
