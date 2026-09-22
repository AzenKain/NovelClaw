-- name: CreateChapter :exec
INSERT INTO chapters (id, project_id, chapter_index, title, content_path, raw_content, status)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetChapter :one
SELECT * FROM chapters WHERE id = ? LIMIT 1;

-- name: GetChapterByIndex :one
SELECT * FROM chapters WHERE project_id = ? AND chapter_index = ? LIMIT 1;

-- name: ListChaptersByProject :many
SELECT * FROM chapters WHERE project_id = ? ORDER BY chapter_index ASC;

-- name: UpdateChapterTranslation :exec
UPDATE chapters 
SET translated_content = ?, status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: CountChaptersByStatus :many
SELECT status, COUNT(*) as count 
FROM chapters 
WHERE project_id = ? 
GROUP BY status;
