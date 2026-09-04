-- name: GetPlayerByDiscordID :one
-- Fixed: created_at → registered_at
SELECT id, discord_id, username, display_color, registered_at, last_active_at, is_deleted 
FROM players
WHERE discord_id = $1 AND is_deleted = false;

-- name: GetPlayerByID :one
-- Fixed: created_at → registered_at
SELECT id, discord_id, username, display_color, registered_at, last_active_at, is_deleted 
FROM players
WHERE id = $1 AND is_deleted = false;

-- name: CreatePlayer :one
INSERT INTO players (discord_id, username, display_color, registered_at, last_active_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, discord_id, username, display_color, registered_at, last_active_at, is_deleted;

-- name: GetOrCreatePlayer :one
-- Idempotent player creation. Updates username and un-deletes if the user returns.
INSERT INTO players (discord_id, username, display_color, registered_at, last_active_at)
VALUES (@discord_id, @username, '#FFFFFF', NOW(), NOW())
ON CONFLICT (discord_id) DO UPDATE 
SET username = EXCLUDED.username, 
    last_active_at = NOW(), 
    is_deleted = false
RETURNING id, discord_id, username, display_color, registered_at, last_active_at, is_deleted;

-- name: UpdateLastActive :exec
UPDATE players SET last_active_at = NOW()
WHERE id = $1;

-- name: SoftDeletePlayer :exec
UPDATE players SET is_deleted = true
WHERE id = $1;

-- name: EnsurePlayerAccounts :exec
-- Provisions the three mandatory double-entry ledger accounts for a player.
-- Uses ON CONFLICT DO NOTHING to ensure this is safe to call on every interaction.
-- Fixed: account_type → type
INSERT INTO accounts (player_id, type, created_at)
VALUES 
    (@player_id, 'WALLET', NOW()),
    (@player_id, 'BANK', NOW()),
    (@player_id, 'ESCROW', NOW())
ON CONFLICT (player_id, type) DO NOTHING;

-- name: CreatePlayerSkill :one
INSERT INTO player_skills (player_id, skill_type, level, experience)
VALUES (@player_id, @skill_type, @level, @experience)
RETURNING player_id, skill_type, level, experience;

-- name: GetPlayerSkill :one
SELECT player_id, skill_type, level, experience
FROM player_skills
WHERE player_id = @player_id AND skill_type = @skill_type;

-- name: GetPlayerSkills :many
SELECT player_id, skill_type, level, experience
FROM player_skills
WHERE player_id = @player_id;

-- name: UpdatePlayerSkill :exec
UPDATE player_skills
SET level = @level, experience = @experience
WHERE player_id = @player_id AND skill_type = @skill_type;

-- name: CreatePlayerReputation :one
INSERT INTO player_reputations (player_id, reputation_type, score)
VALUES (@player_id, @reputation_type, @score)
RETURNING player_id, reputation_type, score;

-- name: GetPlayerReputation :one
SELECT player_id, reputation_type, score
FROM player_reputations
WHERE player_id = @player_id AND reputation_type = @reputation_type;

-- name: GetPlayerReputations :many
SELECT player_id, reputation_type, score
FROM player_reputations
WHERE player_id = @player_id;

-- name: UpdatePlayerReputation :exec
UPDATE player_reputations
SET score = @score
WHERE player_id = @player_id AND reputation_type = @reputation_type;

-- name: CreateRelationship :one
INSERT INTO player_relationships (player_id, related_player_id, relationship_type, metadata)
VALUES (@player_id, @related_player_id, @relationship_type, @metadata)
RETURNING id, player_id, related_player_id, relationship_type, metadata, created_at;

-- name: GetPlayerRelationships :many
SELECT id, player_id, related_player_id, relationship_type, metadata, created_at
FROM player_relationships
WHERE player_id = @player_id AND terminated_at IS NULL;

-- name: GetPlayerRelationshipsByType :many
SELECT id, player_id, related_player_id, relationship_type, metadata, created_at
FROM player_relationships
WHERE player_id = @player_id AND relationship_type = @relationship_type AND terminated_at IS NULL;

-- name: TerminateRelationship :exec
UPDATE player_relationships
SET terminated_at = NOW()
WHERE id = @id;

-- name: CountPlayerRelationships :one
SELECT COUNT(*)::bigint
FROM player_relationships
WHERE player_id = @player_id AND relationship_type = @relationship_type AND terminated_at IS NULL;