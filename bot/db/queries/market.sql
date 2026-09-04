-- name: CreateListing :one
-- Creates a new active market listing. quantity_remaining is initialized to the total quantity.
INSERT INTO market_listings (
    seller_id, 
    item_type, 
    quantity, 
    quantity_remaining, 
    asking_price, 
    escrow_account_id, 
    status, 
    created_at, 
    updated_at
)
VALUES (
    @seller_id, 
    @item_type, 
    @quantity, 
    @quantity, 
    @asking_price::numeric, 
    @escrow_account_id, 
    'ACTIVE', 
    NOW(), 
    NOW()
)
RETURNING 
    id, 
    seller_id, 
    item_type, 
    quantity, 
    quantity_remaining, 
    asking_price, 
    escrow_account_id, 
    status, 
    created_at, 
    updated_at;

-- name: GetListingByID :one
-- Fetches a specific listing by ID. Critical for validating existence and status before executing a trade.
SELECT 
    id, 
    seller_id, 
    item_type, 
    quantity, 
    quantity_remaining, 
    asking_price, 
    escrow_account_id, 
    status, 
    created_at, 
    updated_at
FROM market_listings
WHERE id = @id;

-- name: GetActiveListingsByItem :many
-- Fetches active listings for a specific item, ordered by cheapest price first, then oldest listing.
SELECT 
    id, 
    seller_id, 
    item_type, 
    quantity, 
    quantity_remaining, 
    asking_price, 
    escrow_account_id, 
    status, 
    created_at, 
    updated_at
FROM market_listings
WHERE item_type = @item_type AND status = 'ACTIVE'
ORDER BY asking_price ASC, created_at ASC
LIMIT $1 OFFSET $2;

-- name: UpdateListingStatus :exec
-- Updates the status of a listing (e.g., to 'SOLD', 'EXPIRED', or 'CANCELLED').
UPDATE market_listings 
SET status = @status, updated_at = NOW()
WHERE id = @id;

-- name: DecrementListingQuantity :exec
-- Atomically decrements the remaining quantity of a listing. 
-- The WHERE clause ensures it only succeeds if sufficient quantity remains, preventing overselling.
UPDATE market_listings 
SET quantity_remaining = quantity_remaining - @quantity, updated_at = NOW()
WHERE id = @id AND quantity_remaining >= @quantity;

-- name: RecordTrade :one
-- Records a completed market trade. This is an append-only record used by the pricing engine and analytics.
INSERT INTO market_trades (
    listing_id, 
    buyer_id, 
    seller_id, 
    item_type, 
    quantity, 
    price_per_unit, 
    traded_at
)
VALUES (
    @listing_id, 
    @buyer_id, 
    @seller_id, 
    @item_type, 
    @quantity, 
    @price_per_unit::numeric, 
    NOW()
)
RETURNING 
    id, 
    listing_id, 
    buyer_id, 
    seller_id, 
    item_type, 
    quantity, 
    price_per_unit, 
    traded_at;

-- name: GetRecentTradesByItem :many
-- Fetches the most recent completed trades for a specific item to build price history charts.
SELECT 
    id, 
    listing_id, 
    buyer_id, 
    seller_id, 
    item_type, 
    quantity, 
    price_per_unit, 
    traded_at
FROM market_trades
WHERE item_type = @item_type
ORDER BY traded_at DESC
LIMIT sqlc.arg('limit');

-- name: GetPlayerInventoryQuantity :one
-- Checks how much of a specific item a player has in their inventory.
SELECT COALESCE(SUM(quantity), 0)::integer AS quantity
FROM player_inventory
WHERE player_id = @player_id AND item_type = @item_type;

-- name: TransferItemToEscrow :exec
-- Logically locks an item for a market listing by moving it to an escrow state.
UPDATE player_inventory
SET quantity = quantity - @quantity,
    escrow_reference = @escrow_ref,
    updated_at = NOW()
WHERE player_id = @player_id 
  AND item_type = @item_type 
  AND quantity >= @quantity;

-- name: UpdateListingBid :exec
-- Updates an auction-style listing with a new highest bid and the new bidder's ID.
UPDATE market_listings
SET asking_price = @asking_price::numeric,
    seller_id = @seller_id, -- Reused to track the current highest bidder in auction mode
    updated_at = NOW()
WHERE id = @id AND status = 'ACTIVE';