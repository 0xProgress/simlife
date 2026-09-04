package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/0xProgress/simlife/bot/internal/api/auth"
	"github.com/0xProgress/simlife/bot/internal/api/handlers"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/world"
)

// Server is the HTTP API server for the Discord Activity frontend.
// It handles all player-facing API calls, OAuth exchange, and internal
// service-to-service endpoints (e.g., Python analytics posting snapshots).
type Server struct {
	cfg          *config.Config
	pool         *pgxpool.Pool
	redis        *redis.Client
	nats         *nats.Conn
	ledger       *economy.Ledger
	market       *economy.MarketEngine
	pricing      *economy.PricingEngine
	city         *world.CityManager
	jwt          *auth.JWTManager
	discordAuth  *auth.DiscordAuth
	router       *chi.Mux
	httpServer   *http.Server
	log          zerolog.Logger
}

// NewServer initializes the API server with all dependencies and wires the routes.
func NewServer(
	cfg *config.Config,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	natsConn *nats.Conn,
	ledger *economy.Ledger,
	market *economy.MarketEngine,
	pricing *economy.PricingEngine,
	city *world.CityManager,
	log zerolog.Logger,
) (*Server, error) {
	jwtMgr, err := auth.NewJWTManager(cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWT manager: %w", err)
	}

	discordAuth := auth.NewDiscordAuth(cfg.DiscordAppID, cfg.ActivityClientID)

	s := &Server{
		cfg:         cfg,
		pool:        pool,
		redis:       redisClient,
		nats:        natsConn,
		ledger:      ledger,
		market:      market,
		pricing:     pricing,
		city:        city,
		jwt:         jwtMgr,
		discordAuth: discordAuth,
		log:         log.With().Str("component", "api").Logger(),
	}

	s.router = s.buildRouter()
	return s, nil
}

// buildRouter constructs the chi router with all middleware and routes.
func (s *Server) buildRouter() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware stack (order matters)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.HTTPMiddleware) // Structured request logging
	r.Use(middleware.Recoverer)  // Panic recovery
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(s.corsMiddleware())

	// Public routes (no auth required)
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", s.handleHealth)

		// Discord OAuth exchange (public — Activity sends OAuth code here)
		r.Post("/auth/discord", s.handleDiscordAuth)

		// Token refresh (requires valid, non-expired JWT)
		r.With(s.jwtRefreshMiddleware()).Post("/auth/refresh", s.handleTokenRefresh)

		// Protected routes (require valid player JWT)
		r.Group(func(r chi.Router) {
			r.Use(s.jwtAuthMiddleware())

			// Player state
			r.Get("/player", s.handleGetPlayer)

			// City view
			r.Get("/city", s.handleGetCity)
			r.Get("/city/district/{zone}", s.handleGetDistrict)

			// Market
			r.Get("/market/listings", s.handleGetMarketListings)
			r.Get("/market/listings/{item_type}", s.handleGetListingsByItem)
			r.Get("/market/trades/{item_type}", s.handleGetRecentTrades)
			r.Get("/market/price/{item_type}", s.handleGetPrice)
			r.Get("/market/my-listings", s.handleGetMyListings)

			// Business
			r.Get("/business", s.handleGetMyBusinesses)
			r.Get("/business/{business_id}", s.handleGetBusiness)
			r.Get("/business/{business_id}/workers", s.handleGetBusinessWorkers)
			r.Get("/business/{business_id}/financials", s.handleGetBusinessFinancials)
		})

		// Internal routes (authenticated by shared secret, not player JWT)
		// Used by the Python analytics service to post snapshots and alerts.
		r.Group(func(r chi.Router) {
			r.Use(s.analyticsSecretMiddleware())

			r.Post("/internal/analytics/snapshot", s.handleAnalyticsSnapshot)
			r.Post("/internal/analytics/alert", s.handleAnalyticsAlert)
		})
	})

	return r
}

// Start begins listening on the configured port. Non-blocking.
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         ":" + s.cfg.HTTPPort,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.log.Info().Str("port", s.cfg.HTTPPort).Msg("HTTP API server starting")

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error().Err(fmt.Errorf("http server error: %w", err)).Msg("HTTP server failed")
		}
	}()

	return nil
}

// Shutdown gracefully drains in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info().Msg("HTTP API server shutting down")
	return s.httpServer.Shutdown(ctx)
}

// handleHealth is a simple liveness probe.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// corsMiddleware configures CORS for Discord Activity iframe requests.
// Discord proxies Activity traffic through https://<app-id>.discordsays.com,
// so we must allow that origin (and localhost for development).
func (s *Server) corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Allow Discord Activity proxy origins and localhost
			allowed := false
			if origin == "http://localhost:5173" || origin == "http://localhost:3000" {
				allowed = true
			}
			// Discord Activities use *.discordsays.com as the proxy origin
			if len(origin) > 0 && (isDiscordSaysOrigin(origin) || origin == s.cfg.ActivityClientID) {
				allowed = true
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Requested-With")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isDiscordSaysOrigin checks if an origin matches Discord's Activity proxy pattern.
func isDiscordSaysOrigin(origin string) bool {
	// Match https://<app-id>.discordsays.com
	if len(origin) < 20 {
		return false
	}
	if origin[:8] != "https://" {
		return false
	}
	host := origin[8:]
	// Check for .discordsays.com suffix
	suffix := ".discordsays.com"
	if len(host) <= len(suffix) {
		return false
	}
	return host[len(host)-len(suffix):] == suffix
}