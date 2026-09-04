-- name: CreateBusiness :one
-- Creates a new business with JSONB inventory and production configuration.
INSERT INTO businesses (owner_id, business_type, name, city_plot_id, inventory, production_config, status, opened_at)
VALUES (@owner_id, @business_type, @name, @city_plot_id, @inventory::jsonb, @production_config::jsonb, 'ACTIVE', NOW())
RETURNING id, owner_id, business_type, name, city_plot_id, status, inventory, production_config, opened_at;

-- name: GetBusinessByID :one
-- Fetches a specific business by ID. Includes JSONB fields for inventory and production config.
SELECT id, owner_id, business_type, name, city_plot_id, status, inventory, production_config, opened_at
FROM businesses
WHERE id = @id;

-- name: GetBusinessesByOwner :many
-- Fetches all active businesses owned by a specific player.
SELECT id, owner_id, business_type, name, city_plot_id, status, inventory, production_config, opened_at
FROM businesses
WHERE owner_id = @owner_id AND status = 'ACTIVE';

-- name: GetActiveBusinesses :many
-- Fetches all active businesses in the system. Used by the settlement engine for daily processing.
SELECT id, owner_id, business_type, name, city_plot_id, status, inventory, production_config, opened_at
FROM businesses
WHERE status = 'ACTIVE';

-- name: UpdateBusinessStatus :exec
-- Updates the status of a business (e.g., to 'CLOSED' or 'SUSPENDED').
UPDATE businesses 
SET status = @status
WHERE id = @id;

-- name: UpdateInventory :exec
-- Updates the business inventory JSONB field. Used by the production engine after processing.
UPDATE businesses 
SET inventory = @inventory::jsonb
WHERE id = @id;

-- name: UpdateProductionConfig :exec
-- Updates the business production configuration JSONB field.
UPDATE businesses 
SET production_config = @production_config::jsonb
WHERE id = @id;

-- name: MarkBusinessOperated :exec
-- Updates the business timestamp to indicate it has been manually opened/operated for the current game day.
-- The settlement engine will read this to process daily revenue and payroll.
UPDATE businesses 
SET opened_at = NOW()
WHERE id = @id;

-- name: CreateEmployment :one
-- Creates a new employment relationship between a business and a player.
INSERT INTO employment (business_id, employee_id, wage_rate, min_daily_hours, status, started_at)
VALUES (@business_id, @employee_id, @wage_rate, @min_daily_hours, 'ACTIVE', NOW())
RETURNING id, business_id, employee_id, wage_rate, min_daily_hours, status, started_at;

-- name: GetActiveEmployment :one
-- Fetches the player's current active employment record and the associated business details.
-- Required by the /work command to determine if a player is employed and by whom.
SELECT e.id, e.business_id, e.wage_rate, b.name AS business_name
FROM employment e
JOIN businesses b ON e.business_id = b.id
WHERE e.employee_id = @employee_id AND e.status = 'ACTIVE'
LIMIT 1;

-- name: GetEmploymentByBusiness :many
-- Fetches all active employees for a specific business (used by business owners and settlement engine).
SELECT e.id, e.employee_id, e.wage_rate, e.min_daily_hours, e.status, e.started_at
FROM employment e
WHERE e.business_id = @business_id AND e.status = 'ACTIVE';

-- name: TerminateEmployment :exec
-- Terminates an employment relationship by setting status to 'TERMINATED'.
UPDATE employment 
SET status = 'TERMINATED'
WHERE id = @id;

-- name: UpsertDailyLabor :one
-- Logs work hours for a player at a specific business for the current economic day.
-- Uses ON CONFLICT to safely accumulate hours if the player works multiple shifts in the same day.
-- CRITICAL: Uses EXCLUDED.hours_worked to add the *new* attempted value, not the old one.
INSERT INTO daily_labor (player_id, business_id, economic_day, hours_worked, created_at, updated_at)
VALUES (@player_id, @business_id, @economic_day, @hours_worked, NOW(), NOW())
ON CONFLICT (player_id, business_id, economic_day) 
DO UPDATE SET 
    hours_worked = daily_labor.hours_worked + EXCLUDED.hours_worked,
    updated_at = NOW()
RETURNING id, player_id, business_id, economic_day, hours_worked, created_at, updated_at;

-- name: GetDailyLaborByDay :many
-- Fetches all labor records for a specific economic day. Used by the settlement engine.
SELECT id, player_id, business_id, economic_day, hours_worked, created_at, updated_at
FROM daily_labor
WHERE economic_day = @economic_day;

-- name: GetDailyLaborByPlayerAndDay :one
-- Fetches a specific player's labor record for a specific day.
SELECT id, player_id, business_id, economic_day, hours_worked, created_at, updated_at
FROM daily_labor
WHERE player_id = @player_id AND economic_day = @economic_day;

-- name: GetDailyLaborByBusiness :many
-- Fetches all labor records for a specific business. Used by the production engine to calculate output.
SELECT id, player_id, business_id, economic_day, hours_worked, created_at, updated_at
FROM daily_labor
WHERE business_id = @business_id;

-- name: GetDailyLaborByEmployeeAndBusiness :one
-- Fetches a specific employee's labor record at a specific business. Used by the wage engine.
SELECT id, player_id, business_id, economic_day, hours_worked, created_at, updated_at
FROM daily_labor
WHERE player_id = @player_id AND business_id = @business_id;

-- name: ClearDailyLaborForBusiness :exec
-- Clears all daily labor records for a business after settlement has processed wages.
-- This prepares the table for the next economic day.
DELETE FROM daily_labor
WHERE business_id = @business_id;

-- name: GetBusinessAccount :one
-- Fetches a dedicated business account if it exists (Layer 8+ feature).
-- Falls back to the owner's wallet if no dedicated business account exists.
-- Fixed: account_type → type
SELECT id, player_id, type AS account_type, created_at
FROM accounts
WHERE player_id = @business_id AND type = 'BUSINESS';