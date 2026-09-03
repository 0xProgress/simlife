-- name: CreateListing :one
INSERT INTO market_listings (seller_id, item_type, quantity, quantity_remaining, asking_price, escrow_account_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetActiveListingsByItem :many
SELECT * FROM market_listings
WHERE item_type = $1 AND status = 'ACTIVE'
ORDER BY asking_price ASC, created_at ASC
LIMIT $2 OFFSET $3;

-- name: UpdateListingStatus :exec
UPDATE market_listings SET status = $1, updated_at = NOW()
WHERE id = $2;

-- name: DecrementListingQuantity :exec
UPDATE market_listings 
SET quantity_remaining = quantity_remaining - $1, updated_at = NOW()
WHERE id = $2 AND quantity_remaining >= $1;

-- name: RecordTrade :one
INSERT INTO market_trades (listing_id, buyer_id, seller_id, item_type, quantity, price_per_unit)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRecentTradesByItem :many
SELECT * FROM market_trades
WHERE item_type = $1
ORDER BY traded_at DESC
LIMIT $2;