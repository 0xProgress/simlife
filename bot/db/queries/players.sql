-- name: GetPlayerByDiscordID :one
SELECT * FROM players
WHERE discord_id = $1 AND is_deleted = false;

-- name: CreatePlayer :one
INSERT INTO players (discord_id, username, display_color)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateLastActive :exec
UPDATE players SET last_active_at = NOW()
WHERE id = $1;

-- name: SoftDeletePlayer :exec
UPDATE players SET is_deleted = true
WHERE id = $1;