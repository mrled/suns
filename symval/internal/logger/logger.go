// Package logger provides centralized slog configuration for the application
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Config holds logger configuration
type Config struct {
	// Level sets the minimum log level (debug, info, warn, error)
	Level string
	// Format sets the output format (text or json)
	Format string
	// AddSource adds source file information to log entries
	AddSource bool
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:     getEnvOrDefault("LOG_LEVEL", "info"),
		Format:    getEnvOrDefault("LOG_FORMAT", "json"),
		AddSource: getEnvOrDefault("LOG_ADD_SOURCE", "false") == "true",
	}
}

// NewLogger creates a new slog.Logger with the given configuration
func NewLogger(cfg Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
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

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// NewDefaultLogger creates a new slog.Logger with default configuration
func NewDefaultLogger() *slog.Logger {
	return NewLogger(DefaultConfig())
}

// SetDefault sets the default slog logger
func SetDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// WithContext adds common context fields to a logger
func WithContext(logger *slog.Logger, ctx context.Context, attrs ...slog.Attr) *slog.Logger {
	return logger.With(slog.Group("context", attrsToAny(attrs)...))
}

// WithLambda adds AWS Lambda context fields to a logger
func WithLambda(logger *slog.Logger, functionName, functionVersion, requestID string) *slog.Logger {
	return logger.With(
		slog.Group("lambda",
			slog.String("function_name", functionName),
			slog.String("function_version", functionVersion),
			slog.String("request_id", requestID),
		),
	)
}

// WithService adds service context to a logger
func WithService(logger *slog.Logger, serviceName string) *slog.Logger {
	return logger.With(slog.String("service", serviceName))
}

// WithExecutable adds executable name to a logger for filtering by program
func WithExecutable(logger *slog.Logger, executableName string) *slog.Logger {
	return logger.With(slog.String("executable", executableName))
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// attrsToAny converts slog.Attr slice to []any for use with slog.Group
func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}