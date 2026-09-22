-- name: IndexChapterFTS :exec
INSERT INTO chapters_fts (chapter_id, project_id, chapter_index, title, raw_content, translated_content)
VALUES (?, ?, ?, ?, ?, ?);

-- name: SearchBookContextFTS :many
SELECT chapter_id, project_id, chapter_index, title, raw_content, translated_content
FROM chapters_fts
WHERE raw_content MATCH sqlc.arg('query') 
  AND project_id = sqlc.arg('project_id') 
  AND (CAST(sqlc.arg('max_chapter_index') AS INTEGER) <= 0 OR CAST(chapter_index AS INTEGER) <= CAST(sqlc.arg('max_chapter_index') AS INTEGER))
ORDER BY rank
LIMIT sqlc.arg('limit');

-- name: SearchSnippetFTS :many
SELECT chapter_id, project_id, chapter_index, title,
       snippet(chapters_fts, 4, '[MARK]', '[/MARK]', ' ... ', 15) AS snippet
FROM chapters_fts
WHERE raw_content MATCH sqlc.arg('query') 
  AND project_id = sqlc.arg('project_id') 
  AND (CAST(sqlc.arg('max_chapter_index') AS INTEGER) <= 0 OR CAST(chapter_index AS INTEGER) <= CAST(sqlc.arg('max_chapter_index') AS INTEGER))
ORDER BY rank
LIMIT sqlc.arg('limit');

-- name: UpdateChapterTranslationFTS :exec
UPDATE chapters_fts 
SET translated_content = ? 
WHERE chapter_id = ?;

-- name: SearchBookContextLike :many
SELECT id AS chapter_id, project_id, CAST(chapter_index AS TEXT) AS chapter_index, title, raw_content, translated_content
FROM chapters
WHERE project_id = sqlc.arg('project_id') 
  AND (CAST(sqlc.arg('max_chapter_index') AS INTEGER) <= 0 OR chapter_index <= CAST(sqlc.arg('max_chapter_index') AS INTEGER)) 
  AND (raw_content LIKE sqlc.arg('pattern') OR translated_content LIKE sqlc.arg('pattern'))
ORDER BY chapter_index ASC
LIMIT sqlc.arg('limit');

-- name: DeleteChapterFTS :exec
DELETE FROM chapters_fts WHERE chapter_id = ?;

-- name: ClearProjectFTS :exec
DELETE FROM chapters_fts WHERE project_id = ?;
