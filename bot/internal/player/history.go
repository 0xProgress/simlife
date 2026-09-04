package player

import (
	"context"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/rs/zerolog"
)

// RelationshipType represents the type of relationship between players.
type RelationshipType string

const (
	RelationshipEmployment    RelationshipType = "EMPLOYMENT"
	RelationshipPartnership   RelationshipType = "PARTNERSHIP"
	RelationshipLending       RelationshipType = "LENDING"
	RelationshipOrganization  RelationshipType = "ORGANIZATION"
	RelationshipMentorship    RelationshipType = "MENTORSHIP"
)

// Relationship represents a connection between two players.
type Relationship struct {
	ID               string
	PlayerID         string
	RelatedPlayerID  string
	RelationshipType RelationshipType
	Metadata         string // JSON metadata (e.g., business ID, loan amount, organization ID)
	CreatedAt        string
}

// HistoryManager handles player relationship and history tracking.
type HistoryManager struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewHistoryManager initializes the history manager.
func NewHistoryManager(q *sqlc.Queries, log zerolog.Logger) *HistoryManager {
	return &HistoryManager{
		queries: q,
		log:     log.With().Str("component", "history").Logger(),
	}
}

// RecordRelationship creates a new relationship record between two players.
func (m *HistoryManager) RecordRelationship(ctx context.Context, playerID, relatedPlayerID string, relType RelationshipType, metadata string) (string, error) {
	log := logger.FromContext(ctx, "player.history")

	rel, err := m.queries.CreateRelationship(ctx, sqlc.CreateRelationshipParams{
		PlayerID:         playerID,
		RelatedPlayerID:  relatedPlayerID,
		RelationshipType: string(relType),
		Metadata:         metadata,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create relationship: %w", err)
	}

	log.Info().
		Str("player_id", playerID).
		Str("related_player_id", relatedPlayerID).
		Str("type", string(relType)).
		Msg("relationship recorded")

	return rel.ID, nil
}

// GetRelationships fetches all relationships for a player.
func (m *HistoryManager) GetRelationships(ctx context.Context, playerID string) ([]*Relationship, error) {
	rows, err := m.queries.GetPlayerRelationships(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationships: %w", err)
	}

	rels := make([]*Relationship, 0, len(rows))
	for _, row := range rows {
		rels = append(rels, &Relationship{
			ID:               row.ID,
			PlayerID:         row.PlayerID,
			RelatedPlayerID:  row.RelatedPlayerID,
			RelationshipType: RelationshipType(row.RelationshipType),
			Metadata:         row.Metadata,
			CreatedAt:        row.CreatedAt.Time.String(),
		})
	}

	return rels, nil
}

// GetRelationshipsByType fetches relationships of a specific type for a player.
func (m *HistoryManager) GetRelationshipsByType(ctx context.Context, playerID string, relType RelationshipType) ([]*Relationship, error) {
	rows, err := m.queries.GetPlayerRelationshipsByType(ctx, sqlc.GetPlayerRelationshipsByTypeParams{
		PlayerID:         playerID,
		RelationshipType: string(relType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationships: %w", err)
	}

	rels := make([]*Relationship, 0, len(rows))
	for _, row := range rows {
		rels = append(rels, &Relationship{
			ID:               row.ID,
			PlayerID:         row.PlayerID,
			RelatedPlayerID:  row.RelatedPlayerID,
			RelationshipType: RelationshipType(row.RelationshipType),
			Metadata:         row.Metadata,
			CreatedAt:        row.CreatedAt.Time.String(),
		})
	}

	return rels, nil
}

// TerminateRelationship marks a relationship as ended.
func (m *HistoryManager) TerminateRelationship(ctx context.Context, relationshipID string) error {
	log := logger.FromContext(ctx, "player.history")

	err := m.queries.TerminateRelationship(ctx, relationshipID)
	if err != nil {
		return fmt.Errorf("failed to terminate relationship: %w", err)
	}

	log.Info().Str("relationship_id", relationshipID).Msg("relationship terminated")
	return nil
}

// GetRelationshipCount returns the number of active relationships of a specific type.
func (m *HistoryManager) GetRelationshipCount(ctx context.Context, playerID string, relType RelationshipType) (int, error) {
	count, err := m.queries.CountPlayerRelationships(ctx, sqlc.CountPlayerRelationshipsParams{
		PlayerID:         playerID,
		RelationshipType: string(relType),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count relationships: %w", err)
	}
	return int(count), nil
}