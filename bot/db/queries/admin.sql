-- name: IsPlayerAdmin :one
-- Checks if a player has admin privileges. In Layer 1, this is a hardcoded list or a flag.
-- For production, this should query an admin_roles table.
SELECT EXISTS(
    SELECT 1 FROM players 
    WHERE id = @player_id AND discord_id IN ('YOUR_DISCORD_ID_HERE')
) AS is_admin;

-- name: LogAdminAction :exec
-- Logs an admin action to the audit trail. This is append-only and cannot be modified.
INSERT INTO admin_audit_log (admin_player_id, target_player_id, action_type, parameters, created_at)
VALUES (@admin_player_id, @target_player_id, @action_type, @parameters, NOW());

-- name: ResetPlayerEconomicState :exec
-- Completely resets a player's economic state. Wipes employment, business ownership, and property records.
-- Note: This does NOT delete the player record or their accounts, only their economic activity.
UPDATE employment SET status = 'TERMINATED' WHERE employee_id = @player_id AND status = 'ACTIVE';
UPDATE businesses SET status = 'CLOSED' WHERE owner_id = @player_id AND status = 'ACTIVE';
-- Fixed: owner_player_id → owner_id
UPDATE properties SET owner_id = NULL WHERE owner_id = @player_id;

-- name: GetPlayerTransactionCount :one
-- Returns the total number of transactions involving a player's accounts.
SELECT COUNT(*)::bigint AS transaction_count
FROM transactions t
JOIN accounts a ON t.account_id = a.id
WHERE a.player_id = @player_id;

-- name: GetPlayerBusinessCount :one
-- Returns the number of businesses owned by a player.
SELECT COUNT(*)::integer AS business_count
FROM businesses
WHERE owner_id = @player_id AND status = 'ACTIVE';

-- name: GetPlayerPropertyCount :one
-- Returns the number of properties owned by a player.
-- Fixed: owner_player_id → owner_id
SELECT COUNT(*)::integer AS property_count
FROM properties
WHERE owner_id = @player_id;