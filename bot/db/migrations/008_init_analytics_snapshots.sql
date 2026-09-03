CREATE TYPE review_status AS ENUM ('OPEN', 'REVIEWED', 'DISMISSED', 'ACTIONED');

CREATE TABLE anomaly_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_type VARCHAR(100) NOT NULL,
    implicated_player_ids UUID[] NOT NULL DEFAULT '{}',
    evidence_summary TEXT,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    review_status review_status NOT NULL DEFAULT 'OPEN'
);

CREATE TABLE admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_player_id UUID NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    target_player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    action_type VARCHAR(100) NOT NULL,
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);