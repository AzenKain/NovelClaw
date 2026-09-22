package storage

import (
	"context"
	"fmt"
	"novelclaw/internal/gen/sqlc"
	"strconv"
)

// CreateChapter creates a new chapter.
func (s *Storage) CreateChapter(ctx context.Context, c sqlc.CreateChapterParams) error {
	return s.q.CreateChapter(ctx, c)
}

// GetChapter fetches a chapter by ID.
func (s *Storage) GetChapter(ctx context.Context, id string) (sqlc.Chapter, error) {
	return s.q.GetChapter(ctx, id)
}

// GetChapterByIndex fetches a chapter by ProjectID and its ChapterIndex.
func (s *Storage) GetChapterByIndex(ctx context.Context, projectID string, chapterIndex int64) (sqlc.Chapter, error) {
	return s.q.GetChapterByIndex(ctx, sqlc.GetChapterByIndexParams{
		ProjectID:    projectID,
		ChapterIndex: chapterIndex,
	})
}

// ListChaptersByProject returns all chapters of a project in ascending index order.
func (s *Storage) ListChaptersByProject(ctx context.Context, projectID string) ([]sqlc.Chapter, error) {
	return s.q.ListChaptersByProject(ctx, projectID)
}

// UpdateChapterTranslation updates a chapter's translated content and status, and syncs the FTS5 index within the same transaction.
func (s *Storage) UpdateChapterTranslation(ctx context.Context, id string, translatedContent string, status string) error {
	return s.WithTx(ctx, func(q *sqlc.Queries) error {
		err := q.UpdateChapterTranslation(ctx, sqlc.UpdateChapterTranslationParams{
			ID:                id,
			TranslatedContent: translatedContent,
			Status:            status,
		})
		if err != nil {
			return fmt.Errorf("update chapter: %w", err)
		}

		err = q.UpdateChapterTranslationFTS(ctx, sqlc.UpdateChapterTranslationFTSParams{
			TranslatedContent: translatedContent,
			ChapterID:         id,
		})
		if err != nil {
			return fmt.Errorf("update chapter fts: %w", err)
		}
		return nil
	})
}

// CountChaptersByStatus counts chapters per status (pending, translating, completed...).
func (s *Storage) CountChaptersByStatus(ctx context.Context, projectID string) ([]sqlc.CountChaptersByStatusRow, error) {
	return s.q.CountChaptersByStatus(ctx, projectID)
}

// SaveChapterAndIndex saves a chapter into the chapters table and indexes it into chapters_fts within a single transaction.
func (s *Storage) SaveChapterAndIndex(ctx context.Context, c sqlc.CreateChapterParams) error {
	return s.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.CreateChapter(ctx, c); err != nil {
			return fmt.Errorf("create chapter: %w", err)
		}

		idxParam := sqlc.IndexChapterFTSParams{
			ChapterID:         c.ID,
			ProjectID:         c.ProjectID,
			ChapterIndex:      strconv.FormatInt(c.ChapterIndex, 10),
			Title:             c.Title,
			RawContent:        c.RawContent,
			TranslatedContent: "",
		}
		if err := q.IndexChapterFTS(ctx, idxParam); err != nil {
			return fmt.Errorf("index chapter fts: %w", err)
		}
		return nil
	})
}
