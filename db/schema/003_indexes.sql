-- Migration 003: Bổ sung chỉ mục tối ưu hóa hiệu năng truy vấn và ràng buộc dữ liệu

-- 1. Chỉ mục tra cứu chương theo dự án và số thứ tự
CREATE INDEX IF NOT EXISTS idx_chapters_project_idx ON chapters(project_id, chapter_index);
CREATE INDEX IF NOT EXISTS idx_chapters_status ON chapters(project_id, status);

-- 2. Chỉ mục tra cứu ma trận quan hệ xưng hô L2 (from_char, to_char, since_chapter)
CREATE INDEX IF NOT EXISTS idx_relations_lookup ON character_relations(project_id, from_char, to_char, since_chapter);

-- 3. Chỉ mục tra cứu Checkpoints phục hồi từng bước
CREATE INDEX IF NOT EXISTS idx_checkpoints_lookup ON checkpoints(project_id, chapter_index, id);

-- 4. Ràng buộc và chỉ mục từ điển thuật ngữ L3
CREATE UNIQUE INDEX IF NOT EXISTS idx_glossary_unique ON glossary(project_id, source_term);
CREATE INDEX IF NOT EXISTS idx_glossary_category ON glossary(project_id, category);

-- 5. Chỉ mục dòng thời gian L1
CREATE INDEX IF NOT EXISTS idx_timeline_project_chapter ON chapter_timeline(project_id, chapter_index);
