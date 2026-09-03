CREATE TYPE listing_status AS ENUM ('ACTIVE', 'SOLD', 'EXPIRED', 'CANCELLED');

CREATE TABLE market_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id UUID NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    item_type VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    quantity_remaining INTEGER NOT NULL CHECK (quantity_remaining >= 0),
    asking_price NUMERIC(20, 4) NOT NULL CHECK (asking_price > 0),
    escrow_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    status listing_status NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_market_listings_seller ON market_listings(seller_id);
CREATE INDEX idx_market_listings_status_item ON market_listings(status, item_type);

CREATE TABLE market_trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID NOT NULL REFERENCES market_listings(id) ON DELETE RESTRICT,
    buyer_id UUID NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    seller_id UUID NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    item_type VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price_per_unit NUMERIC(20, 4) NOT NULL CHECK (price_per_unit > 0),
    traded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_market_trades_item ON market_trades(item_type);
CREATE INDEX idx_market_trades_traded_at ON market_trades(traded_at);