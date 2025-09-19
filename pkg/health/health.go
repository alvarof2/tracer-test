package health

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Server provides health check endpoints
type Server struct {
	server   *http.Server
	ready    int32
	requests int64
}

// New creates a new health server
func New(port int) *Server {
	mux := http.NewServeMux()
	
	server := &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	// Health check endpoint
	mux.HandleFunc("/health", server.healthHandler)
	
	// Readiness check endpoint
	mux.HandleFunc("/ready", server.readyHandler)
	
	// Simple metrics endpoint
	mux.HandleFunc("/metrics", server.metricsHandler)

	return server
}

// Start starts the health server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop stops the health server
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// SetReady sets the readiness status
func (s *Server) SetReady(ready bool) {
	if ready {
		atomic.StoreInt32(&s.ready, 1)
	} else {
		atomic.StoreInt32(&s.ready, 0)
	}
}

// IncrementRequests increments the request counter
func (s *Server) IncrementRequests() {
	atomic.AddInt64(&s.requests, 1)
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.server.Addr
}

// healthHandler handles /health endpoint
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339)); err != nil {
		// Log error if we can't write response
		// Note: We can't use a logger here as it's not available in this context
		// The error will be handled by the HTTP server
	}
}

// readyHandler handles /ready endpoint
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	ready := atomic.LoadInt32(&s.ready)
	
	w.Header().Set("Content-Type", "application/json")
	
	if ready == 1 {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, `{"status":"ready","timestamp":"%s"}`, time.Now().Format(time.RFC3339)); err != nil {
			// Log error if we can't write response
			// Note: We can't use a logger here as it's not available in this context
			// The error will be handled by the HTTP server
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := fmt.Fprintf(w, `{"status":"not_ready","timestamp":"%s"}`, time.Now().Format(time.RFC3339)); err != nil {
			// Log error if we can't write response
			// Note: We can't use a logger here as it's not available in this context
			// The error will be handled by the HTTP server
		}
	}
}

// metricsHandler handles /metrics endpoint
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	requests := atomic.LoadInt64(&s.requests)
	ready := atomic.LoadInt32(&s.ready)
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	
	if _, err := fmt.Fprintf(w, `# HTTP Client Metrics
http_requests_total %d
service_ready %d
`, requests, ready); err != nil {
		// Log error if we can't write response
		// Note: We can't use a logger here as it's not available in this context
		// The error will be handled by the HTTP server
	}
}
