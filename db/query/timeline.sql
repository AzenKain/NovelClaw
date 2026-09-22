-- name: UpsertChapterTimeline :exec
INSERT INTO chapter_timeline (id, project_id, chapter_index, summary_text, milestones_json)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(project_id, chapter_index) DO UPDATE SET
    summary_text = excluded.summary_text,
    milestones_json = excluded.milestones_json;

-- name: GetChapterTimeline :one
SELECT * FROM chapter_timeline
WHERE project_id = ? AND chapter_index = ?
LIMIT 1;

-- name: ListTimelineByProject :many
SELECT * FROM chapter_timeline
WHERE project_id = ?
ORDER BY chapter_index ASC;

-- name: DeleteTimelineByProject :exec
DELETE FROM chapter_timeline WHERE project_id = ?;
