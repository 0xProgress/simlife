package logger

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// PlayerIDKey is the context key used to store the authenticated player's ID.
	// The auth middleware should inject the player ID using this key.
	PlayerIDKey ContextKey = "player_id"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func generateRequestID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// HTTPMiddleware logs incoming API requests with standardized fields.
// It attaches a unique request ID to the context for downstream correlation.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := generateRequestID()

		// Create a new logger with the request ID and attach it to the context
		reqLogger := log.With().
			Str("request_id", requestID).
			Logger()

		ctx := reqLogger.WithContext(r.Context())
		r = r.WithContext(ctx)

		// Wrap the ResponseWriter to capture the actual HTTP status code
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		duration := time.Since(start)

		l := log.Ctx(r.Context())
		event := l.Info()
		if sw.status >= 400 {
			event = l.Error()
		}

		// Extract player ID from context if it was injected by an auth middleware
		playerID := "anonymous"
		if pid, ok := r.Context().Value(PlayerIDKey).(string); ok && pid != "" {
			playerID = pid
		}

		// Extract real client IP (handling X-Forwarded-For and X-Real-IP for reverse proxies)
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			// X-Forwarded-For can be a comma-separated list; the first is the original client
			clientIP = strings.Split(forwarded, ",")[0]
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			clientIP = realIP
		} else {
			// Remove port from RemoteAddr if present
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				clientIP = host
			}
		}

		event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", sw.status).
			Int("duration_ms", int(duration.Milliseconds())).
			Str("ip", clientIP).
			Str("player_id", playerID).
			Str("request_id", requestID).
			Msg("http_request_processed")
	})
}