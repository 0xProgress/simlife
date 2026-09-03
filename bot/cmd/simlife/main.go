package main

import (
	"context"
	"net/http"
	"time"

	"github.com/0xProgress/simlife/bot/internal/bot"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
)

func main() {
	// 1. Initialize logger first to capture any early startup failures
	logger.Init("debug", "pretty", "simlife-bot")
	log := logger.Package("main")

	log.Info().Msg("starting Simlife bot service")

	// 2. Load environment configuration and fail fast if invalid
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Re-initialize logger with loaded production config
	logger.Init(cfg.LogLevel, cfg.LogFormat, "simlife-bot")
	log.Info().
		Str("log_level", cfg.LogLevel).
		Str("log_format", cfg.LogFormat).
		Msg("logger configured")

	// 3. Initialize Database (PostgreSQL via pgx)
	log.Info().Msg("initializing PostgreSQL connection pool")
	// TODO: db.Init(cfg.PostgresDSN)

	// 4. Initialize Redis
	log.Info().Msg("initializing Redis connection")
	// TODO: redis.Init(cfg.RedisAddr)

	// 5. Initialize NATS JetStream
	log.Info().Msg("initializing NATS JetStream connection")
	// TODO: nats.Init(cfg.NatsURL)

	// 6. Run all pending database migrations
	log.Info().Msg("running database migrations")
	// TODO: migrations.Run()

	// 7. Pre-load all image compositing background assets into memory
	log.Info().Msg("loading image compositing assets")
	// TODO: imaging.LoadAssets()

	// 8. Initialize Bot and Register Commands
	b, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize bot session")
	}

	if err := b.RegisterCommands(); err != nil {
		log.Fatal().Err(err).Msg("failed to register slash commands with Discord API")
	}

	// 9. Start HTTP API Server on configured port
	log.Info().Str("port", cfg.HTTPPort).Msg("starting HTTP API server")
	apiServer := startHTTPServer(cfg.HTTPPort)

	// 10. Start background job scheduler (settlements, etc.)
	log.Info().Msg("starting background job scheduler")
	// TODO: scheduler.Start(cfg.SettlementCron)

	// 11. Open Discord gateway connection (blocks until shutdown signal)
	if err := b.Start(); err != nil {
		log.Fatal().Err(err).Msg("bot exited with error")
	}

	// 12. Graceful shutdown sequence
	log.Info().Msg("shutdown signal received, initiating graceful drain")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		// Fix: Assign to variable to make it addressable for pointer method
		httpLog := logger.Package("http")
		httpLog.Error().Err(err).Msg("HTTP server forced shutdown due to timeout")
	}

	log.Info().Msg("Simlife bot service stopped cleanly")
}

func startHTTPServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap all routes in the standardized HTTP logging middleware
	handler := logger.HTTPMiddleware(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpLog := logger.Package("http")
			httpLog.Error().Err(err).Msg("HTTP server encountered unexpected error")
		}
	}()

	return srv
}
