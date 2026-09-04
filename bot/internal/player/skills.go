package player

import (
	"context"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

// SkillType represents a profession skill category.
type SkillType string

const (
	SkillPhysical   SkillType = "PHYSICAL"
	SkillService    SkillType = "SERVICE"
	SkillTechnology SkillType = "TECHNOLOGY"
	SkillBusiness   SkillType = "BUSINESS"
	SkillFinance    SkillType = "FINANCE"
	SkillLegal      SkillType = "LEGAL"
	SkillPolitical  SkillType = "POLITICAL"
	SkillCriminal   SkillType = "CRIMINAL"
)

// Skill represents a player's proficiency in a profession.
type Skill struct {
	PlayerID    string
	SkillType   SkillType
	Level       int
	Experience  int
	NextLevelXP int
}

// SkillManager handles skill progression and experience tracking.
type SkillManager struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewSkillManager initializes the skill manager.
func NewSkillManager(q *sqlc.Queries, log zerolog.Logger) *SkillManager {
	return &SkillManager{
		queries: q,
		log:     log.With().Str("component", "skills").Logger(),
	}
}

// AddExperience increments a player's skill experience and handles level-ups.
// Returns true if the skill leveled up.
func (m *SkillManager) AddExperience(ctx context.Context, playerID string, skillType SkillType, xp int) (bool, error) {
	log := logger.FromContext(ctx, "player.skills")

	if xp <= 0 {
		return false, nil
	}

	// Fetch current skill state
	skill, err := m.queries.GetPlayerSkill(ctx, sqlc.GetPlayerSkillParams{
		PlayerID:  playerID,
		SkillType: string(skillType),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Skill doesn't exist yet, create it
			skill, err = m.queries.CreatePlayerSkill(ctx, sqlc.CreatePlayerSkillParams{
				PlayerID:   playerID,
				SkillType:  string(skillType),
				Level:      1,
				Experience: xp,
			})
			if err != nil {
				return false, fmt.Errorf("failed to create skill: %w", err)
			}
			log.Info().
				Str("player_id", playerID).
				Str("skill", string(skillType)).
				Int("xp", xp).
				Msg("skill initialized")
			return false, nil
		}
		return false, fmt.Errorf("failed to fetch skill: %w", err)
	}

	// Add experience
	newXP := int(skill.Experience) + xp
	newLevel := int(skill.Level)
	leveledUp := false

	// Check for level-up (each level requires 100 * level XP)
	xpRequired := calculateXPForLevel(newLevel + 1)
	for newXP >= xpRequired {
		newXP -= xpRequired
		newLevel++
		leveledUp = true
		xpRequired = calculateXPForLevel(newLevel + 1)
	}

	// Update skill in database
	err = m.queries.UpdatePlayerSkill(ctx, sqlc.UpdatePlayerSkillParams{
		PlayerID:   playerID,
		SkillType:  string(skillType),
		Level:      int32(newLevel),
		Experience: int32(newXP),
	})
	if err != nil {
		return false, fmt.Errorf("failed to update skill: %w", err)
	}

	if leveledUp {
		log.Info().
			Str("player_id", playerID).
			Str("skill", string(skillType)).
			Int("new_level", newLevel).
			Msg("skill leveled up")
	}

	return leveledUp, nil
}

// GetSkill fetches a player's skill level and experience.
func (m *SkillManager) GetSkill(ctx context.Context, playerID string, skillType SkillType) (*Skill, error) {
	row, err := m.queries.GetPlayerSkill(ctx, sqlc.GetPlayerSkillParams{
		PlayerID:  playerID,
		SkillType: string(skillType),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default skill state if not initialized
			return &Skill{
				PlayerID:    playerID,
				SkillType:   skillType,
				Level:       1,
				Experience:  0,
				NextLevelXP: calculateXPForLevel(2),
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch skill: %w", err)
	}

	return &Skill{
		PlayerID:    row.PlayerID,
		SkillType:   SkillType(row.SkillType),
		Level:       int(row.Level),
		Experience:  int(row.Experience),
		NextLevelXP: calculateXPForLevel(int(row.Level) + 1),
	}, nil
}

// GetAllSkills fetches all skills for a player.
func (m *SkillManager) GetAllSkills(ctx context.Context, playerID string) ([]*Skill, error) {
	rows, err := m.queries.GetPlayerSkills(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch skills: %w", err)
	}

	skills := make([]*Skill, 0, len(rows))
	for _, row := range rows {
		skills = append(skills, &Skill{
			PlayerID:    row.PlayerID,
			SkillType:   SkillType(row.SkillType),
			Level:       int(row.Level),
			Experience:  int(row.Experience),
			NextLevelXP: calculateXPForLevel(int(row.Level) + 1),
		})
	}

	return skills, nil
}

// HasSkillRequirement checks if a player meets a minimum skill level requirement.
func (m *SkillManager) HasSkillRequirement(ctx context.Context, playerID string, skillType SkillType, requiredLevel int) (bool, error) {
	skill, err := m.GetSkill(ctx, playerID, skillType)
	if err != nil {
		return false, err
	}
	return skill.Level >= requiredLevel, nil
}

// calculateXPForLevel returns the XP required to reach a given level.
// Formula: 100 * level (linear progression)
func calculateXPForLevel(level int) int {
	if level <= 1 {
		return 100
	}
	return 100 * level
}