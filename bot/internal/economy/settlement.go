package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
)

// Engine orchestrates the daily economic settlement.
type Engine struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	ledger  *Ledger
	monitor *Monitor
	tax     *TaxCalculator
	nats    *nats.Conn
	redis   *redis.Client
	log     zerolog.Logger
}

// NewEngine initializes the settlement engine.
func NewEngine(pool *pgxpool.Pool, q *db.Queries, l *Ledger, m *Monitor, t *TaxCalculator, n *nats.Conn, r *redis.Client, log zerolog.Logger) *Engine {
	return &Engine{
		pool:    pool,
		queries: q,
		ledger:  l,
		monitor: m,
		tax:     t,
		nats:    n,
		redis:   r,
		log:     log.With().Str("component", "settlement").Logger(),
	}
}

// RunDailySettlement executes all settlement phases in strict sequence.
// It enforces a hard 10-minute timeout. If exceeded, it cancels and logs FATAL.
func (e *Engine) RunDailySettlement(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	phases := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"expire_listings", e.expireListings},
		{"process_production", e.processProduction},
		{"pay_wages", e.payWages},
		{"collect_taxes", e.collectTaxes},
		{"distribute_dividends", e.distributeDividends},
		{"compute_snapshot", e.computeSnapshot},
		{"notify_analytics", e.notifyAnalytics},
	}

	for _, phase := range phases {
		select {
		case <-ctx.Done():
			e.log.Fatal().Str("phase", phase.name).Msg("settlement hard timeout exceeded, aborting")
			return ctx.Err()
		default:
		}

		start := time.Now()
		if err := phase.fn(ctx); err != nil {
			e.log.Error().Err(err).Str("phase", phase.name).Msg("settlement phase failed")
			return fmt.Errorf("phase %s failed: %w", phase.name, err)
		}
		e.log.Info().Str("phase", phase.name).Dur("duration", time.Since(start)).Msg("phase completed")
	}

	return nil
}

func (e *Engine) expireListings(ctx context.Context) error {
	// TODO: Update market_listings status to EXPIRED where updated_at < now() - interval
	// and release escrow funds back to seller via Ledger.
	return nil
}

func (e *Engine) processProduction(ctx context.Context) error {
	// TODO: Iterate active businesses, check production_config, consume inventory, produce output.
	return nil
}

func (e *Engine) payWages(ctx context.Context) error {
	// TODO: Read daily_labor, calculate wage, post Ledger transactions from business to employee.
	// Clear daily_labor for the day.
	return nil
}

func (e *Engine) collectTaxes(ctx context.Context) error {
	// Layer 5 property tax collection. Layer 1 transaction tax is handled in real-time.
	return e.tax.CollectPropertyTaxes(ctx)
}

func (e *Engine) distributeDividends(ctx context.Context) error {
	// TODO: Distribute treasury surplus to business owners based on shares.
	return nil
}

func (e *Engine) computeSnapshot(ctx context.Context) error {
	return e.monitor.ComputeAndStoreSnapshot(ctx)
}

func (e *Engine) notifyAnalytics(ctx context.Context) error {
	// Publish settlement.complete event to NATS JetStream for the Python analytics service
	return e.nats.Publish("settlement.complete", []byte(`{"status":"success"}`))
}