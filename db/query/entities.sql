-- name: UpsertEntity :exec
INSERT INTO entities (
    id, project_id, name, aliases_json, category, gender, role, first_seen_chapter, metadata_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, name) DO UPDATE SET
    aliases_json = excluded.aliases_json,
    category = excluded.category,
    gender = excluded.gender,
    role = excluded.role,
    first_seen_chapter = MIN(entities.first_seen_chapter, excluded.first_seen_chapter),
    metadata_json = excluded.metadata_json,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetEntityByName :one
SELECT * FROM entities
WHERE project_id = ? AND name = ?
LIMIT 1;

-- name: ListEntitiesByProject :many
SELECT * FROM entities
WHERE project_id = ?
ORDER BY first_seen_chapter ASC, name ASC;

-- name: ListActiveEntitiesAtChapter :many
SELECT * FROM entities
WHERE project_id = ? AND first_seen_chapter <= ?
ORDER BY first_seen_chapter ASC;

-- name: DeleteEntity :exec
DELETE FROM entities WHERE id = ?;
