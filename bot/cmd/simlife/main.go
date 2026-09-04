package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xProgress/simlife/bot/internal/bot"
	"github.com/0xProgress/simlife/bot/internal/cache"
	"github.com/0xProgress/simlife/bot/internal/config"
	database "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/events"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/jobs"
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
		log.Fatal().Err(fmt.Errorf("config load failed: %w", err)).Msg("failed to load configuration")
	}

	// Re-initialize logger with loaded production config
	logger.Init(cfg.LogLevel, cfg.LogFormat, "simlife-bot")
	log = logger.Package("main")
	log.Info().
		Str("log_level", cfg.LogLevel).
		Str("log_format", cfg.LogFormat).
		Msg("logger configured")

	// 3. Initialize Database (PostgreSQL via pgx)
	log.Info().Msg("initializing PostgreSQL connection pool")
	dbPool, err := database.Init(cfg.PostgresDSN)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("database init failed: %w", err)).Msg("failed to initialize database connection pool")
	}
	defer dbPool.Close()

	// 4. Initialize Redis
	log.Info().Msg("initializing Redis connection")
	redisClient, err := cache.Init(cfg.RedisAddr)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("redis init failed: %w", err)).Msg("failed to initialize Redis connection")
	}
	defer redisClient.Close()

	// 5. Initialize NATS JetStream
	log.Info().Msg("initializing NATS JetStream connection")
	natsConn, err := events.Init(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("nats init failed: %w", err)).Msg("failed to initialize NATS JetStream connection")
	}
	defer natsConn.Close()

	// 6. Run all pending database migrations
	log.Info().Msg("running database migrations")
	if err := database.RunMigrations(cfg.PostgresDSN); err != nil {
		log.Fatal().Err(fmt.Errorf("migrations failed: %w", err)).Msg("failed to run database migrations")
	}

	// 7. Pre-load all image compositing background assets into memory
	log.Info().Msg("loading image compositing assets")
	if err := imaging.LoadAssets(); err != nil {
		log.Fatal().Err(fmt.Errorf("imaging assets load failed: %w", err)).Msg("failed to load image compositing assets")
	}

	// 8. Initialize Bot and Register Commands
	log.Info().Msg("initializing Discord bot session")
	b, err := bot.NewBot(cfg, dbPool, redisClient, natsConn)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("bot init failed: %w", err)).Msg("failed to initialize bot session")
	}

	if err := b.RegisterCommands(); err != nil {
		log.Fatal().Err(fmt.Errorf("command registration failed: %w", err)).Msg("failed to register slash commands with Discord API")
	}

	// 9. Start HTTP API Server on configured port
	log.Info().Str("port", cfg.HTTPPort).Msg("starting HTTP API server")
	apiServer := startHTTPServer(cfg.HTTPPort)

	// 10. Start background job scheduler (settlements, etc.)
	log.Info().Msg("starting background job scheduler")
	scheduler := jobs.NewScheduler(cfg, dbPool, redisClient, natsConn)
	scheduler.Start()

	// 11. Setup graceful shutdown signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start bot in a goroutine so we can wait for signals
	go func() {
		log.Info().Msg("opening Discord gateway connection")
		if err := b.Start(); err != nil {
			log.Error().Err(fmt.Errorf("bot start failed: %w", err)).Msg("bot exited with error")
		}
		// When b.Start() returns, trigger shutdown
		sigChan <- syscall.SIGTERM
	}()

	// Block until a shutdown signal is received
	sig := <-sigChan
	log.Info().Str("signal", sig.String()).Msg("shutdown signal received, initiating graceful drain")

	// Close bot gateway first to stop receiving new interactions
	b.Close()

	// 12. Graceful shutdown sequence
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		httpLog := logger.Package("http")
		httpLog.Error().Err(fmt.Errorf("http shutdown failed: %w", err)).Msg("HTTP server forced shutdown due to timeout")
	}

	scheduler.Stop()
	
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
			httpLog.Error().Err(fmt.Errorf("http server error: %w", err)).Msg("HTTP server encountered unexpected error")
		}
	}()

	return srv
}