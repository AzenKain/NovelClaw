-- name: UpsertGlossaryTerm :exec
INSERT INTO glossary (id, project_id, source_term, target_term, category, notes)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, source_term) DO UPDATE SET
    target_term = excluded.target_term,
    category = excluded.category,
    notes = excluded.notes;

-- name: GetGlossaryTerm :one
SELECT * FROM glossary
WHERE project_id = ? AND source_term = ?
LIMIT 1;

-- name: ListGlossaryByProject :many
SELECT * FROM glossary 
WHERE project_id = ? 
ORDER BY category ASC, source_term ASC;

-- name: DeleteGlossaryTerm :exec
DELETE FROM glossary WHERE id = ?;
