-- Migration 006: Dynamic Progressive Entity Graph (L2 Nodes)
CREATE TABLE IF NOT EXISTS entities (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    aliases_json TEXT NOT NULL DEFAULT '[]',
    category TEXT NOT NULL DEFAULT 'character', -- character, faction, location, item
    gender TEXT NOT NULL DEFAULT 'unknown',      -- male, female, other, unknown
    role TEXT NOT NULL DEFAULT '',               -- protagonist, antagonist, supporting, etc.
    first_seen_chapter INTEGER NOT NULL DEFAULT 1,
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_entities_project ON entities(project_id);
CREATE INDEX IF NOT EXISTS idx_entities_category ON entities(project_id, category);
CREATE INDEX IF NOT EXISTS idx_entities_timeline ON entities(project_id, first_seen_chapter);
