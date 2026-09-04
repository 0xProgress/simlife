package player

import (
	"context"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

// ReputationType represents different reputation categories.
type ReputationType string

const (
	ReputationGeneral  ReputationType = "GENERAL"
	ReputationEmployer ReputationType = "EMPLOYER"
	ReputationEmployee ReputationType = "EMPLOYEE"
	ReputationBusiness ReputationType = "BUSINESS"
	ReputationCriminal ReputationType = "CRIMINAL"
)

// Reputation represents a player's standing in a specific category.
type Reputation struct {
	PlayerID       string
	ReputationType ReputationType
	Score          int // 0-1000 scale
}

// ReputationManager handles reputation tracking and adjustments.
type ReputationManager struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewReputationManager initializes the reputation manager.
func NewReputationManager(q *sqlc.Queries, log zerolog.Logger) *ReputationManager {
	return &ReputationManager{
		queries: q,
		log:     log.With().Str("component", "reputation").Logger(),
	}
}

// AdjustReputation modifies a player's reputation score in a specific category.
// Positive delta increases reputation, negative delta decreases it.
func (m *ReputationManager) AdjustReputation(ctx context.Context, playerID string, repType ReputationType, delta int) error {
	log := logger.FromContext(ctx, "player.reputation")

	if delta == 0 {
		return nil
	}

	// Fetch current reputation
	rep, err := m.queries.GetPlayerReputation(ctx, sqlc.GetPlayerReputationParams{
		PlayerID:       playerID,
		ReputationType: string(repType),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Initialize reputation at 500 (neutral)
			newScore := 500 + delta
			if newScore < 0 {
				newScore = 0
			}
			if newScore > 1000 {
				newScore = 1000
			}

			err = m.queries.CreatePlayerReputation(ctx, sqlc.CreatePlayerReputationParams{
				PlayerID:       playerID,
				ReputationType: string(repType),
				Score:          int32(newScore),
			})
			if err != nil {
				return fmt.Errorf("failed to create reputation: %w", err)
			}

			log.Info().
				Str("player_id", playerID).
				Str("type", string(repType)).
				Int("delta", delta).
				Int("new_score", newScore).
				Msg("reputation initialized")
			return nil
		}
		return fmt.Errorf("failed to fetch reputation: %w", err)
	}

	// Adjust score with bounds checking
	newScore := int(rep.Score) + delta
	if newScore < 0 {
		newScore = 0
	}
	if newScore > 1000 {
		newScore = 1000
	}

	err = m.queries.UpdatePlayerReputation(ctx, sqlc.UpdatePlayerReputationParams{
		PlayerID:       playerID,
		ReputationType: string(repType),
		Score:          int32(newScore),
	})
	if err != nil {
		return fmt.Errorf("failed to update reputation: %w", err)
	}

	log.Info().
		Str("player_id", playerID).
		Str("type", string(repType)).
		Int("delta", delta).
		Int("new_score", newScore).
		Msg("reputation adjusted")

	return nil
}

// GetReputation fetches a player's reputation score in a specific category.
func (m *ReputationManager) GetReputation(ctx context.Context, playerID string, repType ReputationType) (*Reputation, error) {
	row, err := m.queries.GetPlayerReputation(ctx, sqlc.GetPlayerReputationParams{
		PlayerID:       playerID,
		ReputationType: string(repType),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default neutral reputation
			return &Reputation{
				PlayerID:       playerID,
				ReputationType: repType,
				Score:          500,
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch reputation: %w", err)
	}

	return &Reputation{
		PlayerID:       row.PlayerID,
		ReputationType: ReputationType(row.ReputationType),
		Score:          int(row.Score),
	}, nil
}

// GetAllReputations fetches all reputation categories for a player.
func (m *ReputationManager) GetAllReputations(ctx context.Context, playerID string) ([]*Reputation, error) {
	rows, err := m.queries.GetPlayerReputations(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reputations: %w", err)
	}

	reps := make([]*Reputation, 0, len(rows))
	for _, row := range rows {
		reps = append(reps, &Reputation{
			PlayerID:       row.PlayerID,
			ReputationType: ReputationType(row.ReputationType),
			Score:          int(row.Score),
		})
	}

	return reps, nil
}

// GetReputationTier returns a human-readable tier based on the reputation score.
func GetReputationTier(score int) string {
	switch {
	case score >= 900:
		return "Legendary"
	case score >= 750:
		return "Respected"
	case score >= 600:
		return "Trusted"
	case score >= 400:
		return "Neutral"
	case score >= 250:
		return "Suspicious"
	case score >= 100:
		return "Distrusted"
	default:
		return "Notorious"
	}
}