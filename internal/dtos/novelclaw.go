package dtos

import (
	"database/sql"
	"time"

	"novelclaw/internal/gen/sqlc"
)

// NovelClawThreadDTO represents a conversation thread in NovelClaw.
type NovelClawThreadDTO struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Title        string    `json:"title"`
	IsMain       bool      `json:"is_main"`
	VolumeIndex  int64     `json:"volume_index"`
	ChapterIndex int64     `json:"chapter_index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TokenCount   int64     `json:"token_count,omitempty"`
}

// NovelClawMessageDTO represents a message or activity event within a thread.
type NovelClawMessageDTO struct {
	ID                string    `json:"id"`
	ThreadID          string    `json:"thread_id"`
	ProjectID         string    `json:"project_id"`
	Sender            string    `json:"sender"` // "user" | "novelclaw" | "system" | "agent_stream"
	Role              string    `json:"role"`   // "user" | "assistant" | "system" | "tool"
	Content           string    `json:"content"`
	ThinkingContent   string    `json:"thinking_content,omitempty"`
	StepType          string    `json:"step_type,omitempty"`   // e.g. "thinking", "tool_call", "action_dispatched"
	StepStatus        string    `json:"step_status,omitempty"` // "running" | "completed" | "error"
	ActionCallJSON    string    `json:"action_call_json,omitempty"`
	IsCollapsed       bool      `json:"is_collapsed"`
	IsArchivedCompact bool      `json:"is_archived_compact"`
	TokenCount        int64     `json:"token_count"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreateSubThreadRequest represents parameters to create a new sub-chat thread.
type CreateSubThreadRequest struct {
	ProjectID    string `json:"project_id"`
	Title        string `json:"title"`
	VolumeIndex  int64  `json:"volume_index"`
	ChapterIndex int64  `json:"chapter_index"`
}

// SendNovelClawMessageRequest represents parameters to send a message to NovelClaw.
type SendNovelClawMessageRequest struct {
	ThreadID  string `json:"thread_id"`
	ProjectID string `json:"project_id"`
	Content   string `json:"content"`
}

// AppActionPayload represents an action dispatched by NovelClaw Omni-App tools to frontend/backend.
type AppActionPayload struct {
	Action      string         `json:"action"`
	Category    string         `json:"category"` // "navigation", "graph", "world_bible", "glossary", "pipeline", "execution", "export"
	Description string         `json:"description"`
	Data        map[string]any `json:"data"`
}

// CompactThreadRequest represents parameters to trigger a thread auto-compact.
type CompactThreadRequest struct {
	ThreadID string `json:"thread_id"`
}

// CompactThreadResponse contains results after compacting a thread.
type CompactThreadResponse struct {
	ThreadID         string `json:"thread_id"`
	OriginalTokens   int64  `json:"original_tokens"`
	CompactedTokens  int64  `json:"compacted_tokens"`
	ArchivedMessages int    `json:"archived_messages"`
	SummaryNotice    string `json:"summary_notice"`
}

// ToNovelClawThreadDTO converts a sqlc.NovelclawThread to NovelClawThreadDTO.
func ToNovelClawThreadDTO(t sqlc.NovelclawThread) NovelClawThreadDTO {
	return NovelClawThreadDTO{
		ID:           t.ID,
		ProjectID:    t.ProjectID,
		Title:        t.Title,
		IsMain:       t.IsMain != 0,
		VolumeIndex:  t.VolumeIndex,
		ChapterIndex: t.ChapterIndex,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

// ToNovelClawMessageDTO converts a sqlc.NovelclawMessage to NovelClawMessageDTO.
func ToNovelClawMessageDTO(m sqlc.NovelclawMessage) NovelClawMessageDTO {
	return NovelClawMessageDTO{
		ID:                m.ID,
		ThreadID:          m.ThreadID,
		ProjectID:         m.ProjectID,
		Sender:            m.Sender,
		Role:              m.Role,
		Content:           m.Content,
		ThinkingContent:   m.ThinkingContent.String,
		StepType:          m.StepType.String,
		StepStatus:        m.StepStatus.String,
		ActionCallJSON:    m.ActionCallJson.String,
		IsCollapsed:       m.IsCollapsed != 0,
		IsArchivedCompact: m.IsArchivedCompact != 0,
		TokenCount:        m.TokenCount,
		CreatedAt:         m.CreatedAt,
	}
}

// NullStringHelper converts a string to sql.NullString.
func NullStringHelper(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
