-- name: SaveCheckpoint :exec
INSERT INTO checkpoints (id, project_id, chapter_index, step_name, state_json)
VALUES (?, ?, ?, ?, ?);

-- name: GetLatestCheckpoint :one
SELECT * FROM checkpoints 
WHERE project_id = ? 
ORDER BY rowid DESC 
LIMIT 1;

-- name: GetLatestChapterCheckpoint :one
SELECT * FROM checkpoints 
WHERE project_id = ? AND chapter_index = ? 
ORDER BY rowid DESC 
LIMIT 1;

-- name: ClearCheckpointsByProject :exec
DELETE FROM checkpoints WHERE project_id = ?;
