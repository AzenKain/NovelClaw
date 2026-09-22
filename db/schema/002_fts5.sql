-- Bảng ảo FTS5 phục vụ Agentic RAG toàn văn
CREATE VIRTUAL TABLE IF NOT EXISTS chapters_fts USING fts5(
    chapter_id UNINDEXED,
    project_id UNINDEXED,
    chapter_index UNINDEXED,
    title,
    raw_content,
    translated_content,
    tokenize = 'trigram'
);
