package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"novelclaw/internal/gen/sqlc"
)

// GetOrCreateMainThread ensures a main NovelClaw thread exists for the given project.
func (s *Storage) GetOrCreateMainThread(ctx context.Context, projectID string) (sqlc.NovelclawThread, error) {
	thread, err := s.q.GetMainNovelClawThread(ctx, projectID)
	if err == nil {
		return thread, nil
	}
	if err != sql.ErrNoRows {
		return sqlc.NovelclawThread{}, fmt.Errorf("get main thread: %w", err)
	}

	// Create main thread
	newID := "nc_thread_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	return s.q.CreateNovelClawThread(ctx, sqlc.CreateNovelClawThreadParams{
		ID:           newID,
		ProjectID:    projectID,
		Title:        "Main Discussion Thread",
		IsMain:       1,
		VolumeIndex:  0,
		ChapterIndex: 0,
	})
}

// CreateSubThread creates a new specialized sub-chat thread for a project, volume, or chapter.
func (s *Storage) CreateSubThread(ctx context.Context, projectID, title string, volIndex, chapIndex int64) (sqlc.NovelclawThread, error) {
	newID := "nc_thread_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	if strings.TrimSpace(title) == "" {
		title = "New Sub-chat"
	}
	return s.q.CreateNovelClawThread(ctx, sqlc.CreateNovelClawThreadParams{
		ID:           newID,
		ProjectID:    projectID,
		Title:        title,
		IsMain:       0,
		VolumeIndex:  volIndex,
		ChapterIndex: chapIndex,
	})
}

// GetNovelClawThread returns a thread by its unique ID.
func (s *Storage) GetNovelClawThread(ctx context.Context, threadID string) (sqlc.NovelclawThread, error) {
	return s.q.GetNovelClawThread(ctx, threadID)
}

// ListNovelClawThreads returns all threads for a project ordered with main thread first.
func (s *Storage) ListNovelClawThreads(ctx context.Context, projectID string) ([]sqlc.NovelclawThread, error) {
	return s.q.ListNovelClawThreadsByProject(ctx, projectID)
}

// UpdateNovelClawThreadTitle updates the title of a thread.
func (s *Storage) UpdateNovelClawThreadTitle(ctx context.Context, threadID, title string) error {
	return s.q.UpdateNovelClawThreadTitle(ctx, sqlc.UpdateNovelClawThreadTitleParams{
		Title: title,
		ID:    threadID,
	})
}

// DeleteNovelClawThread deletes a thread and all associated messages.
func (s *Storage) DeleteNovelClawThread(ctx context.Context, threadID string) error {
	return s.q.DeleteNovelClawThread(ctx, threadID)
}

// SaveNovelClawMessage inserts a new message into a thread and updates the thread's updated_at timestamp.
func (s *Storage) SaveNovelClawMessage(ctx context.Context, params sqlc.CreateNovelClawMessageParams) (sqlc.NovelclawMessage, error) {
	if params.ID == "" {
		params.ID = "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	}
	msg, err := s.q.CreateNovelClawMessage(ctx, params)
	if err != nil {
		return sqlc.NovelclawMessage{}, err
	}
	_ = s.q.TouchNovelClawThread(ctx, params.ThreadID)
	return msg, nil
}

// ListNovelClawMessages returns all messages in chronological order for a thread.
func (s *Storage) ListNovelClawMessages(ctx context.Context, threadID string) ([]sqlc.NovelclawMessage, error) {
	return s.q.ListNovelClawMessagesByThread(ctx, threadID)
}

// ListActiveNovelClawMessages returns non-compacted messages in chronological order for a thread.
func (s *Storage) ListActiveNovelClawMessages(ctx context.Context, threadID string) ([]sqlc.NovelclawMessage, error) {
	return s.q.ListActiveNovelClawMessagesByThread(ctx, threadID)
}

// SumActiveThreadTokens calculates the total token count of uncompacted messages in a thread.
func (s *Storage) SumActiveThreadTokens(ctx context.Context, threadID string) (int64, error) {
	tokens, err := s.q.SumActiveThreadTokens(ctx, threadID)
	if err != nil {
		return 0, err
	}
	return int64(tokens), nil
}

// ArchiveThreadMessagesForCompact marks all uncompacted messages in a thread as compacted.
func (s *Storage) ArchiveThreadMessagesForCompact(ctx context.Context, threadID string) error {
	return s.q.ArchiveThreadMessagesForCompact(ctx, threadID)
}

// ClearNovelClawThreadMessages deletes all messages in a thread.
func (s *Storage) ClearNovelClawThreadMessages(ctx context.Context, threadID string) error {
	return s.q.DeleteNovelClawMessagesByThread(ctx, threadID)
}
