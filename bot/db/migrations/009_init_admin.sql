-- Create admin_audit_log table
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_player_id TEXT NOT NULL,
    target_player_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    parameters TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create world_events table for tracking daily economic events
CREATE TABLE IF NOT EXISTS world_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id TEXT NOT NULL,
    event_name TEXT NOT NULL,
    event_type TEXT NOT NULL,
    economic_day INTEGER NOT NULL,
    effect JSONB NOT NULL,
    flavor_text TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_world_events_expires ON world_events(expires_at);
CREATE INDEX IF NOT EXISTS idx_world_events_economic_day ON world_events(economic_day);

-- Create infrastructure_projects table for city development
CREATE TABLE IF NOT EXISTS infrastructure_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    project_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PLANNED',
    completion_percent INTEGER NOT NULL DEFAULT 0,
    funded_amount NUMERIC NOT NULL DEFAULT 0,
    total_cost NUMERIC NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_infrastructure_status ON infrastructure_projects(status);

-- Create index for efficient querying
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_admin ON admin_audit_log(admin_player_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_target ON admin_audit_log(target_player_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_created ON admin_audit_log(created_at DESC);