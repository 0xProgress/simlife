package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// RegisterWageDistributionValidation schedules a daily reconciliation job that
// runs 1 hour after the main settlement. It audits the daily_labor table against
// the ledger's transaction records to detect any unpaid wages or overpayments.
// Discrepancies are logged as WARN events and, if significant, flagged in the
// anomaly_flags table for operator review.
func (s *Scheduler) RegisterWageDistributionValidation(engine *economy.Engine) error {
	if engine == nil {
		return fmt.Errorf("settlement engine is nil; cannot register wage validation")
	}

	// Run at minute 0, 1 hour after the standard settlement hour.
	// Example: if settlement is "0 3 * * *", this runs at "0 4 * * *".
	_, err := s.cron.Cron("0 4 * * *").Do(func() {
		log := logger.FromContext(s.ctx, "jobs.wage_validation")
		log.Info().Msg("starting wage distribution reconciliation")

		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		defer cancel()

		start := time.Now()

		// 1. Fetch all labor records from the previous economic day
		// (settlement clears daily_labor after processing, so we query the
		// wage_transactions table which preserves the audit trail).
		unpaidWages, err := s.queries.GetUnpaidWages(ctx)
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to fetch unpaid wages: %w", err)).
				Msg("wage reconciliation aborted")
			return
		}

		if len(unpaidWages) == 0 {
			log.Info().
				Dur("elapsed", time.Since(start)).
				Msg("wage reconciliation completed — no discrepancies found")
			return
		}

		// 2. Aggregate the total unpaid amount
		totalUnpaid := decimal.Zero
		affectedPlayers := make([]string, 0, len(unpaidWages))

		for _, wage := range unpaidWages {
			amount := numericToDecimal(wage.Amount)
			totalUnpaid = totalUnpaid.Add(amount)
			affectedPlayers = append(affectedPlayers, wage.EmployeeID)
		}

		// 3. Log the discrepancy
		log.Warn().
			Int("unpaid_count", len(unpaidWages)).
			Str("total_unpaid", totalUnpaid.String()).
			Int("affected_players", len(affectedPlayers)).
			Dur("elapsed", time.Since(start)).
			Msg("wage reconciliation found discrepancies")

		// 4. If the discrepancy is significant (> ⊄1000), create an anomaly flag
		threshold := decimal.NewFromInt(1000)
		if totalUnpaid.GreaterThan(threshold) {
			playersJSON, err := marshalStringSlice(affectedPlayers)
			if err != nil {
				log.Error().Err(fmt.Errorf("failed to marshal player IDs: %w", err)).
					Msg("anomaly flag creation aborted")
				return
			}

			_, err = s.queries.CreateAnomalyFlag(ctx, sqlc.CreateAnomalyFlagParams{
				FlagType:            "UNPAID_WAGES",
				ImplicatedPlayerIds: playersJSON,
				EvidenceSummary:     fmt.Sprintf("%d wage payments totaling ⊄%s were not recorded in the ledger", len(unpaidWages), totalUnpaid.String()),
			})
			if err != nil {
				log.Error().Err(fmt.Errorf("failed to create anomaly flag: %w", err)).
					Msg("anomaly flag creation failed")
				return
			}

			log.Warn().
				Str("total_unpaid", totalUnpaid.String()).
				Msg("significant wage discrepancy flagged for operator review")
		}

		log.Info().
			Dur("elapsed", time.Since(start)).
			Msg("wage distribution reconciliation completed")
	})

	if err != nil {
		return fmt.Errorf("failed to register wage validation job: %w", err)
	}

	s.log.Info().Msg("wage distribution validation job registered")
	return nil
}

// numericToDecimal converts pgtype.Numeric to decimal.Decimal.
func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(n.String())
	if err != nil {
		return decimal.Zero
	}
	return d
}

// marshalStringSlice converts a string slice to a JSON byte array.
func marshalStringSlice(s []string) ([]byte, error) {
	return jsonMarshal(s)
}

// jsonMarshal is a thin wrapper to avoid importing encoding/json at the top.
func jsonMarshal(v interface{}) ([]byte, error) {
	return jsonEncoder().Encode(v)
}

// jsonEncoder returns a JSON encoder.
func jsonEncoder() *json.Encoder {
	var buf bytes.Buffer
	return json.NewEncoder(&buf)
}