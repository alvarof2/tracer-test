package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps the zap logger
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string
	Format string
}

// Custom log writer that converts standard log output to JSON
type jsonLogWriter struct {
	logger *zap.Logger
}

// Write implements io.Writer interface
func (w *jsonLogWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	message := string(p)
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	
	// Log the message using zap logger
	w.logger.Info("otlp_export", zap.String("message", message))
	return len(p), nil
}

// New creates a new logger instance
func New(config Config) (*Logger, error) {
	// Parse log level
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if config.Format == "console" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Create encoder
	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Note: OTLP export errors will be handled by the exporter itself
	// We can't easily redirect them to our structured logger

	return &Logger{Logger: logger}, nil
}


// WithTraceContext adds trace and span context to the logger
func (l *Logger) WithTraceContext(traceID, spanID string) *zap.Logger {
	return l.Logger.With(
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	)
}
