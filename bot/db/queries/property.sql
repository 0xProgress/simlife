-- name: CreateProperty :one
INSERT INTO properties (x_coord, y_coord, zone_type, assessed_value)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPropertyByCoords :one
SELECT * FROM properties
WHERE x_coord = $1 AND y_coord = $2;

-- name: UpdatePropertyOwner :exec
UPDATE properties SET owner_id = $1
WHERE id = $2;

-- name: UpdateAssessedValue :exec
UPDATE properties SET assessed_value = $1
WHERE id = $2;

-- name: RecordTaxPayment :exec
UPDATE properties SET last_tax_paid_at = NOW()
WHERE id = $1;