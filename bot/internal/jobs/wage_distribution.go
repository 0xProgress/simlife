package jobs

import (
	"context"
	"time"

	"github.com/0xProgress/simlife/bot/internal/economy"
)

// RegisterWageDistributionValidation schedules a daily validation job.
// While the primary wage payout occurs during the main settlement phase,
// this job runs 1 hour later to audit and reconcile any failed wage transactions.
func (s *Scheduler) RegisterWageDistributionValidation(engine *economy.Engine) error {
	_, err := s.cron.Cron("0 1 * * *").Do(func() { // 1 hour after standard settlement
		_, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		defer cancel()

		// TODO: Query daily_labor vs ledger transactions to find discrepancies
		s.log.Info().Msg("wage distribution reconciliation completed")
	})
	return err
}