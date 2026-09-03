CREATE TYPE entry_type AS ENUM ('DEBIT', 'CREDIT');
CREATE TYPE transaction_type AS ENUM (
    'PLAYER_TRANSFER', 'WAGE', 'TAX', 'MARKET_SALE', 'MARKET_PURCHASE', 
    'ESCROW_LOCK', 'ESCROW_RELEASE', 'BUSINESS_REVENUE', 'PROPERTY_TAX', 'INITIAL_GRANT'
);

-- Append-only ledger. No UPDATE or DELETE operations are permitted.
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount NUMERIC(20, 4) NOT NULL CHECK (amount > 0),
    entry_type entry_type NOT NULL,
    transaction_type transaction_type NOT NULL,
    reference_id UUID, -- Correlates the two sides of a double-entry transaction
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_reference_id ON transactions(reference_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);