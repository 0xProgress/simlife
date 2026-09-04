package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/api/auth"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// ActivityHandler handles the Discord OAuth exchange for the Activity frontend.
type ActivityHandler struct {
	cfg         *config.Config
	pool        *pgxpool.Pool
	queries     *sqlc.Queries
	jwt         *auth.JWTManager
	discordAuth *auth.DiscordAuth
	redis       *redis.Client
	log         zerolog.Logger
}

// NewActivityHandler initializes the activity handler.
func NewActivityHandler(
	cfg *config.Config,
	pool *pgxpool.Pool,
	queries *sqlc.Queries,
	jwt *auth.JWTManager,
	discordAuth *auth.DiscordAuth,
	redisClient *redis.Client,
	log zerolog.Logger,
) *ActivityHandler {
	return &ActivityHandler{
		cfg:         cfg,
		pool:        pool,
		queries:     queries,
		jwt:         jwt,
		discordAuth: discordAuth,
		redis:       redisClient,
		log:         log.With().Str("handler", "activity").Logger(),
	}
}

// DiscordAuthRequest is the payload sent by the Activity to exchange an OAuth code.
type DiscordAuthRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

// DiscordAuthResponse is returned to the Activity after successful authentication.
type DiscordAuthResponse struct {
	Token  string       `json:"token"`
	Player PlayerState  `json:"player"`
}

// PlayerState represents the player's current state, returned on auth for immediate rendering.
type PlayerState struct {
	ID           string `json:"id"`
	DiscordID    string `json:"discord_id"`
	Username     string `json:"username"`
	Wallet       string `json:"wallet"`
	Bank         string `json:"bank"`
	NetWorth     string `json:"net_worth"`
	CreditScore  int    `json:"credit_score"`
	Reputation   int    `json:"reputation"`
	EconomicDay  int    `json:"economic_day"`
}

// HandleDiscordAuth exchanges a Discord OAuth code for a JWT and player state.
// This is the entry point for the Activity's authentication flow.
func (h *ActivityHandler) HandleDiscordAuth(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.activity")

	var req DiscordAuthRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" || req.RedirectURI == "" {
		writeError(w, http.StatusBadRequest, "code and redirect_uri are required")
		return
	}

	// Exchange the OAuth code with Discord
	// Note: In production, the client secret should come from a secure source (env var),
	// not from the request. For the Activity flow, Discord requires the redirect_uri
	// to match what was configured in the Discord Developer Portal.
	discordUser, err := h.discordAuth.ExchangeCode(r.Context(), req.Code, req.RedirectURI, h.cfg.DiscordClientSecret)
	if err != nil {
		log.Error().Err(fmt.Errorf("discord exchange failed: %w", err)).Msg("OAuth exchange failed")
		writeError(w, http.StatusUnauthorized, "failed to authenticate with Discord")
		return
	}

	// Find or create the player in our database
	player, err := h.queries.GetOrCreatePlayer(r.Context(), sqlc.GetOrCreatePlayerParams{
		DiscordID: discordUser.ID,
		Username:  discordUser.Username,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to get/create player: %w", err)).
			Str("discord_id", discordUser.ID).
			Msg("player creation failed")
		writeError(w, http.StatusInternalServerError, "failed to initialize player")
		return
	}

	// Ensure player has the 3 core ledger accounts
	if err := h.queries.EnsurePlayerAccounts(r.Context(), sqlc.EnsurePlayerAccountsParams{
		PlayerID: player.ID,
	}); err != nil {
		log.Warn().Err(fmt.Errorf("failed to ensure accounts: %w", err)).
			Str("player_id", player.ID).
			Msg("account provisioning failed (non-fatal)")
	}

	// Issue a JWT for the Activity session
	token, err := h.jwt.IssueToken(player.ID, discordUser.ID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to issue JWT: %w", err)).Msg("JWT issuance failed")
		writeError(w, http.StatusInternalServerError, "failed to issue session token")
		return
	}

	// Fetch current player state for immediate rendering
	playerState, err := h.buildPlayerState(r.Context(), player.ID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to build player state: %w", err)).Msg("player state fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to load player state")
		return
	}

	log.Info().
		Str("player_id", player.ID).
		Str("discord_id", discordUser.ID).
		Str("username", discordUser.Username).
		Msg("activity authentication successful")

	writeJSON(w, http.StatusOK, DiscordAuthResponse{
		Token:  token,
		Player: playerState,
	})
}

// HandleTokenRefresh issues a new JWT for an existing valid session.
func (h *ActivityHandler) HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.activity")

	playerID := getPlayerIDFromContext(r.Context())
	discordID := getDiscordIDFromContext(r.Context())

	if playerID == "" || discordID == "" {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	newToken, err := h.jwt.RefreshToken(playerID, discordID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to refresh token: %w", err)).Msg("token refresh failed")
		writeError(w, http.StatusInternalServerError, "failed to refresh session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": newToken})
}

// buildPlayerState fetches the player's current economic state for the Activity dashboard.
func (h *ActivityHandler) buildPlayerState(ctx context.Context, playerID string) (PlayerState, error) {
	player, err := h.queries.GetPlayerByID(ctx, playerID)
	if err != nil {
		return PlayerState{}, fmt.Errorf("failed to fetch player: %w", err)
	}

	accounts, err := h.queries.GetAccountsByPlayer(ctx, playerID)
	if err != nil {
		return PlayerState{}, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var walletBal, bankBal string
	for _, acc := range accounts {
		bal, err := h.queries.GetAccountBalance(ctx, acc.ID)
		if err != nil {
			continue
		}
		balStr := numericToString(bal)
		switch acc.AccountType {
		case sqlc.AccountTypeWALLET:
			walletBal = balStr
		case sqlc.AccountTypeBANK:
			bankBal = balStr
		}
	}

	// TODO: Fetch actual economic day from world state
	economicDay := 1

	return PlayerState{
		ID:          player.ID,
		DiscordID:   player.DiscordID,
		Username:    player.Username,
		Wallet:      walletBal,
		Bank:        bankBal,
		NetWorth:    walletBal, // Simplified for Layer 1; should sum all accounts
		CreditScore: 650,       // Default starting score
		Reputation:  500,       // Default neutral reputation
		EconomicDay: economicDay,
	}, nil
}

// getPlayerIDFromContext extracts the player ID from the request context.
func getPlayerIDFromContext(ctx context.Context) string {
	type contextKey string
	if v, ok := ctx.Value(contextKey("api_player_id")).(string); ok {
		return v
	}
	return ""
}

// getDiscordIDFromContext extracts the Discord user ID from the request context.
func getDiscordIDFromContext(ctx context.Context) string {
	type contextKey string
	if v, ok := ctx.Value(contextKey("api_discord_id")).(string); ok {
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

// numericToString converts pgtype.Numeric to a string representation.
func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0"
	}
	return n.String()
}