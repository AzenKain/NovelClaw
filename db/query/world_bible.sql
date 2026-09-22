-- name: ListWorldCategories :many
SELECT * FROM world_categories
WHERE project_id = ?
ORDER BY display_order ASC, name ASC;

-- name: GetWorldCategoryBySlug :one
SELECT * FROM world_categories
WHERE project_id = ? AND slug = ?;

-- name: GetWorldCategoryByID :one
SELECT * FROM world_categories
WHERE id = ?;

-- name: UpsertWorldCategory :one
INSERT INTO world_categories (
    id, project_id, slug, name, icon, description, display_order, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    icon = excluded.icon,
    description = excluded.description,
    display_order = excluded.display_order
RETURNING *;

-- name: DeleteWorldCategory :exec
DELETE FROM world_categories
WHERE id = ?;

-- name: ListWorldEntries :many
SELECT * FROM world_entries
WHERE project_id = ?
ORDER BY updated_at DESC, name ASC;

-- name: ListWorldEntriesByCategory :many
SELECT * FROM world_entries
WHERE project_id = ? AND category_id = ?
ORDER BY updated_at DESC, name ASC;

-- name: GetWorldEntryByID :one
SELECT * FROM world_entries
WHERE id = ?;

-- name: GetWorldEntryByName :one
SELECT * FROM world_entries
WHERE project_id = ? AND LOWER(name) = sqlc.arg(lower_name);

-- name: UpsertWorldEntry :one
INSERT INTO world_entries (
    id, project_id, category_id, name, aliases_json, summary, full_description,
    attributes_json, discovered_by, source_chapter_index, is_verified, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT(id) DO UPDATE SET
    category_id = excluded.category_id,
    name = excluded.name,
    aliases_json = excluded.aliases_json,
    summary = excluded.summary,
    full_description = excluded.full_description,
    attributes_json = excluded.attributes_json,
    discovered_by = excluded.discovered_by,
    source_chapter_index = excluded.source_chapter_index,
    is_verified = excluded.is_verified,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: VerifyWorldEntry :exec
UPDATE world_entries
SET is_verified = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteWorldEntry :exec
DELETE FROM world_entries
WHERE id = ?;

-- name: CountWorldEntries :one
SELECT COUNT(*) FROM world_entries
WHERE project_id = ?;
