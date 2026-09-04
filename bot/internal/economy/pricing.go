package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
)

// PricingEngine computes market prices using WMA and supply/demand signals.
type PricingEngine struct {
	queries *db.Queries
	redis   *redis.Client
	log     zerolog.Logger
}

// NewPricingEngine initializes the pricing engine.
func NewPricingEngine(q *db.Queries, r *redis.Client, log zerolog.Logger) *PricingEngine {
	return &PricingEngine{
		queries: q,
		redis:   r,
		log:     log.With().Str("component", "pricing").Logger(),
	}
}

// ComputeAndPublishPrices calculates WMA for all traded items and publishes to Redis.
// This is called by the settlement engine after daily market trades are closed.
func (p *PricingEngine) ComputeAndPublishPrices(ctx context.Context) error {
	// TODO: Fetch distinct item types from market_trades
	itemTypes := []string{"Iron Ore", "Basic Rations", "Standard Toolkit"}

	for _, item := range itemTypes {
		price, err := p.computeWMA(ctx, item)
		if err != nil {
			p.log.Error().Err(err).Str("item", item).Msg("failed to compute WMA")
			continue
		}

		// Apply supply/demand elasticity multiplier
		supply, _ := p.redis.Get(ctx, fmt.Sprintf("market:supply:%s", item)).Int64()
		demand, _ := p.redis.Get(ctx, fmt.Sprintf("market:demand:%s", item)).Int64()
		
		if demand > supply*2 {
			price *= 1.10 // 10% increase for high demand
		} else if supply > demand*2 {
			price *= 0.90 // 10% decrease for oversupply
		}

		// Enforce floor price (cost of production inputs)
		if price < 5.0 {
			price = 5.0
		}

		// Publish to Redis with 25h TTL (refreshed daily at settlement)
		p.redis.Set(ctx, fmt.Sprintf("market:price:%s", item), price, 25*time.Hour)
	}

	return nil
}

// computeWMA calculates the 7-day Weighted Moving Average for a specific item.
// The blank identifiers (_) are used to satisfy the unusedparams linter while the DB query is stubbed.
func (p *PricingEngine) computeWMA(_ context.Context, _ string) (float64, error) {
	// TODO: Fetch last 7 days of trades for itemType from DB using p.queries
	// Example: trades, err := p.queries.GetRecentTradesByItem(ctx, db.GetRecentTradesByItemParams{ItemType: itemType, Limit: 100})
	
	// Stubbed trades for compilation
	trades := []struct {
		Price    float64
		Quantity int64
		DaysAgo  int
	}{
		{10.0, 5, 0},
		{9.5, 10, 2},
		{11.0, 2, 6},
	}

	var weightedSum float64
	var weightSum float64

	for _, t := range trades {
		// Weight decreases with age (7 days max window)
		weight := float64(7 - t.DaysAgo)
		if weight < 1 {
			weight = 1
		}
		
		totalWeight := weight * float64(t.Quantity)
		weightedSum += t.Price * totalWeight
		weightSum += totalWeight
	}

	if weightSum == 0 {
		return 10.0, nil // Default base price if no historical trades
	}

	return weightedSum / weightSum, nil
}

// GetPrice retrieves the current market price from Redis cache.
// Command handlers use this to avoid computing prices on every interaction.
func (p *PricingEngine) GetPrice(ctx context.Context, itemType string) (float64, error) {
	val, err := p.redis.Get(ctx, fmt.Sprintf("market:price:%s", itemType)).Float64()
	if err == redis.Nil {
		return 10.0, nil // Fallback base price
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read price from redis: %w", err)
	}
	return val, nil
}

// GetPriceHistory fetches the last 30 days of prices for sparkline charts.
func (p *PricingEngine) GetPriceHistory(ctx context.Context, itemType string) ([]float64, error) {
	// TODO: Query DB for daily average prices over last 30 days
	_ = ctx
	_ = itemType
	return []float64{10.0, 10.5, 9.8, 11.2, 10.9, 12.0, 11.5}, nil
}