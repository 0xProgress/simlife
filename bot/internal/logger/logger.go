package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global structured logger.
// It configures the output format (JSON for production, pretty-print for development),
// sets the global log level, and attaches the service name to all log entries.
func Init(levelStr, format, serviceName string) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		// Fallback to Info if the provided level string is invalid to prevent silent failures
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	
	// Enforce nanosecond precision for timestamps to accurately order high-throughput events
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"

	var writer io.Writer = os.Stdout
	if format == "pretty" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			// Exclude the 'package' field from the pretty console output to reduce noise,
			// but it will still be present in the underlying JSON structure.
			FieldsExclude: []string{"package"}, 
		}
	}

	// Initialize the global logger with base fields.
	// Every log entry will automatically include the timestamp and service name.
	log.Logger = zerolog.New(writer).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

// Package returns a logger instance with the package name attached for tracing.
// This should be called at the package level or initialization phase to create
// a scoped logger that automatically includes the package name in every log entry.
func Package(name string) zerolog.Logger {
	return log.Logger.With().Str("package", name).Logger()
}

// FromContext extracts a logger from the context if one exists (e.g., injected by HTTP middleware
// with a request ID or player ID), otherwise falls back to the global logger. It then attaches 
// the package name. This is crucial for correlating logs across a single request lifecycle.
func FromContext(ctx context.Context, pkgName string) zerolog.Logger {
	l := log.Ctx(ctx)
	
	// zerolog's log.Ctx returns a pointer to a disabled logger if no logger is found in the context.
	// We check the level to determine if a valid logger was injected by the middleware.
	if l != nil && l.GetLevel() != zerolog.Disabled {
		return l.With().Str("package", pkgName).Logger()
	}
	
	// Fallback to the global logger with the package name attached
	return Package(pkgName)
}