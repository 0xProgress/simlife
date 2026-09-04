package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// PlayerHandler handles player state API endpoints.
type PlayerHandler struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewPlayerHandler initializes the player handler.
func NewPlayerHandler(queries *sqlc.Queries, log zerolog.Logger) *PlayerHandler {
	return &PlayerHandler{
		queries: queries,
		log:     log.With().Str("handler", "player").Logger(),
	}
}

// PlayerResponse is the full player state returned by the API.
type PlayerResponse struct {
	ID            string            `json:"id"`
	DiscordID     string            `json:"discord_id"`
	Username      string            `json:"username"`
	Accounts      []AccountResponse `json:"accounts"`
	NetWorth      string            `json:"net_worth"`
	CreditScore   int               `json:"credit_score"`
	Reputation    int               `json:"reputation"`
	LastActiveAt  string            `json:"last_active_at"`
	CreatedAt     string            `json:"created_at"`
}

// AccountResponse represents a single ledger account.
type AccountResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Balance   string `json:"balance"`
}

// HandleGetPlayer returns the authenticated player's full state.
func (h *PlayerHandler) HandleGetPlayer(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.player")

	playerID := getPlayerIDFromContext(r.Context())
	if playerID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	player, err := h.queries.GetPlayerByID(r.Context(), playerID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).
			Str("player_id", playerID).
			Msg("player fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch player")
		return
	}

	accounts, err := h.queries.GetAccountsByPlayer(r.Context(), playerID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch accounts: %w", err)).
			Str("player_id", playerID).
			Msg("accounts fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch accounts")
		return
	}

	accountResponses := make([]AccountResponse, 0, len(accounts))
	var netWorth decimal.Decimal
	for _, acc := range accounts {
		bal, err := h.queries.GetAccountBalance(r.Context(), acc.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", acc.ID).Msg("balance fetch failed")
			continue
		}
		balDecimal := numericToDecimal(bal)
		netWorth = netWorth.Add(balDecimal)
		accountResponses = append(accountResponses, AccountResponse{
			ID:      acc.ID,
			Type:    string(acc.AccountType),
			Balance: balDecimal.String(),
		})
	}

	response := PlayerResponse{
		ID:           player.ID,
		DiscordID:    player.DiscordID,
		Username:     player.Username,
		Accounts:     accountResponses,
		NetWorth:     netWorth.String(),
		CreditScore:  650, // TODO: Fetch from credit_scores table when implemented
		Reputation:   500, // TODO: Fetch from reputations table
		LastActiveAt: player.LastActiveAt.Time.String(),
		CreatedAt:    player.CreatedAt.Time.String(),
	}

	writeJSON(w, http.StatusOK, response)
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