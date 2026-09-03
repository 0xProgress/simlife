// bot/internal/logger/middleware.go
package logger

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
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
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := generateRequestID()
		
		// Inject request ID into context for downstream correlation
		ctx := log.With().Str("request_id", requestID).Logger().WithContext(r.Context())
		r = r.WithContext(ctx)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		duration := time.Since(start)
		
		l := log.Ctx(r.Context())
		event := l.Info()
		if sw.status >= 400 {
			event = l.Error()
		}
		
		// TODO: Extract player ID from JWT if authenticated via auth middleware
		playerID := "anonymous" 

		event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", sw.status).
			Dur("duration", duration).
			Str("ip", r.RemoteAddr).
			Str("player_id", playerID).
			Str("request_id", requestID).
			Msg("http_request_processed")
	})
}