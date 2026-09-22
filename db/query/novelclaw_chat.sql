-- name: CreateNovelClawThread :one
INSERT INTO novelclaw_threads (
    id, project_id, title, is_main, volume_index, chapter_index, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: GetNovelClawThread :one
SELECT * FROM novelclaw_threads
WHERE id = ? LIMIT 1;

-- name: GetMainNovelClawThread :one
SELECT * FROM novelclaw_threads
WHERE project_id = ? AND is_main = 1
LIMIT 1;

-- name: ListNovelClawThreadsByProject :many
SELECT * FROM novelclaw_threads
WHERE project_id = ?
ORDER BY is_main DESC, updated_at DESC;

-- name: UpdateNovelClawThreadTitle :exec
UPDATE novelclaw_threads
SET title = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: TouchNovelClawThread :exec
UPDATE novelclaw_threads
SET updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteNovelClawThread :exec
DELETE FROM novelclaw_threads
WHERE id = ?;

-- name: CreateNovelClawMessage :one
INSERT INTO novelclaw_messages (
    id, thread_id, project_id, sender, role, content,
    thinking_content, step_type, step_status, action_call_json,
    is_collapsed, is_archived_compact, token_count, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: ListNovelClawMessagesByThread :many
SELECT * FROM novelclaw_messages
WHERE thread_id = ?
ORDER BY created_at ASC;

-- name: ListActiveNovelClawMessagesByThread :many
SELECT * FROM novelclaw_messages
WHERE thread_id = ? AND is_archived_compact = 0
ORDER BY created_at ASC;

-- name: SumActiveThreadTokens :one
SELECT CAST(COALESCE(SUM(token_count), 0) AS INTEGER) AS total_tokens
FROM novelclaw_messages
WHERE thread_id = ? AND is_archived_compact = 0;

-- name: ArchiveThreadMessagesForCompact :exec
UPDATE novelclaw_messages
SET is_archived_compact = 1
WHERE thread_id = ? AND is_archived_compact = 0;

-- name: DeleteNovelClawMessagesByThread :exec
DELETE FROM novelclaw_messages
WHERE thread_id = ?;
