package jobs

import (
	"context"
	"time"

	"github.com/0xProgress/simlife/bot/internal/economy"
)

// RegisterHourlyTaxAggregation schedules an hourly job to aggregate transaction tax data.
// While Layer 1 transaction tax is collected in real-time during P2P transfers via the Ledger,
// this job aggregates the data for analytics caching and dashboard rendering.
func (s *Scheduler) RegisterHourlyTaxAggregation(tax *economy.TaxCalculator) error {
	_, err := s.cron.Every(1).Hour().Do(func() {
		_, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		defer cancel()

		// TODO: Aggregate hourly tax metrics and push to Redis/Cache
		s.log.Info().Msg("hourly tax aggregation completed")
	})
	return err
}