package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// MarketHandler handles market data and action API endpoints.
type MarketHandler struct {
	queries *sqlc.Queries
	market  *economy.MarketEngine
	pricing *economy.PricingEngine
	ledger  *economy.Ledger
	log     zerolog.Logger
}

// NewMarketHandler initializes the market handler.
func NewMarketHandler(
	queries *sqlc.Queries,
	market *economy.MarketEngine,
	pricing *economy.PricingEngine,
	ledger *economy.Ledger,
	log zerolog.Logger,
) *MarketHandler {
	return &MarketHandler{
		queries: queries,
		market:  market,
		pricing: pricing,
		ledger:  ledger,
		log:     log.With().Str("handler", "market").Logger(),
	}
}

// ListingResponse represents a market listing in API responses.
type ListingResponse struct {
	ID                string `json:"id"`
	SellerID          string `json:"seller_id"`
	SellerUsername    string `json:"seller_username"`
	ItemType          string `json:"item_type"`
	Quantity          int32  `json:"quantity"`
	QuantityRemaining int32  `json:"quantity_remaining"`
	AskingPrice       string `json:"asking_price"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
}

// TradeResponse represents a completed trade in API responses.
type TradeResponse struct {
	ID           string `json:"id"`
	BuyerID      string `json:"buyer_id"`
	SellerID     string `json:"seller_id"`
	ItemType     string `json:"item_type"`
	Quantity     int32  `json:"quantity"`
	PricePerUnit string `json:"price_per_unit"`
	TradedAt     string `json:"traded_at"`
}

// CreateListingRequest is the payload for creating a new market listing via the Activity.
type CreateListingRequest struct {
	ItemType    string `json:"item_type"`
	Quantity    int32  `json:"quantity"`
	AskingPrice string `json:"asking_price"`
}

// BuyListingRequest is the payload for purchasing from a listing via the Activity.
type BuyListingRequest struct {
	ListingID string `json:"listing_id"`
	Quantity  int32  `json:"quantity"`
}

// HandleGetMarketListings returns all active market listings, optionally filtered by item type.
func (h *MarketHandler) HandleGetMarketListings(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.market")

	itemType := r.URL.Query().Get("item_type")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(50)
	offset := int32(0)
	if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 && l <= 200 {
		limit = int32(l)
	}
	if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil && o >= 0 {
		offset = int32(o)
	}

	var listings []sqlc.MarketListing
	var err error

	if itemType != "" {
		listings, err = h.queries.GetActiveListingsByItem(r.Context(), sqlc.GetActiveListingsByItemParams{
			ItemType: itemType,
			Limit:    limit,
			Offset:   offset,
		})
	} else {
		// Fetch across all item types — would need a separate query for this
		writeError(w, http.StatusBadRequest, "item_type parameter is required")
		return
	}

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch listings: %w", err)).Msg("listings fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch listings")
		return
	}

	responses := make([]ListingResponse, 0, len(listings))
	for _, listing := range listings {
		// Fetch seller username
		seller, _ := h.queries.GetPlayerByID(r.Context(), listing.SellerID)
		sellerName := "Unknown"
		if seller != nil {
			sellerName = seller.Username
		}

		responses = append(responses, ListingResponse{
			ID:                listing.ID,
			SellerID:          listing.SellerID,
			SellerUsername:    sellerName,
			ItemType:          listing.ItemType,
			Quantity:          listing.Quantity,
			QuantityRemaining: listing.QuantityRemaining,
			AskingPrice:       numericToDecimal(listing.AskingPrice).String(),
			Status:            string(listing.Status),
			CreatedAt:         listing.CreatedAt.Time.String(),
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

// HandleGetListingsByItem returns active listings for a specific item type.
func (h *MarketHandler) HandleGetListingsByItem(w http.ResponseWriter, r *http.Request) {
	itemType := chi.URLParam(r, "item_type")
	if itemType == "" {
		writeError(w, http.StatusBadRequest, "item_type is required")
		return
	}

	limit := int32(50)
	offset := int32(0)

	listings, err := h.queries.GetActiveListingsByItem(r.Context(), sqlc.GetActiveListingsByItemParams{
		ItemType: itemType,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch listings")
		return
	}

	responses := make([]ListingResponse, 0, len(listings))
	for _, listing := range listings {
		seller, _ := h.queries.GetPlayerByID(r.Context(), listing.SellerID)
		sellerName := "Unknown"
		if seller != nil {
			sellerName = seller.Username
		}
		responses = append(responses, ListingResponse{
			ID:                listing.ID,
			SellerID:          listing.SellerID,
			SellerUsername:    sellerName,
			ItemType:          listing.ItemType,
			Quantity:          listing.Quantity,
			QuantityRemaining: listing.QuantityRemaining,
			AskingPrice:       numericToDecimal(listing.AskingPrice).String(),
			Status:            string(listing.Status),
			CreatedAt:         listing.CreatedAt.Time.String(),
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

// HandleGetRecentTrades returns the most recent trades for a specific item type.
func (h *MarketHandler) HandleGetRecentTrades(w http.ResponseWriter, r *http.Request) {
	itemType := chi.URLParam(r, "item_type")
	if itemType == "" {
		writeError(w, http.StatusBadRequest, "item_type is required")
		return
	}

	trades, err := h.queries.GetRecentTradesByItem(r.Context(), sqlc.GetRecentTradesByItemParams{
		ItemType: itemType,
		Limit:    50,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch trades")
		return
	}

	responses := make([]TradeResponse, 0, len(trades))
	for _, trade := range trades {
		responses = append(responses, TradeResponse{
			ID:           trade.ID,
			BuyerID:      trade.BuyerID,
			SellerID:     trade.SellerID,
			ItemType:     trade.ItemType,
			Quantity:     trade.Quantity,
			PricePerUnit: numericToDecimal(trade.PricePerUnit).String(),
			TradedAt:     trade.TradedAt.Time.String(),
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

// HandleGetPrice returns the current market price for a specific item type.
func (h *MarketHandler) HandleGetPrice(w http.ResponseWriter, r *http.Request) {
	itemType := chi.URLParam(r, "item_type")
	if itemType == "" {
		writeError(w, http.StatusBadRequest, "item_type is required")
		return
	}

	price, err := h.pricing.GetPrice(r.Context(), itemType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch price")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"item_type": itemType,
		"price":     price.String(),
	})
}

// HandleGetMyListings returns the authenticated player's active listings.
func (h *MarketHandler) HandleGetMyListings(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.market")

	playerID := getPlayerIDFromContext(r.Context())
	if playerID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Fetch all listings for this player (would need a new query: GetListingsBySeller)
	// For now, return empty array as placeholder
	_ = log
	_ = playerID
	writeJSON(w, http.StatusOK, []ListingResponse{})
}

// HandleCreateListing creates a new market listing via the Activity UI.
// This goes through the same ledger validation as bot commands.
func (h *MarketHandler) HandleCreateListing(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.market")

	playerID := getPlayerIDFromContext(r.Context())
	if playerID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateListingRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	price, err := decimal.NewFromString(req.AskingPrice)
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		writeError(w, http.StatusBadRequest, "invalid asking_price")
		return
	}

	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	// Fetch player's wallet and escrow accounts
	accounts, err := h.queries.GetAccountsByPlayer(r.Context(), playerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch accounts")
		return
	}

	var walletID, escrowID string
	for _, acc := range accounts {
		switch acc.AccountType {
		case sqlc.AccountTypeWALLET:
			walletID = acc.ID
		case sqlc.AccountTypeESCROW:
			escrowID = acc.ID
		}
	}

	if walletID == "" || escrowID == "" {
		writeError(w, http.StatusInternalServerError, "player accounts not found")
		return
	}

	listingID, err := h.market.CreateListing(r.Context(), playerID, walletID, escrowID, req.ItemType, req.Quantity, price)
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			writeError(w, http.StatusPaymentRequired, "insufficient funds for listing deposit")
			return
		}
		log.Error().Err(fmt.Errorf("failed to create listing: %w", err)).Msg("listing creation failed")
		writeError(w, http.StatusInternalServerError, "failed to create listing")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"listing_id": listingID,
		"status":     "created",
	})
}

// HandleBuyListing executes a market purchase via the Activity UI.
func (h *MarketHandler) HandleBuyListing(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.market")

	buyerID := getPlayerIDFromContext(r.Context())
	if buyerID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req BuyListingRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	// Fetch the listing
	listing, err := h.queries.GetListingByID(r.Context(), req.ListingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "listing not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch listing")
		return
	}

	if listing.Status != "ACTIVE" {
		writeError(w, http.StatusGone, "listing is no longer active")
		return
	}

	if int32(req.Quantity) > listing.QuantityRemaining {
		writeError(w, http.StatusBadRequest, "insufficient quantity available")
		return
	}

	// Fetch buyer and seller wallets
	buyerAccounts, _ := h.queries.GetAccountsByPlayer(r.Context(), buyerID)
	sellerAccounts, _ := h.queries.GetAccountsByPlayer(r.Context(), listing.SellerID)

	var buyerWalletID, sellerWalletID string
	for _, acc := range buyerAccounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			buyerWalletID = acc.ID
			break
		}
	}
	for _, acc := range sellerAccounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			sellerWalletID = acc.ID
			break
		}
	}

	if buyerWalletID == "" || sellerWalletID == "" {
		writeError(w, http.StatusInternalServerError, "account not found")
		return
	}

	err = h.market.ExecuteTrade(r.Context(), req.ListingID, buyerID, buyerWalletID, sellerWalletID, listing.EscrowAccountID, int32(req.Quantity))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			writeError(w, http.StatusPaymentRequired, "insufficient funds")
			return
		}
		log.Error().Err(fmt.Errorf("trade execution failed: %w", err)).Msg("trade failed")
		writeError(w, http.StatusInternalServerError, "trade execution failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "completed",
	})
}

// numericToDecimal converts pgtype.Numeric to shopspring/decimal.
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

// getPlayerIDFromContext extracts the player ID from the request context.
func getPlayerIDFromContext(ctx context.Context) string {
	type contextKey string
	if v, ok := ctx.Value(contextKey("api_player_id")).(string); ok {
		return v
	}
	return ""
}

// readJSON decodes the request body into the provided target.
func readJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is nil")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// writeError writes a standardized error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}