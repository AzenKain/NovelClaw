-- Migration 010: Add reasoning_effort to llm_configs
ALTER TABLE llm_configs ADD COLUMN reasoning_effort TEXT NOT NULL DEFAULT 'off';
