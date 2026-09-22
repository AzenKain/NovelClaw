-- Bảng Dự Án (Projects)
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT '',
    source_lang TEXT NOT NULL,
    target_lang TEXT NOT NULL,
    original_format TEXT NOT NULL DEFAULT '',
    total_chapters INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Các Chương của Sách (Chapters)
CREATE TABLE IF NOT EXISTS chapters (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    chapter_index INTEGER NOT NULL,
    title TEXT NOT NULL,
    content_path TEXT NOT NULL DEFAULT '',
    raw_content TEXT NOT NULL,
    translated_content TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending', -- pending, translating, audited, completed, failed
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, chapter_index)
);

-- Bảng Sơ đồ Quan hệ Nhân vật & Ngôi xưng (Character Relations - L2)
CREATE TABLE IF NOT EXISTS character_relations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    from_char TEXT NOT NULL,
    to_char TEXT NOT NULL,
    call_as TEXT NOT NULL,          -- Nhân vật A gọi B là gì
    self_call_as TEXT NOT NULL,     -- Nhân vật A tự xưng với B là gì
    since_chapter INTEGER NOT NULL DEFAULT 1,
    tone TEXT NOT NULL DEFAULT '',  -- Sắc thái: thân mật, thù địch, tôn kính
    is_locked INTEGER NOT NULL DEFAULT 0, -- 1: Đã được người dùng khóa cứng
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Từ điển Thuật ngữ Dự án (Glossary - L3)
CREATE TABLE IF NOT EXISTS glossary (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_term TEXT NOT NULL,
    target_term TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general', -- proper_name, skill, realm, item, location
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Nhật ký Dòng thời gian & Tóm tắt Chương (Episodic Timeline - L1)
CREATE TABLE IF NOT EXISTS chapter_timeline (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    chapter_index INTEGER NOT NULL,
    summary_text TEXT NOT NULL,
    milestones_json TEXT NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, chapter_index)
);

-- Bảng Checkpoints Phục hồi Từng bước (Học hỏi từ ainovel-cli)
CREATE TABLE IF NOT EXISTS checkpoints (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    chapter_index INTEGER NOT NULL,
    step_name TEXT NOT NULL,
    state_json TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Cấu hình Dự án (Style Guide, Model Routing, Soul)
CREATE TABLE IF NOT EXISTS project_settings (
    project_id TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    style_guide_json TEXT NOT NULL DEFAULT '{}',
    model_routing_json TEXT NOT NULL DEFAULT '{}',
    soul_id TEXT NOT NULL DEFAULT 'default_neko',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
