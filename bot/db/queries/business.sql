-- name: CreateBusiness :one
INSERT INTO businesses (owner_id, business_type, name, city_plot_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBusinessesByOwner :many
SELECT * FROM businesses
WHERE owner_id = $1 AND status = 'ACTIVE';

-- name: UpdateInventory :exec
UPDATE businesses SET inventory = $1
WHERE id = $2;

-- name: UpdateProductionConfig :exec
UPDATE businesses SET production_config = $1
WHERE id = $2;

-- name: CreateEmployment :one
INSERT INTO employment (business_id, employee_id, wage_rate, min_daily_hours)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: TerminateEmployment :exec
UPDATE employment SET status = 'TERMINATED'
WHERE id = $1;

-- name: UpsertDailyLabor :one
INSERT INTO daily_labor (player_id, business_id, economic_day, hours_worked)
VALUES ($1, $2, $3, $4)
ON CONFLICT (player_id, business_id, economic_day) 
DO UPDATE SET hours_worked = daily_labor.hours_worked + $4, updated_at = NOW()
RETURNING *;

-- name: GetDailyLaborByDay :many
SELECT * FROM daily_labor
WHERE economic_day = $1;