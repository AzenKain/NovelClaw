-- Schema migration for NovelClaw Agentic Chat & Activity Stream
CREATE TABLE IF NOT EXISTS novelclaw_threads (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    is_main INTEGER NOT NULL DEFAULT 0,
    volume_index INTEGER NOT NULL DEFAULT 0,
    chapter_index INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_novelclaw_threads_project ON novelclaw_threads(project_id);

CREATE TABLE IF NOT EXISTS novelclaw_messages (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    sender TEXT NOT NULL,                  -- 'user' | 'novelclaw' | 'system' | 'agent_stream'
    role TEXT NOT NULL,                    -- 'user' | 'assistant' | 'system' | 'tool'
    content TEXT NOT NULL,
    thinking_content TEXT,
    step_type TEXT,                        -- 'fts5_search', 'web_lookup', 'action_dispatched', etc.
    step_status TEXT,                      -- 'running' | 'done' | 'error'
    action_call_json TEXT,                 -- JSON of function call and args/results
    is_collapsed INTEGER NOT NULL DEFAULT 1,
    is_archived_compact INTEGER NOT NULL DEFAULT 0,
    token_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(thread_id) REFERENCES novelclaw_threads(id) ON DELETE CASCADE,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_novelclaw_messages_thread ON novelclaw_messages(thread_id, created_at);
CREATE INDEX IF NOT EXISTS idx_novelclaw_messages_project ON novelclaw_messages(project_id);
