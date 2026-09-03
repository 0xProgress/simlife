CREATE TYPE account_type AS ENUM ('WALLET', 'BANK', 'ESCROW', 'TREASURY');

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID REFERENCES players(id) ON DELETE RESTRICT,
    type account_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(player_id, type)
);

CREATE INDEX idx_accounts_player_id ON accounts(player_id);