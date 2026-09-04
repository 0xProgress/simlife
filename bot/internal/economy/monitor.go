package economy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
)

// Monitor produces the economic health snapshot after settlement.
type Monitor struct {
	queries *db.Queries
	redis   *redis.Client
	log     zerolog.Logger
}

// NewMonitor initializes the economic monitor.
func NewMonitor(q *db.Queries, r *redis.Client, log zerolog.Logger) *Monitor {
	return &Monitor{
		queries: q,
		redis:   r,
		log:     log.With().Str("component", "monitor").Logger(),
	}
}

// SnapshotMetrics defines the JSON structure stored in economic_snapshots.
type SnapshotMetrics struct {
	MoneySupply     float64  `json:"money_supply"`
	Velocity        float64  `json:"velocity"`
	TopEarners      []string `json:"top_earners"`
	InequalityRatio float64  `json:"inequality_ratio"`
	BaseWageRate    float64  `json:"base_wage_rate"`
}

// ComputeAndStoreSnapshot calculates daily metrics and writes them to the database.
func (m *Monitor) ComputeAndStoreSnapshot(ctx context.Context) error {
	// 1. Compute Money Supply (Sum of all WALLET and BANK balances)
	// TODO: Implement efficient SQL query for total money supply
	
	// 2. Compute Velocity (24h volume / money supply)
	// TODO: Query sum of transaction amounts in last 24h
	
	// 3. Top 10 Wealthiest & Inequality Ratio (Top 10% / Bottom 50%)
	// TODO: Query all player net worths, sort, compute ratios
	
	// Stubbed data for compilation
	metrics := SnapshotMetrics{
		MoneySupply:      1000000.0,
		Velocity:         0.45,
		TopEarners:       []string{"Player1", "Player2", "Player3"},
		InequalityRatio:  3.2,
		BaseWageRate:     15.0,
	}

	m.updateBaseWageRate(metrics.Velocity, &metrics.BaseWageRate)

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// TODO: Get current economic day from world state
	currentDay := int32(142)

	// sqlc maps jsonb to []byte by default in pgx/v5. 
	// If you used a custom override in sqlc.yaml (like null.JSON), cast accordingly.
	_, err = m.queries.RecordEconomicSnapshot(ctx, db.RecordEconomicSnapshotParams{
		EconomicDay: currentDay,
		Metrics:     metricsJSON, 
	})
	if err != nil {
		return fmt.Errorf("failed to record snapshot: %w", err)
	}

	// Cache base wage in Redis for /work command (expires in 25h to ensure daily refresh)
	m.redis.Set(ctx, "economy:base_wage_rate", metrics.BaseWageRate, 25*time.Hour)

	return nil
}

// updateBaseWageRate adjusts the base city wage based on economic velocity.
// A slow economy slightly lowers the base wage, naturally incentivizing players to seek employment.
func (m *Monitor) updateBaseWageRate(velocity float64, baseWage *float64) {
	targetWage := 15.0
	if velocity < 0.3 {
		targetWage = 12.0
	} else if velocity > 0.7 {
		targetWage = 18.0
	}
	*baseWage = targetWage
}