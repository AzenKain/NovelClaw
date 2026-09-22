-- Migration 011: Add style_guide and style_name to projects
ALTER TABLE projects ADD COLUMN style_guide TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN style_name TEXT NOT NULL DEFAULT '';
