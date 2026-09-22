-- name: CreateProject :exec
INSERT INTO projects (id, title, author, source_lang, target_lang, original_format, total_chapters)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetProject :one
SELECT * FROM projects WHERE id = ? LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects ORDER BY updated_at DESC;

-- name: UpdateProjectProgress :exec
UPDATE projects 
SET total_chapters = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;

-- name: UpdateProjectStyle :exec
UPDATE projects
SET style_name = ?, style_guide = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

