-- Migration 009: Universal World Bible & Lorebook Categories/Entries
CREATE TABLE IF NOT EXISTS world_categories (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT 'Boxes',
    description TEXT NOT NULL DEFAULT '',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, slug)
);

CREATE TABLE IF NOT EXISTS world_entries (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    category_id TEXT NOT NULL,
    name TEXT NOT NULL,
    aliases_json TEXT NOT NULL DEFAULT '[]',
    summary TEXT NOT NULL DEFAULT '',
    full_description TEXT NOT NULL DEFAULT '',
    attributes_json TEXT NOT NULL DEFAULT '{}',
    discovered_by TEXT NOT NULL DEFAULT 'ai_scan',
    source_chapter_index INTEGER NOT NULL DEFAULT 1,
    is_verified INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(category_id) REFERENCES world_categories(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_world_categories_project ON world_categories(project_id);
CREATE INDEX IF NOT EXISTS idx_world_entries_project ON world_entries(project_id);
CREATE INDEX IF NOT EXISTS idx_world_entries_category ON world_entries(category_id);
CREATE INDEX IF NOT EXISTS idx_world_entries_verified ON world_entries(project_id, is_verified);
