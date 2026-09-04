package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Engine orchestrates the daily economic settlement.
type Engine struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
	ledger  *Ledger
	market  *MarketEngine
	pricing *PricingEngine
	monitor *Monitor
	tax     *TaxCalculator
	nats    *nats.Conn
	redis   *redis.Client
	log     zerolog.Logger
}

// NewEngine initializes the settlement engine.
func NewEngine(pool *pgxpool.Pool, q *sqlc.Queries, l *Ledger, m *MarketEngine, p *PricingEngine, mon *Monitor, t *TaxCalculator, n *nats.Conn, r *redis.Client, log zerolog.Logger) *Engine {
	return &Engine{
		pool:    pool,
		queries: q,
		ledger:  l,
		market:  m,
		pricing: p,
		monitor: mon,
		tax:     t,
		nats:    n,
		redis:   r,
		log:     log.With().Str("component", "settlement").Logger(),
	}
}

// RunDailySettlement executes all settlement phases in strict sequence.
// It enforces a hard 10-minute timeout.
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
		{"compute_prices", e.computePrices},
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

	e.log.Info().Msg("daily settlement completed successfully")
	return nil
}

func (e *Engine) expireListings(ctx context.Context) error {
	// Fetch expired listings
	listings, err := e.queries.GetExpiredListings(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch expired listings: %w", err)
	}

	for _, listing := range listings {
		// Release escrow back to seller
		if listing.EscrowDeposit.GreaterThan(decimal.Zero) {
			_ = e.ledger.Transfer(ctx, listing.EscrowAccountID, listing.SellerWalletID, listing.EscrowDeposit, "ESCROW_RELEASE", listing.SellerID, "Listing expired, deposit returned")
		}
		_ = e.queries.UpdateListingStatus(ctx, sqlc.UpdateListingStatusParams{
			ID:     listing.ID,
			Status: "EXPIRED",
		})
	}
	return nil
}

func (e *Engine) processProduction(ctx context.Context) error {
	// Layer 4+: Iterate active businesses, check production_config, consume inventory, produce output.
	// Stubbed for Layer 1-3 compatibility.
	return nil
}

func (e *Engine) payWages(ctx context.Context) error {
	// Layer 4+: Read daily_labor, calculate wage, post Ledger transactions from business to employee.
	// Stubbed for Layer 1-3 compatibility.
	return nil
}

func (e *Engine) collectTaxes(ctx context.Context) error {
	return e.tax.CollectPropertyTaxes(ctx)
}

func (e *Engine) distributeDividends(ctx context.Context) error {
	// Layer 8+: Distribute treasury surplus to business owners based on shares.
	// Stubbed for Layer 1-3 compatibility.
	return nil
}

func (e *Engine) computePrices(ctx context.Context) error {
	return e.pricing.ComputeAndPublishPrices(ctx)
}

func (e *Engine) computeSnapshot(ctx context.Context) error {
	return e.monitor.ComputeAndStoreSnapshot(ctx)
}

func (e *Engine) notifyAnalytics(ctx context.Context) error {
	// Publish settlement.complete event to NATS JetStream
	return e.nats.Publish("settlement.complete", []byte(`{"status":"success"}`))
}