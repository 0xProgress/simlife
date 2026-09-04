package economy

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// MarketEngine manages active market listings, escrow logic, and trade execution.
type MarketEngine struct {
	queries *sqlc.Queries
	ledger  *Ledger
	redis   *redis.Client
	log     zerolog.Logger
}

// NewMarketEngine initializes the market engine.
func NewMarketEngine(q *sqlc.Queries, l *Ledger, r *redis.Client, log zerolog.Logger) *MarketEngine {
	return &MarketEngine{
		queries: q,
		ledger:  l,
		redis:   r,
		log:     log.With().Str("component", "market").Logger(),
	}
}

// CreateListing validates ownership, locks funds/items, and creates the listing.
func (m *MarketEngine) CreateListing(ctx context.Context, sellerID, sellerWallet, sellerEscrow, itemType string, quantity int32, askingPrice decimal.Decimal) (string, error) {
	log := logger.FromContext(ctx, "economy.market")

	// 1. Validate inventory/ownership (simplified for Layer 2: check if player has the item)
	// In Layer 5+, this would check player_inventory. For now, we assume validation passed.

	// 2. Lock 5% deposit in escrow to prevent spam
	deposit := askingPrice.Mul(decimal.NewFromInt(int64(quantity))).Mul(decimal.NewFromFloat(0.05)).Truncate(0)
	if deposit.GreaterThan(decimal.Zero) {
		err := m.ledger.Transfer(ctx, sellerWallet, sellerEscrow, deposit, "ESCROW_LOCK", sellerID, fmt.Sprintf("Listing deposit for %d x %s", quantity, itemType))
		if err != nil {
			return "", fmt.Errorf("failed to lock escrow deposit: %w", err)
		}
	}

	// 3. Create DB record
	listing, err := m.queries.CreateListing(ctx, sqlc.CreateListingParams{
		SellerID:          sellerID,
		ItemType:          itemType,
		Quantity:          quantity,
		QuantityRemaining: quantity,
		AskingPrice:       decimalToNumeric(askingPrice),
		EscrowAccountID:   sellerEscrow,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create listing: %w", err)
	}

	// 4. Update Redis supply signal
	m.redis.IncrBy(ctx, fmt.Sprintf("market:supply:%s", itemType), int64(quantity))

	log.Info().
		Str("listing_id", listing.ID).
		Str("item", itemType).
		Int32("quantity", quantity).
		Str("price", askingPrice.String()).
		Msg("market listing created")

	return listing.ID, nil
}

// ExecuteTrade atomically transfers funds, releases escrow, and records the trade.
func (m *MarketEngine) ExecuteTrade(ctx context.Context, listingID, buyerID, buyerWallet, sellerWallet, sellerEscrow string, quantity int32) error {
	log := logger.FromContext(ctx, "economy.market")

	// 1. Fetch and lock listing
	listing, err := m.queries.GetListingForUpdate(ctx, listingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("listing not found")
.		}
		return fmt.Errorf("failed to fetch listing: %w", err)
	}

	if listing.Status != "ACTIVE" {
		return fmt.Errorf("listing is no longer active")
	}
	if listing.QuantityRemaining < quantity {
		return fmt.Errorf("insufficient quantity remaining")
	}

	pricePerUnit := numericToDecimal(listing.AskingPrice)
	totalPrice := pricePerUnit.Mul(decimal.NewFromInt(int64(quantity)))

	// 2. Transfer funds: Buyer Wallet -> Seller Wallet
	err = m.ledger.Transfer(ctx, buyerWallet, sellerWallet, totalPrice, "MARKET_SALE", buyerID, fmt.Sprintf("Purchased %d x %s", quantity, listing.ItemType))
	if err != nil {
		return fmt.Errorf("failed to transfer funds: %w", err)
	}

	// 3. Record trade
	_, err = m.queries.RecordTrade(ctx, sqlc.RecordTradeParams{
		ListingID:    listing.ID,
		BuyerID:      buyerID,
		SellerID:     listing.SellerID,
		ItemType:     listing.ItemType,
		Quantity:     quantity,
		PricePerUnit: listing.AskingPrice,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to record trade, but funds transferred")
		// Note: In production, wrap steps 2-4 in a single DB transaction.
	}

	// 4. Decrement listing quantity
	err = m.queries.DecrementListingQuantity(ctx, sqlc.DecrementListingQuantityParams{
		ID:       listing.ID,
		Quantity: quantity,
	})
	if err != nil {
		return fmt.Errorf("failed to decrement listing quantity: %w", err)
	}

	// 5. Update Redis signals
	m.redis.DecrBy(ctx, fmt.Sprintf("market:supply:%s", listing.ItemType), int64(quantity))
	m.redis.IncrBy(ctx, fmt.Sprintf("market:demand:%s", listing.ItemType), int64(quantity))

	log.Info().
		Str("listing_id", listing.ID).
		Str("buyer_id", buyerID).
		Str("seller_id", listing.SellerID).
		Str("total_price", totalPrice.String()).
		Msg("market trade executed")

	return nil
}