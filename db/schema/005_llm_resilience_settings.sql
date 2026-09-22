-- Migration 005: Add max_retries and timeout_seconds to llm_configs
ALTER TABLE llm_configs ADD COLUMN max_retries INTEGER NOT NULL DEFAULT 3;
ALTER TABLE llm_configs ADD COLUMN timeout_seconds INTEGER NOT NULL DEFAULT 120;
