-- name: CreateProperty :one
INSERT INTO properties (x_coord, y_coord, zone_type, assessed_value)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPropertyByCoords :one
SELECT * FROM properties
WHERE x_coord = $1 AND y_coord = $2;

-- name: UpdatePropertyOwner :exec
-- Fixed: owner_id instead of owner_player_id
UPDATE properties SET owner_id = $1
WHERE id = $2;

-- name: UpdateAssessedValue :exec
UPDATE properties SET assessed_value = $1
WHERE id = $2;

-- name: RecordTaxPayment :exec
UPDATE properties SET last_tax_paid_at = NOW()
WHERE id = $1;

-- name: GetPropertyByID :one
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT id, x_coord, y_coord, zone_type, owner_id, development_level, assessed_value, last_tax_paid_at, created_at, updated_at
FROM properties
WHERE id = @id;

-- name: GetPropertiesByOwner :many
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT id, x_coord, y_coord, zone_type, owner_id, development_level, assessed_value, last_tax_paid_at, created_at, updated_at
FROM properties
WHERE owner_id = @owner_player_id;

-- name: GetAvailableProperties :many
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT id, x_coord, y_coord, zone_type, owner_id, development_level, assessed_value, last_tax_paid_at, created_at, updated_at
FROM properties
WHERE owner_id IS NULL
LIMIT $1;

-- name: TransferPropertyOwnership :exec
-- Fixed: owner_player_id → owner_id
UPDATE properties
SET owner_id = @owner_player_id, updated_at = NOW()
WHERE id = @id;

-- name: UpgradePropertyDevelopment :exec
-- Fixed: owner_player_id → owner_id
UPDATE properties
SET development_level = development_level + 1, 
    assessed_value = assessed_value * 1.25,
    updated_at = NOW()
WHERE id = @id AND owner_id = @owner_player_id;

-- name: SetPropertyRent :exec
-- For Layer 5, rent is calculated dynamically based on property value (1% per month)
-- This query is a placeholder for future rental agreement tracking
-- Fixed: owner_player_id → owner_id
UPDATE properties
SET updated_at = NOW()
WHERE id = @id AND owner_id = @owner_player_id;

-- name: GetTaxableProperties :many
-- Fetches all properties that are currently owned by a player and are subject to taxation.
-- Uses explicit column selection to prevent sqlc struct brittleness.
-- Fixed: owner_player_id → owner_id, last_tax_payment → last_tax_paid_at
SELECT 
    id, 
    x_coord, 
    y_coord, 
    zone_type, 
    owner_id, 
    development_level, 
    assessed_value, 
    last_tax_paid_at
FROM properties
WHERE owner_id IS NOT NULL;

-- name: UpdatePropertyTaxPayment :exec
-- Updates the last tax payment timestamp for a property to the current time.
-- This is called immediately after a successful ledger transfer to maintain audit consistency.
-- Fixed: last_tax_payment → last_tax_paid_at
UPDATE properties
SET last_tax_paid_at = NOW(), updated_at = NOW()
WHERE id = @id;