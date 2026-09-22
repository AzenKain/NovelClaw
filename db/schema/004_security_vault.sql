-- Migration 004: Bảng Quản lý Cấu hình Nhà cung cấp LLM & Két sắt API Key Mã hóa (Security Vault)
CREATE TABLE IF NOT EXISTS llm_configs (
    id TEXT PRIMARY KEY,
    provider_name TEXT NOT NULL,
    api_url TEXT NOT NULL,
    encrypted_token TEXT NOT NULL,
    model_name TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    is_default INTEGER NOT NULL DEFAULT 0,
    rate_limit_rpm INTEGER NOT NULL DEFAULT 60,
    custom_headers_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_llm_configs_active ON llm_configs(is_active, is_default);
