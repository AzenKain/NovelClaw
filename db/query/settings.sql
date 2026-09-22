-- name: UpsertProjectSettings :exec
INSERT INTO project_settings (project_id, style_guide_json, model_routing_json, soul_id)
VALUES (?, ?, ?, ?)
ON CONFLICT(project_id) DO UPDATE SET
    style_guide_json = excluded.style_guide_json,
    model_routing_json = excluded.model_routing_json,
    soul_id = excluded.soul_id,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetProjectSettings :one
SELECT * FROM project_settings
WHERE project_id = ?
LIMIT 1;

-- name: DeleteProjectSettings :exec
DELETE FROM project_settings WHERE project_id = ?;

-- name: UpsertProjectSkillsConfig :exec
INSERT INTO project_settings (project_id, skills_config_json)
VALUES (?, ?)
ON CONFLICT(project_id) DO UPDATE SET
    skills_config_json = excluded.skills_config_json,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetProjectSkillsConfig :one
SELECT skills_config_json FROM project_settings
WHERE project_id = ?
LIMIT 1;

