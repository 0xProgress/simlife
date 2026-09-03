// bot/internal/logger/logger.go
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global structured logger.
func Init(levelStr, format, serviceName string) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	var writer io.Writer = os.Stdout
	if format == "pretty" {
		writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	// Set global logger with standard fields
	log.Logger = zerolog.New(writer).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

// Package returns a logger instance with the package name attached for tracing.
func Package(name string) zerolog.Logger {
	return log.Logger.With().Str("package", name).Logger()
}