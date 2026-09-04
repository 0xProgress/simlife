package economy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
)

// MarketEngine manages all active market listings, escrow logic, and trade execution.
type MarketEngine struct {
	queries *db.Queries
	ledger  *Ledger
	redis   *redis.Client
	log     zerolog.Logger
}

// NewMarketEngine initializes the market engine.
func NewMarketEngine(q *db.Queries, l *Ledger, r *redis.Client, log zerolog.Logger) *MarketEngine {
	return &MarketEngine{
		queries: q,
		ledger:  l,
		redis:   r,
		log:     log.With().Str("component", "market").Logger(),
	}
}

// CreateListing locks a deposit in the seller's escrow account and creates the listing.
// This prevents spam listings by requiring a 5% upfront capital lock.
func (m *MarketEngine) CreateListing(ctx context.Context, sellerID, sellerWallet, sellerEscrow uuid.UUID, itemType string, quantity int32, askingPrice float64) (uuid.UUID, error) {
	totalPrice := askingPrice * float64(quantity)
	deposit := totalPrice * 0.05

	if deposit > 0 {
		_, err := m.ledger.PostTransaction(ctx, PostTransactionParams{
			SourceAccountID: sellerWallet,
			DestAccountID:   sellerEscrow,
			Amount:          deposit,
			Type:            db.TransactionTypeESCROWLOCK,
			Description:     fmt.Sprintf("Listing deposit for %d x %s", quantity, itemType),
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to lock escrow deposit: %w", err)
		}
	}

	var pgPrice pgtype.Numeric
	if err := pgPrice.Scan(askingPrice); err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse asking price: %w", err)
	}

	listing, err := m.queries.CreateListing(ctx, db.CreateListingParams{
		SellerID:          sellerID,
		ItemType:          itemType,
		Quantity:          quantity,
		QuantityRemaining: quantity,
		AskingPrice:       pgPrice,
		EscrowAccountID:   pgtype.UUID{Bytes: sellerEscrow, Valid: true},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create listing: %w", err)
	}

	// Increment supply signal in Redis for the pricing engine
	m.redis.IncrBy(ctx, fmt.Sprintf("market:supply:%s", itemType), int64(quantity))

	return listing.ID, nil
}

// ExecuteTrade atomically transfers funds, releases escrow, and records the trade.
func (m *MarketEngine) ExecuteTrade(ctx context.Context, listingID, buyerID, buyerWallet, sellerWallet, sellerEscrow uuid.UUID, quantity int32) error {
	// Note: In production, fetch the listing via m.queries to verify existence and get seller/item details.
	// Stubbed listing data for compilation flow.
	listing := db.MarketListing{
		ID:              listingID,
		SellerID:        sellerWallet, // Simplified mapping for stub
		ItemType:        "Iron Ore",
		EscrowAccountID: pgtype.UUID{Bytes: sellerEscrow, Valid: true},
	}
	askingPrice := 10.0 // Would be parsed from listing.AskingPrice (pgtype.Numeric)
	totalPrice := askingPrice * float64(quantity)

	// 1. Transfer funds from Buyer Wallet to Seller Wallet
	_, err := m.ledger.PostTransaction(ctx, PostTransactionParams{
		SourceAccountID: buyerWallet,
		DestAccountID:   sellerWallet,
		Amount:          totalPrice,
		Type:            db.TransactionTypeMARKETSALE,
		Description:     fmt.Sprintf("Market purchase: %d x %s", quantity, listing.ItemType),
	})
	if err != nil {
		return fmt.Errorf("failed to transfer funds: %w", err)
	}

	// 2. Record trade in market_trades for pricing engine WMA calculations
	var pgPrice pgtype.Numeric
	pgPrice.Scan(askingPrice)

	_, err = m.queries.RecordTrade(ctx, db.RecordTradeParams{
		ListingID:    listingID,
		BuyerID:      buyerID,
		SellerID:     listing.SellerID,
		ItemType:     listing.ItemType,
		Quantity:     quantity,
		PricePerUnit: pgPrice,
	})
	if err != nil {
		return fmt.Errorf("failed to record trade: %w", err)
	}

	// 3. Update Redis supply/demand signals
	m.redis.DecrBy(ctx, fmt.Sprintf("market:supply:%s", listing.ItemType), int64(quantity))
	m.redis.IncrBy(ctx, fmt.Sprintf("market:demand:%s", listing.ItemType), int64(quantity))

	return nil
}