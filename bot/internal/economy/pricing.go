package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// PricingEngine computes market prices using WMA and supply/demand signals.
type PricingEngine struct {
	queries *sqlc.Queries
	redis   *redis.Client
	log     zerolog.Logger
}

// NewPricingEngine initializes the pricing engine.
func NewPricingEngine(q *sqlc.Queries, r *redis.Client, log zerolog.Logger) *PricingEngine {
	return &PricingEngine{
		queries: q,
		redis:   r,
		log:     log.With().Str("component", "pricing").Logger(),
	}
}

// ComputeAndPublishPrices calculates WMA for all traded items and publishes to Redis.
func (p *PricingEngine) ComputeAndPublishPrices(ctx context.Context) error {
	log := logger.FromContext(ctx, "economy.pricing")
	log.Info().Msg("starting price computation phase")

	// Fetch distinct item types from recent trades
	items, err := p.queries.GetDistinctTradedItems(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch traded items: %w", err)
	}

	for _, itemType := range items {
		price, err := p.computeWMA(ctx, itemType)
		if err != nil {
			p.log.Error().Err(err).Str("item", itemType).Msg("failed to compute WMA")
			continue
		}

		// Apply supply/demand elasticity
		supplyStr, _ := p.redis.Get(ctx, fmt.Sprintf("market:supply:%s", itemType)).Result()
		demandStr, _ := p.redis.Get(ctx, fmt.Sprintf("market:demand:%s", itemType)).Result()

		supply := decimal.RequireFromString(supplyStr).IntPart()
		demand := decimal.RequireFromString(demandStr).IntPart()

		if demand > supply*2 {
			price = price.Mul(decimal.NewFromFloat(1.10)) // 10% increase
		} else if supply > demand*2 {
			price = price.Mul(decimal.NewFromFloat(0.90)) // 10% decrease
		}

		// Enforce floor price (e.g., 5.00)
		floorPrice := decimal.NewFromFloat(5.00)
		if price.LessThan(floorPrice) {
			price = floorPrice
		}

		// Publish to Redis with 25h TTL
		p.redis.Set(ctx, fmt.Sprintf("market:price:%s", itemType), price.String(), 25*time.Hour)
	}

	log.Info().Msg("price computation phase completed")
	return nil
}

// computeWMA calculates the 7-day Weighted Moving Average for a specific item.
func (p *PricingEngine) computeWMA(ctx context.Context, itemType string) (decimal.Decimal, error) {
	trades, err := p.queries.GetRecentTradesForWMA(ctx, sqlc.GetRecentTradesForWMAParams{
		ItemType: itemType,
		Limit:    100,
	})
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to fetch trades for WMA: %w", err)
	}

	if len(trades) == 0 {
		return decimal.NewFromFloat(10.00), nil // Default base price
	}

	var weightedSum decimal.Decimal
	var weightSum decimal.Decimal

	for _, t := range trades {
		// Weight decreases with age (7 days max window). 
		// daysAgo is calculated in SQL or assumed here based on traded_at.
		// For simplicity, we use a uniform weight of 1.0 per trade in this stub, 
		// but in production, calculate: weight = max(1, 7 - daysAgo)
		weight := decimal.NewFromInt(1)
		qty := decimal.NewFromInt(int64(t.Quantity))
		
		totalWeight := weight.Mul(qty are decimal) // Pseudo-code for decimal math
		// Correct decimal math:
		totalWeightDec := weight.Mul(qty)
		priceDec := numericToDecimal(t.PricePerUnit)
		
		weightedSum = weightedSum.Add(priceDec.Mul(totalWeightDec))
		weightSum = weightSum.Add(totalWeightDec)
	}

	if weightSum.IsZero() {
		return decimal.NewFromFloat(10.00), nil
	}

	return weightedSum.Div(weightSum).Truncate(2), nil
}

// GetPrice retrieves the current market price from Redis cache.
func (p *PricingEngine) GetPrice(ctx context.Context, itemType string) (decimal.Decimal, error) {
	val, err := p.redis.Get(ctx, fmt.Sprintf("market:price:%s", itemType)).Result()
	if err == redis.Nil {
		// Fallback to DB or default
		return decimal.NewFromFloat(10.00), nil
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to read price from redis: %w", err)
	}
	return decimal.RequireFromString(val), nil
}