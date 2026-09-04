package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/shopspring/decimal"
)

// RegisterHourlyTaxAggregation schedules an hourly job that aggregates transaction
// tax data from the past hour. While Layer 1 transaction tax is collected in
// real-time during P2P transfers via the Ledger, this job aggregates the data
// into Redis for dashboard rendering and analytics caching.
func (s *Scheduler) RegisterHourlyTaxAggregation(tax *economy.TaxCalculator) error {
	if tax == nil {
		return fmt.Errorf("tax calculator is nil; cannot register tax aggregation")
	}

	_, err := s.cron.Every(1).Hour().Do(func() {
		log := logger.FromContext(s.ctx, "jobs.tax_aggregation")
		log.Debug().Msg("starting hourly tax aggregation")

		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		defer cancel()

		start := time.Now()

		// 1. Fetch tax metrics from the last hour
		metrics, err := s.queries.GetHourlyTaxMetrics(ctx)
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to fetch hourly tax metrics: %w", err)).
				Msg("tax aggregation aborted")
			return
		}

		// 2. Parse the metrics
		totalTaxCollected := numericToDecimal(metrics.TotalTaxCollected)
		transactionCount := metrics.TransactionCount

		// 3. Push aggregated metrics to Redis for the Activity dashboard
		// Key format: tax:hourly:<YYYY-MM-DD-HH>
		hourKey := time.Now().UTC().Format("2006-01-02-15")
		redisKey := fmt.Sprintf("tax:hourly:%s", hourKey)

		// Store as a simple JSON blob for the dashboard to consume
		metricsJSON := fmt.Sprintf(`{"total_collected":"%s","transaction_count":%d,"hour":"%s"}`,
			totalTaxCollected.String(), transactionCount, hourKey)

		if err := s.redis.Set(ctx, redisKey, metricsJSON, 48*time.Hour).Err(); err != nil {
			log.Warn().Err(fmt.Errorf("failed to cache hourly tax metrics: %w", err)).
				Str("redis_key", redisKey).
				Msg("tax metrics cache update failed")
		}

		// 4. Update the rolling 24-hour tax total for the economic monitor
		rollingKey := "tax:rolling_24h"
		if err := s.redis.HSet(ctx, rollingKey, hourKey, totalTaxCollected.String()).Err(); err != nil {
			log.Warn().Err(fmt.Errorf("failed to update rolling tax total: %w", err)).
				Msg("rolling tax total update failed")
		}

		// Set expiry on the rolling hash to prevent unbounded growth
		_ = s.redis.Expire(ctx, rollingKey, 48*time.Hour).Err()

		log.Info().
			Str("hour", hourKey).
			Str("total_collected", totalTaxCollected.String()).
			Int64("transaction_count", transactionCount).
			Dur("elapsed", time.Since(start)).
			Msg("hourly tax aggregation completed")
	})

	if err != nil {
		return fmt.Errorf("failed to register hourly tax aggregation job: %w", err)
	}

	s.log.Info().Msg("hourly tax aggregation job registered")
	return nil
}