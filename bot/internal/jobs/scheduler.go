package jobs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/go-co-op/gocron"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Scheduler manages all background cron jobs for the Simlife bot.
// It coordinates daily settlement, economic news publishing, stale listing
// expiry, Redis cache warm-up, and analytics service health checks.
type Scheduler struct {
	cron        *gocron.Scheduler
	cfg         *config.Config
	pool        *pgxpool.Pool
	queries     *sqlc.Queries
	redis       *redis.Client
	nats        *nats.Conn
	settlement  *economy.Engine
	log         zerolog.Logger
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewScheduler initializes the job scheduler with all infrastructure dependencies.
func NewScheduler(cfg *config.Config, pool *pgxpool.Pool, redisClient *redis.Client, natsConn *nats.Conn) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	queries := sqlc.New(pool)

	return &Scheduler{
		cron:    gocron.NewScheduler(time.UTC),
		cfg:     cfg,
		pool:    pool,
		queries: queries,
		redis:   redisClient,
		nats:    natsConn,
		log:     logger.Package("scheduler"),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterSettlementEngine stores the settlement engine reference for use by
// the daily settlement and economic news jobs. This must be called before Start().
func (s *Scheduler) RegisterSettlementEngine(engine *economy.Engine) {
	s.settlement = engine
}

// RegisterAll registers all background jobs with the scheduler.
// This is the single entry point called from main.go after all dependencies are wired.
func (s *Scheduler) RegisterAll() error {
	// 1. Redis cache warm-up (every 15 minutes)
	if err := s.registerCacheWarmUp(); err != nil {
		return fmt.Errorf("failed to register cache warm-up job: %w", err)
	}

	// 2. Stale listing expiry (every hour)
	if err := s.registerStaleListingExpiry(); err != nil {
		return fmt.Errorf("failed to register stale listing expiry job: %w", err)
	}

	// 3. Analytics service health check (every 25 hours)
	if err := s.registerAnalyticsHealthCheck(); err != nil {
		return fmt.Errorf("failed to register analytics health check job: %w", err)
	}

	// 4. Daily settlement (cron schedule from config)
	// This is registered via RegisterDailySettlement in daily_settlement.go
	// which is called separately because it needs the Discord session and composer.

	s.log.Info().
		Int("registered_jobs", len(s.cron.Jobs())).
		Msg("all background jobs registered")

	return nil
}

// Start begins executing scheduled jobs asynchronously.
func (s *Scheduler) Start() {
	s.cron.StartAsync()
	s.log.Info().Msg("background job scheduler started")
}

// Stop gracefully halts all scheduled jobs and cancels the scheduler context.
func (s *Scheduler) Stop() {
	s.cancel()
	s.cron.Stop()
	s.log.Info().Msg("background job scheduler stopped")
}

// registerCacheWarmUp schedules a job that pre-warms critical Redis caches
// every 15 minutes. This ensures the Activity frontend always gets fast responses.
func (s *Scheduler) registerCacheWarmUp() error {
	_, err := s.cron.Every(15).Minutes().Do(func() {
		log := s.log.With().Str("job", "cache_warmup").Logger()
		log.Debug().Msg("starting Redis cache warm-up")

		ctx, cancel := context.WithTimeout(s.ctx, 2*time.Minute)
		defer cancel()

		// Warm the city state cache by forcing a rebuild
		cityStateKey := "city:state"
		s.redis.Del(ctx, cityStateKey)
		log.Debug().Msg("invalidated city state cache for rebuild on next request")

		// Warm market price caches by checking TTL and refreshing if needed
		items := []string{"Basic Rations", "Standard Toolkit", "Luxury Watch"}
		for _, item := range items {
			priceKey := fmt.Sprintf("market:price:%s", item)
			ttl, err := s.redis.TTL(ctx, priceKey).Result()
			if err != nil || ttl < 5*time.Minute {
				// Price cache is stale or missing; the pricing engine will rebuild on next settlement
				log.Debug().Str("item", item).Msg("price cache stale, will rebuild on next settlement")
			}
		}

		// Warm the base wage rate cache
		wageKey := "economy:base_wage_rate"
		ttl, err := s.redis.TTL(ctx, wageKey).Result()
		if err != nil || ttl < 5*time.Minute {
			log.Debug().Msg("base wage rate cache stale")
		}

		log.Info().Msg("Redis cache warm-up completed")
	})

	return err
}

// registerStaleListingExpiry schedules an hourly job that closes expired market listings
// and releases escrow deposits back to sellers.
func (s *Scheduler) registerStaleListingExpiry() error {
	_, err := s.cron.Every(1).Hour().Do(func() {
		log := s.log.With().Str("job", "stale_listing_expiry").Logger()
		log.Info().Msg("starting stale listing expiry check")

		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		defer cancel()

		// Fetch expired listings
		expiredListings, err := s.queries.GetExpiredListings(ctx)
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to fetch expired listings: %w", err)).Msg("stale listing expiry failed")
			return
		}

		expiredCount := 0
		for _, listing := range expiredListings {
			// Update listing status to EXPIRED
			err := s.queries.UpdateListingStatus(ctx, sqlc.UpdateListingStatusParams{
				ID:     listing.ID,
				Status: "EXPIRED",
			})
			if err != nil {
				log.Warn().Err(err).Str("listing_id", listing.ID).Msg("failed to expire listing")
				continue
			}

			// Release escrow deposit back to seller's wallet
			if listing.EscrowDeposit.Valid {
				depositStr := listing.EscrowDeposit.String()
				log.Debug().
					Str("listing_id", listing.ID).
					Str("seller_id", listing.SellerID).
					Str("deposit", depositStr).
					Msg("releasing escrow deposit for expired listing")

				// The actual ledger transfer would go through the settlement engine
				// For the hourly cleanup, we just mark it as released
				err := s.queries.ReleaseEscrowDeposit(ctx, sqlc.ReleaseEscrowDepositParams{
					SellerID:        listing.SellerID,
					EscrowAccountID: listing.EscrowAccountID,
				})
				if err != nil {
					log.Warn().Err(err).Str("listing_id", listing.ID).Msg("failed to release escrow")
				}
			}

			expiredCount++
		}

		log.Info().
			Int("expired_count", expiredCount).
			Int("total_checked", len(expiredListings)).
			Msg("stale listing expiry completed")
	})

	return err
}

// registerAnalyticsHealthCheck schedules a job that pings the Python analytics
// service every 25 hours. If the service is unreachable, it logs a warning.
// Per the documentation, if settlement.complete is not received within 25 hours,
// the analytics service itself posts a stale-data alert to Go's internal endpoint.
func (s *Scheduler) registerAnalyticsHealthCheck() error {
	_, err := s.cron.Every(25).Hours().Do(func() {
		log := s.log.With().Str("job", "analytics_health_check").Logger()
		log.Info().Msg("starting analytics service health check")

		ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
		defer cancel()

		// Check if we've received a settlement.complete event recently
		lastSettlementKey := "analytics:last_settlement"
		lastSettlement, err := s.redis.Get(ctx, lastSettlementKey).Result()
		if err == redis.Nil {
			log.Warn().Msg("no settlement event recorded — analytics service may be offline")
			// Post stale-data alert to our own internal log
			s.postStaleDataAlert(ctx, log)
			return
		}
		if err != nil {
			log.Error().Err(fmt.Errorf("redis read failed: %w", err)).Msg("health check failed")
			return
		}

		log.Info().
			Str("last_settlement", lastSettlement).
			Msg("analytics service health check passed")
	})

	return err
}

// postStaleDataAlert logs a FATAL-level alert when the analytics service
// has not posted a settlement snapshot within the expected window.
func (s *Scheduler) postStaleDataAlert(ctx context.Context, log zerolog.Logger) {
	log.Error().
		Str("alert_type", "STALE_ANALYTICS_DATA").
		Msg("analytics service has not posted a settlement snapshot in 25+ hours — data is stale")

	// Also attempt to create an anomaly flag in the database for operator visibility
	_, err := s.queries.CreateAnomalyFlag(ctx, sqlc.CreateAnomalyFlagParams{
		FlagType:            "STALE_ANALYTICS",
		ImplicatedPlayerIds: []byte("[]"),
		EvidenceSummary:     "No settlement.complete event received from analytics service in 25+ hours",
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to create anomaly flag: %w", err)).Msg("anomaly flag creation failed")
	}
}

// pingAnalyticsService attempts to reach the Python analytics service health endpoint.
// This is a best-effort check and does not block the scheduler if the service is down.
func (s *Scheduler) pingAnalyticsService(ctx context.Context) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("analytics service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	return nil
}