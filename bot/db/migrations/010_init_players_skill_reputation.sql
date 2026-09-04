-- Create player_skills table for skill progression tracking
CREATE TABLE IF NOT EXISTS player_skills (
    player_id TEXT NOT NULL,
    skill_type TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    experience INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (player_id, skill_type)
);

CREATE INDEX IF NOT EXISTS idx_player_skills_player ON player_skills(player_id);

-- Create player_reputations table for reputation tracking
CREATE TABLE IF NOT EXISTS player_reputations (
    player_id TEXT NOT NULL,
    reputation_type TEXT NOT NULL,
    score INTEGER NOT NULL DEFAULT 500,
    PRIMARY KEY (player_id, reputation_type)
);

CREATE INDEX IF NOT EXISTS idx_player_reputations_player ON player_reputations(player_id);

-- Create player_relationships table for relationship history
CREATE TABLE IF NOT EXISTS player_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id TEXT NOT NULL,
    related_player_id TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    terminated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_player_relationships_player ON player_relationships(player_id);
CREATE INDEX IF NOT EXISTS idx_player_relationships_related ON player_relationships(related_player_id);
CREATE INDEX IF NOT EXISTS idx_player_relationships_type ON player_relationships(relationship_type);
CREATE INDEX IF NOT EXISTS idx_player_relationships_active ON player_relationships(player_id, terminated_at) WHERE terminated_at IS NULL;