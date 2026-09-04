package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/0xProgress/simlife/bot/internal/api/auth"
	"github.com/0xProgress/simlife/bot/internal/logger"
)

// contextKey is a private type for context keys to prevent collisions.
type contextKey string

const (
	// PlayerIDKey is the context key for the authenticated player's database ID.
	PlayerIDKey contextKey = "api_player_id"
	// DiscordIDKey is the context key for the authenticated player's Discord user ID.
	DiscordIDKey contextKey = "api_discord_id"
)

// jwtAuthMiddleware extracts and validates the JWT from the Authorization header.
// On success, it injects the player ID and Discord ID into the request context.
func (s *Server) jwtAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context(), "api.middleware")

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			// Expect "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "invalid Authorization header format")
				return
			}

			tokenString := parts[1]
			claims, err := s.jwt.Validate(tokenString)
			if err != nil {
				log.Debug().Err(err).Msg("JWT validation failed")
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Inject claims into context
			ctx := context.WithValue(r.Context(), PlayerIDKey, claims.PlayerID)
			ctx = context.WithValue(ctx, DiscordIDKey, claims.DiscordID)

			// Enrich logger with player identity for downstream logging
			reqLogger := log.With().
				Str("player_id", claims.PlayerID).
				Str("discord_id", claims.DiscordID).
				Logger()
			ctx = reqLogger.WithContext(ctx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// jwtRefreshMiddleware validates that the JWT is still valid (not expired)
// but allows it to be close to expiry. Used by the /auth/refresh endpoint.
func (s *Server) jwtRefreshMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "invalid Authorization header format")
				return
			}

			claims, err := s.jwt.Validate(parts[1])
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), PlayerIDKey, claims.PlayerID)
			ctx = context.WithValue(ctx, DiscordIDKey, claims.DiscordID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// analyticsSecretMiddleware validates the shared secret used by the Python
// analytics service. This is NOT a player JWT — it's a server-to-server
// authentication mechanism using a fixed header.
func (s *Server) analyticsSecretMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context(), "api.middleware")

			secret := r.Header.Get("X-Analytics-Secret")
			if secret == "" {
				writeError(w, http.StatusUnauthorized, "missing analytics secret")
				return
			}

			// Constant-time comparison to prevent timing attacks
			if !secureCompare(secret, s.cfg.AnalyticsSecret) {
				log.Warn().
					Str("remote_addr", r.RemoteAddr).
					Msg("invalid analytics secret attempt")
				writeError(w, http.StatusUnauthorized, "invalid analytics secret")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// secureCompare performs a constant-time string comparison.
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// getPlayerIDFromContext extracts the player ID from the request context.
// Returns empty string if not present (should never happen if middleware is applied).
func getPlayerIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(PlayerIDKey).(string); ok {
		return v
	}
	return ""
}

// getDiscordIDFromContext extracts the Discord user ID from the request context.
func getDiscordIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(DiscordIDKey).(string); ok {
		return v
	}
	return ""
}

// Ensure auth package is imported (used indirectly through s.jwt and s.discordAuth).
var _ = auth.JWTManager{}