-- Migration 008: Add skills_config_json column to project_settings table
ALTER TABLE project_settings ADD COLUMN skills_config_json TEXT NOT NULL DEFAULT '{}';
