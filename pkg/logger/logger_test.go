package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name: "valid json logger",
			config: Config{
				Level:  "info",
				Format: "json",
			},
			want: true,
		},
		{
			name: "valid console logger",
			config: Config{
				Level:  "debug",
				Format: "console",
			},
			want: true,
		},
		{
			name: "invalid log level defaults to info",
			config: Config{
				Level:  "invalid",
				Format: "json",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if (err == nil) != tt.want {
				t.Errorf("New() error = %v, wantErr %v", err, !tt.want)
				return
			}
			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestLogger_WithTraceContext(t *testing.T) {
	// Create a test logger with observer
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	// Test WithTraceContext
	traceID := "1234567890abcdef"
	spanID := "abcdef1234567890"
	
	contextLogger := logger.WithTraceContext(traceID, spanID)
	contextLogger.Info("test message")

	// Check that the log entry contains trace context
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.ContextMap()["trace_id"] != traceID {
		t.Errorf("Expected trace_id %s, got %v", traceID, log.ContextMap()["trace_id"])
	}
	if log.ContextMap()["span_id"] != spanID {
		t.Errorf("Expected span_id %s, got %v", spanID, log.ContextMap()["span_id"])
	}
}

func TestJsonLogWriter(t *testing.T) {
	// Create a test logger with observer
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Create jsonLogWriter
	writer := &jsonLogWriter{logger: logger}

	// Test Write method
	testMessage := "test log message\n"
	n, err := writer.Write([]byte(testMessage))
	
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(testMessage) {
		t.Errorf("Write() returned %d, expected %d", n, len(testMessage))
	}

	// Check that the message was logged
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	expectedMessage := "test log message" // without newline
	if log.ContextMap()["message"] != expectedMessage {
		t.Errorf("Expected message %s, got %v", expectedMessage, log.ContextMap()["message"])
	}
}

func TestJsonLogWriter_WithoutNewline(t *testing.T) {
	// Create a test logger with observer
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Create jsonLogWriter
	writer := &jsonLogWriter{logger: logger}

	// Test Write method without newline
	testMessage := "test log message without newline"
	n, err := writer.Write([]byte(testMessage))
	
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(testMessage) {
		t.Errorf("Write() returned %d, expected %d", n, len(testMessage))
	}

	// Check that the message was logged
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.ContextMap()["message"] != testMessage {
		t.Errorf("Expected message %s, got %v", testMessage, log.ContextMap()["message"])
	}
}

func TestLogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			config := Config{
				Level:  level,
				Format: "json",
			}
			
			logger, err := New(config)
			if err != nil {
				t.Errorf("New() error = %v", err)
				return
			}
			
			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestLogFormats(t *testing.T) {
	formats := []string{"json", "console"}
	
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			config := Config{
				Level:  "info",
				Format: format,
			}
			
			logger, err := New(config)
			if err != nil {
				t.Errorf("New() error = %v", err)
				return
			}
			
			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}
