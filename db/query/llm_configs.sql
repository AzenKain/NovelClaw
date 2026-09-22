-- name: UpsertLLMConfig :exec
INSERT INTO llm_configs (
    id, provider_name, api_url, encrypted_token, model_name,
    is_active, is_default, rate_limit_rpm, custom_headers_json,
    max_retries, timeout_seconds, reasoning_effort
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    provider_name = excluded.provider_name,
    api_url = excluded.api_url,
    encrypted_token = excluded.encrypted_token,
    model_name = excluded.model_name,
    is_active = excluded.is_active,
    is_default = excluded.is_default,
    rate_limit_rpm = excluded.rate_limit_rpm,
    custom_headers_json = excluded.custom_headers_json,
    max_retries = excluded.max_retries,
    timeout_seconds = excluded.timeout_seconds,
    reasoning_effort = excluded.reasoning_effort,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetDefaultLLMConfig :one
SELECT * FROM llm_configs
WHERE is_default = 1 AND is_active = 1
LIMIT 1;

-- name: GetLLMConfigByID :one
SELECT * FROM llm_configs
WHERE id = ?
LIMIT 1;

-- name: ListActiveLLMConfigs :many
SELECT * FROM llm_configs
WHERE is_active = 1
ORDER BY is_default DESC, updated_at DESC;

-- name: DeleteLLMConfig :exec
DELETE FROM llm_configs WHERE id = ?;
