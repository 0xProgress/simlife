CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    x_coord INTEGER NOT NULL,
    y_coord INTEGER NOT NULL,
    zone_type VARCHAR(50) NOT NULL,
    owner_id UUID REFERENCES players(id) ON DELETE SET NULL,
    development_level INTEGER NOT NULL DEFAULT 0,
    assessed_value NUMERIC(20, 4) NOT NULL DEFAULT 0,
    last_tax_paid_at TIMESTAMPTZ,
    UNIQUE(x_coord, y_coord)
);

CREATE INDEX idx_properties_owner ON properties(owner_id);