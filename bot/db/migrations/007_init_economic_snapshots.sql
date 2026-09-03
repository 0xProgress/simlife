CREATE TABLE economic_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    economic_day INTEGER NOT NULL UNIQUE,
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);