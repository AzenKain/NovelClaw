package storage

import (
	"context"
	"fmt"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/jsonx"
)

// UpsertChapterTimeline inserts or updates summary and event milestones for a chapter (L1 Episodic Memory).
func (s *Storage) UpsertChapterTimeline(ctx context.Context, projectID string, chapterIndex int64, summaryText string, milestones any) error {
	var milestonesJSON string = "[]"
	if milestones != nil {
		b, err := jsonx.Marshal(milestones)
		if err != nil {
			return fmt.Errorf("marshal milestones: %w", err)
		}
		milestonesJSON = string(b)
	}

	id := fmt.Sprintf("tl_%s_%d", projectID, chapterIndex)
	return s.q.UpsertChapterTimeline(ctx, sqlc.UpsertChapterTimelineParams{
		ID:             id,
		ProjectID:      projectID,
		ChapterIndex:   chapterIndex,
		SummaryText:    summaryText,
		MilestonesJson: milestonesJSON,
	})
}

// GetChapterTimeline retrieves timeline summary for a specific chapter.
func (s *Storage) GetChapterTimeline(ctx context.Context, projectID string, chapterIndex int64) (sqlc.ChapterTimeline, error) {
	return s.q.GetChapterTimeline(ctx, sqlc.GetChapterTimelineParams{
		ProjectID:    projectID,
		ChapterIndex: chapterIndex,
	})
}

// ListTimelineByProject retrieves all chapter summaries for L1 context memory window.
func (s *Storage) ListTimelineByProject(ctx context.Context, projectID string) ([]sqlc.ChapterTimeline, error) {
	return s.q.ListTimelineByProject(ctx, projectID)
}

// DeleteTimelineByProject removes all timeline entries for a project.
func (s *Storage) DeleteTimelineByProject(ctx context.Context, projectID string) error {
	return s.q.DeleteTimelineByProject(ctx, projectID)
}
