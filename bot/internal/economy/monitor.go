package economy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// Monitor produces the economic health snapshot after settlement.
type Monitor struct {
	queries *sqlc.Queries
	redis   *redis.Client
	log     zerolog.Logger
}

// NewMonitor initializes the economic monitor.
func NewMonitor(q *sqlc.Queries, r *redis.Client, log zerolog.Logger) *Monitor {
	return &Monitor{
		queries: q,
		redis:   r,
		log:     log.With().Str("component", "monitor").Logger(),
	}
}

// SnapshotMetrics defines the JSON structure stored in economic_snapshots.
type SnapshotMetrics struct {
	MoneySupply     string   `json:"money_supply"`
	Velocity        string   `json:"velocity"`
	TopEarners      []string `json:"top_earners"`
	InequalityRatio string   `json:"inequality_ratio"`
	BaseWageRate    string   `json:"base_wage_rate"`
}

// ComputeAndStoreSnapshot calculates daily metrics and writes them to the database.
func (m *Monitor) ComputeAndStoreSnapshot(ctx context.Context) error {
	log := logger.FromContext(ctx, "economy.monitor")
	log.Info().Msg("starting economic snapshot computation")

	// 1. Compute Money Supply
	moneySupplyNum, err := m.queries.GetTotalMoneySupply(ctx)
	if err != nil {
		return fmt.Errorf("failed to compute money supply: %w", err)
	}
	moneySupply := numericToDecimal(moneySupplyNum)

	// 2. Compute Velocity (24h volume / money supply)
	volumeNum, err := m.queries.Get24hTransactionVolume(ctx)
	if err != nil {
		return fmt.Errorf("failed to compute 24h volume: %w", err)
	}
	volume := numericToDecimal(volumeNum)

	velocity := decimal.Zero
	if !moneySupply.IsZero() {
		velocity = volume.Div(moneySupply).Truncate(4)
	}

	// 3. Top 10 Wealthiest
	topEarners, err := m.queries.GetTopEarners(ctx, 10)
	if err != nil {
		log.Warn().Err(err).Msg("failed to fetch top earners")
		topEarners = []string{}
	}

	// 4. Inequality Ratio (Top 10% avg / Bottom 50% avg)
	inequalityNum, err := m.queries.GetInequalityRatio(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("failed to compute inequality ratio")
		inequalityNum = decimalToNumeric(decimal.NewFromFloat(1.0))
	}
	inequality := numericToDecimal(inequalityNum)

	// 5. Update Base Wage Rate based on velocity
	baseWage := m.updateBaseWageRate(velocity)

	metrics := SnapshotMetrics{
		MoneySupply:     moneySupply.String(),
		Velocity:        velocity.String(),
		TopEarners:      topEarners,
		InequalityRatio: inequality.String(),
		BaseWageRate:    baseWage.String(),
	}

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Get current economic day (stubbed, should come from world state)
	currentDay := int32(1)

	_, err = m.queries.RecordEconomicSnapshot(ctx, sqlc.RecordEconomicSnapshotParams{
		EconomicDay: currentDay,
		Metrics:     metricsJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to record snapshot: %w", err)
	}

	// Cache base wage in Redis for /work command (25h TTL)
	m.redis.Set(ctx, "economy:base_wage_rate", baseWage.String(), 25*time.Hour)

	log.Info().
		Str("money_supply", moneySupply.String()).
		Str("velocity", velocity.String()).
		Str("base_wage", baseWage.String()).
		Msg("economic snapshot computed and stored")

	return nil
}

// updateBaseWageRate adjusts the base city wage based on economic velocity.
func (m *Monitor) updateBaseWageRate(velocity decimal.Decimal) decimal.Decimal {
	targetWage := decimal.NewFromInt(150) // Base 150
	if velocity.LessThan(decimal.NewFromFloat(0.30)) {
		targetWage = decimal.NewFromInt(120)
	} else if velocity.GreaterThan(decimal.NewFromFloat(0.70)) {
		targetWage = decimal.NewFromInt(180)
	}
	return targetWage
}