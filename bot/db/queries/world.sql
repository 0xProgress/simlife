-- name: RecordEconomicSnapshot :one
INSERT INTO economic_snapshots (economic_day, metrics)
VALUES ($1, $2)
RETURNING *;

-- name: CreateAnomalyFlag :one
INSERT INTO anomaly_flags (flag_type, implicated_player_ids, evidence_summary)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOpenAnomalies :many
SELECT * FROM anomaly_flags
WHERE review_status = 'OPEN'
ORDER BY detected_at DESC;

-- name: UpdateAnomalyStatus :exec
UPDATE anomaly_flags SET review_status = $1
WHERE id = $2;

-- name: LogAdminAction :one
INSERT INTO admin_audit_log (admin_player_id, target_player_id, action_type, parameters)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPlotByID :one
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT id, x_coord, y_coord, zone_type, owner_id, development_level, assessed_value, last_tax_paid_at, created_at, updated_at
FROM properties
WHERE id = @id;

-- name: GetAllPlots :many
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT p.id, p.x_coord, p.y_coord, p.zone_type, p.owner_id, p.development_level, p.assessed_value, p.last_tax_paid_at,
       pl.username AS owner_username
FROM properties p
LEFT JOIN players pl ON p.owner_id = pl.id
ORDER BY p.y_coord, p.x_coord;

-- name: GetAvailablePlots :many
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT id, x_coord, y_coord, zone_type, owner_id, development_level, assessed_value, last_tax_paid_at, created_at, updated_at
FROM properties
WHERE owner_id IS NULL
ORDER BY y_coord, x_coord
LIMIT $1;

-- name: GetAvailablePlotsByZone :many
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT id, x_coord, y_coord, zone_type, owner_id, development_level, assessed_value, last_tax_paid_at, created_at, updated_at
FROM properties
WHERE owner_id IS NULL AND zone_type = @zone_type
ORDER BY y_coord, x_coord
LIMIT $2;

-- name: GetPlotsByOwner :many
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT id, x_coord, y_coord, zone_type, owner_id, development_level, assessed_value, last_tax_paid_at, created_at, updated_at
FROM properties
WHERE owner_id = @owner_player_id
ORDER BY y_coord, x_coord;

-- name: TransferPlotOwnership :exec
-- Fixed: owner_player_id → owner_id
UPDATE properties
SET owner_id = @owner_player_id, updated_at = NOW()
WHERE id = @id AND owner_id IS NULL;

-- name: ReleasePlotOwnership :exec
-- Fixed: owner_player_id → owner_id
UPDATE properties
SET owner_id = NULL, updated_at = NOW()
WHERE id = @id;

-- name: UpgradePlotDevelopment :exec
-- Fixed: owner_player_id → owner_id
UPDATE properties
SET development_level = development_level + 1,
    assessed_value = assessed_value * 1.25,
    updated_at = NOW()
WHERE id = @id AND owner_id IS NOT NULL;

-- name: GetActiveBusinessesWithPlots :many
SELECT id, owner_id, business_type, name, city_plot_id, status
FROM businesses
WHERE status = 'ACTIVE' AND city_plot_id IS NOT NULL AND city_plot_id != '';

-- name: GetOnlinePlayers :many
-- Players active in the last 15 minutes
SELECT id, username, last_active_at
FROM players
WHERE last_active_at >= NOW() - INTERVAL '15 minutes'
  AND is_deleted = false
ORDER BY last_active_at DESC
LIMIT 100;

-- name: GetInfrastructureProjects :many
SELECT id, name, project_type, status, completion_percent, created_at, updated_at
FROM infrastructure_projects
WHERE status IN ('PLANNED', 'IN_PROGRESS')
ORDER BY created_at DESC;

-- name: GetDistrictSummary :one
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT 
    zone_type,
    COUNT(*)::integer AS total_plots,
    COUNT(CASE WHEN owner_id IS NOT NULL THEN 1 END)::integer AS owned_plots,
    COALESCE(AVG(development_level), 0)::numeric AS avg_development,
    COALESCE(SUM(assessed_value), 0)::numeric AS total_value,
    (SELECT COUNT(*)::integer FROM businesses b 
     JOIN properties p ON b.city_plot_id = p.id 
     WHERE p.zone_type = @zone_type AND b.status = 'ACTIVE') AS active_businesses
FROM properties
WHERE zone_type = @zone_type
GROUP BY zone_type;

-- name: RecordWorldEvent :exec
INSERT INTO world_events (event_id, event_name, event_type, economic_day, effect, flavor_text, expires_at, created_at)
VALUES (@event_id, @event_name, @event_type, @economic_day, @effect::jsonb, @flavor_text, @expires_at, NOW());

-- name: GetActiveWorldEvent :one
SELECT event_id, event_name, event_type, economic_day, effect, flavor_text, expires_at
FROM world_events
WHERE expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: GetAccountByID :one
-- Fetches a single account by ID. Used by the velocity monitor to resolve player IDs.
SELECT id, player_id, type AS account_type, created_at
FROM accounts
WHERE id = @id;

-- name: GetAnomalyFlagByID :one
-- Fetches a single anomaly flag by ID.
SELECT id, flag_type, implicated_player_ids, evidence_summary, detected_at, review_status
FROM anomaly_flags
WHERE id = @id;