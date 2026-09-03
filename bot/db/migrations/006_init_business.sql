CREATE TYPE business_status AS ENUM ('ACTIVE', 'CLOSED', 'SUSPENDED');
CREATE TYPE employment_status AS ENUM ('ACTIVE', 'TERMINATED');

CREATE TABLE businesses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    business_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    city_plot_id UUID REFERENCES properties(id) ON DELETE SET NULL,
    status business_status NOT NULL DEFAULT 'ACTIVE',
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inventory JSONB NOT NULL DEFAULT '{}'::jsonb,
    production_config JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_businesses_owner ON businesses(owner_id);

CREATE TABLE employment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    wage_rate NUMERIC(20, 4) NOT NULL CHECK (wage_rate >= 0),
    min_daily_hours NUMERIC(5, 2) NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status employment_status NOT NULL DEFAULT 'ACTIVE'
);

CREATE TABLE daily_labor (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE RESTRICT,
    economic_day INTEGER NOT NULL,
    hours_worked NUMERIC(5, 2) NOT NULL DEFAULT 0 CHECK (hours_worked >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(player_id, business_id, economic_day)
);