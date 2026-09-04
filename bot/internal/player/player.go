package player

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
)

// Manager handles all player record operations.
type Manager struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewManager initializes the player manager.
func NewManager(q *sqlc.Queries, log zerolog.Logger) *Manager {
	return &Manager{
		queries: q,
		log:     log.With().Str("component", "player").Logger(),
	}
}

// Player represents a player in the system.
type Player struct {
	ID           string
	DiscordID    string
	Username     string
	DisplayColor string
	CreatedAt    string
	LastActiveAt string
	IsDeleted    bool
}

// GetOrCreatePlayer fetches a player by Discord ID, creating a new record if they don't exist.
// This is called by the AuthMiddleware on every interaction.
func (m *Manager) GetOrCreatePlayer(ctx context.Context, discordID, username string) (*Player, error) {
	log := logger.FromContext(ctx, "player")

	// Try to fetch existing player
	row, err := m.queries.GetPlayerByDiscordID(ctx, discordID)
	if err == nil {
		// Player exists, update last active timestamp
		err = m.queries.UpdateLastActive(ctx, row.ID)
		if err != nil {
			log.Warn().Err(err).Str("player_id", row.ID).Msg("failed to update last active timestamp")
		}
		return playerFromRow(row), nil
	}

	// Player not found, create new record
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to fetch player: %w", err)
	}

	// Create new player with default display color
	newPlayer, err := m.queries.CreatePlayer(ctx, sqlc.CreatePlayerParams{
		DiscordID:   discordID,
		Username:    username,
		DisplayColor: "#FFFFFF", // Default white
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	log.Info().
		Str("player_id", newPlayer.ID).
		Str("discord_id", discordID).
		Str("username", username).
		Msg("new player created")

	return playerFromRow(newPlayer), nil
}

// GetPlayer fetches a player by their database ID.
func (m *Manager) GetPlayer(ctx context.Context, playerID string) (*Player, error) {
	row, err := m.queries.GetPlayerByID(ctx, playerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to fetch player: %w", err)
	}
	return playerFromRow(row), nil
}

// GetPlayerByDiscordID fetches a player by their Discord user ID.
func (m *Manager) GetPlayerByDiscordID(ctx context.Context, discordID string) (*Player, error) {
	row, err := m.queries.GetPlayerByDiscordID(ctx, discordID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to fetch player: %w", err)
	}
	return playerFromRow(row), nil
}

// UpdateLastActive updates the player's last active timestamp to the current time.
func (m *Manager) UpdateLastActive(ctx context.Context, playerID string) error {
	err := m.queries.UpdateLastActive(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}
	return nil
}

// SoftDeletePlayer marks a player as deleted without removing their record.
func (m *Manager) SoftDeletePlayer(ctx context.Context, playerID string) error {
	log := logger.FromContext(ctx, "player")

	err := m.queries.SoftDeletePlayer(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to soft delete player: %w", err)
	}

	log.Info().Str("player_id", playerID).Msg("player soft deleted")
	return nil
}

// playerFromRow converts a sqlc Player row to our domain Player struct.
func playerFromRow(row sqlc.Player) *Player {
	return &Player{
		ID:           row.ID,
		DiscordID:    row.DiscordID,
		Username:     row.Username,
		DisplayColor: row.DisplayColor,
		CreatedAt:    row.CreatedAt.Time.String(),
		LastActiveAt: row.LastActiveAt.Time.String(),
		IsDeleted:    row.IsDeleted,
	}
}