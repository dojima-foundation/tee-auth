package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"go.opentelemetry.io/otel/trace"
)

// Logger represents a structured logger
type Logger struct {
	*slog.Logger
}

// New creates a new logger instance based on configuration
func New(cfg *config.LoggingConfig) (*Logger, error) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var writer io.Writer
	switch cfg.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	case "file":
		if cfg.Filename == "" {
			return nil, fmt.Errorf("filename is required when output is 'file'")
		}
		file, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writer = file
	default:
		writer = os.Stdout
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add timestamp formatting
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().UTC().Format(time.RFC3339))
			}
			return a
		},
	}

	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewJSONHandler(writer, opts)
	}

	logger := slog.New(handler)

	return &Logger{Logger: logger}, nil
}

// NewDefault creates a default logger for development
func NewDefault() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	logger := slog.New(handler)
	return &Logger{Logger: logger}
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithField adds a single structured field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(key, value)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With("error", err)}
}

// WithTrace adds OpenTelemetry trace information to the logger
func (l *Logger) WithTrace(span trace.Span) *Logger {
	if span == nil {
		return l
	}

	spanCtx := span.SpanContext()
	if !spanCtx.IsValid() {
		return l
	}

	return &Logger{Logger: l.Logger.With(
		"trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String(),
	)}
}

// Convenience methods for different log levels with key-value pairs

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Debug(msg, keysAndValues...)
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Warn(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(msg, keysAndValues...)
}

// HTTP request logging helpers

func (l *Logger) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, keysAndValues ...interface{}) {
	args := []interface{}{
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ns", duration.Nanoseconds(),
	}
	args = append(args, keysAndValues...)
	l.Logger.Info("HTTP request", args...)
}

func (l *Logger) LogGRPCRequest(method string, duration time.Duration, err error, keysAndValues ...interface{}) {
	args := []interface{}{
		"method", method,
		"duration_ns", duration.Nanoseconds(),
	}
	if err != nil {
		args = append(args, "error", err.Error())
	}
	args = append(args, keysAndValues...)

	if err != nil {
		l.Logger.Error("gRPC request", args...)
	} else {
		l.Logger.Info("gRPC request", args...)
	}
}

// Database operation logging helpers

func (l *Logger) LogDatabaseQuery(query string, duration time.Duration, err error, keysAndValues ...interface{}) {
	args := []interface{}{
		"query", query,
		"duration_ns", duration.Nanoseconds(),
	}
	if err != nil {
		args = append(args, "error", err.Error())
	}
	args = append(args, keysAndValues...)

	if err != nil {
		l.Logger.Error("Database query", args...)
	} else {
		l.Logger.Debug("Database query", args...)
	}
}

// Security event logging

func (l *Logger) LogSecurityEvent(event string, userID, organizationID string, keysAndValues ...interface{}) {
	args := []interface{}{
		"event", event,
		"user_id", userID,
		"organization_id", organizationID,
	}
	args = append(args, keysAndValues...)
	l.Logger.Warn("Security event", args...)
}

func (l *Logger) LogAuthenticationAttempt(userID, organizationID string, success bool, keysAndValues ...interface{}) {
	args := []interface{}{
		"event", "authentication_attempt",
		"user_id", userID,
		"organization_id", organizationID,
		"success", success,
	}
	args = append(args, keysAndValues...)

	if success {
		l.Logger.Info("Authentication successful", args...)
	} else {
		l.Logger.Warn("Authentication failed", args...)
	}
}

func (l *Logger) LogAuthorizationAttempt(userID, organizationID, activityType string, success bool, keysAndValues ...interface{}) {
	args := []interface{}{
		"event", "authorization_attempt",
		"user_id", userID,
		"organization_id", organizationID,
		"activity_type", activityType,
		"success", success,
	}
	args = append(args, keysAndValues...)

	if success {
		l.Logger.Info("Authorization successful", args...)
	} else {
		l.Logger.Warn("Authorization failed", args...)
	}
}

// Activity logging

func (l *Logger) LogActivity(activityType, activityID, userID, organizationID string, keysAndValues ...interface{}) {
	args := []interface{}{
		"activity_type", activityType,
		"activity_id", activityID,
		"user_id", userID,
		"organization_id", organizationID,
	}
	args = append(args, keysAndValues...)
	l.Logger.Info("Activity executed", args...)
}

// Service health logging

func (l *Logger) LogServiceHealth(serviceName, status string, keysAndValues ...interface{}) {
	args := []interface{}{
		"service", serviceName,
		"status", status,
	}
	args = append(args, keysAndValues...)

	if status == "healthy" {
		l.Logger.Debug("Service health check", args...)
	} else {
		l.Logger.Error("Service unhealthy", args...)
	}
}
