-- name: UpsertRelation :exec
INSERT INTO character_relations (id, project_id, from_char, to_char, call_as, self_call_as, since_chapter, tone, is_locked)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    call_as = excluded.call_as,
    self_call_as = excluded.self_call_as,
    since_chapter = excluded.since_chapter,
    tone = excluded.tone,
    is_locked = excluded.is_locked,
    updated_at = CURRENT_TIMESTAMP;

-- name: ListRelationsByProject :many
SELECT * FROM character_relations 
WHERE project_id = ? 
ORDER BY since_chapter ASC;

-- name: GetRelationBetween :one
SELECT * FROM character_relations 
WHERE project_id = ? AND from_char = ? AND to_char = ? AND since_chapter <= ?
ORDER BY since_chapter DESC 
LIMIT 1;

-- name: LockRelation :exec
UPDATE character_relations 
SET is_locked = 1, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteRelation :exec
DELETE FROM character_relations WHERE id = ?;

