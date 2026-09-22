package storage

import (
	"context"
	"fmt"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/jsonx"
	"time"
)

// SaveCheckpoint saves a step-level checkpoint with serialized JSON state.
func (s *Storage) SaveCheckpoint(ctx context.Context, projectID string, chapterIndex int64, stepName string, state any) error {
	jsonBytes, err := jsonx.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state json: %w", err)
	}

	checkpointID := fmt.Sprintf("cp_%s_%d_%s_%d", projectID, chapterIndex, stepName, time.Now().UnixNano())
	return s.q.SaveCheckpoint(ctx, sqlc.SaveCheckpointParams{
		ID:           checkpointID,
		ProjectID:    projectID,
		ChapterIndex: chapterIndex,
		StepName:     stepName,
		StateJson:    string(jsonBytes),
	})
}

// GetLatestCheckpoint retrieves the latest checkpoint across an entire project.
func (s *Storage) GetLatestCheckpoint(ctx context.Context, projectID string) (sqlc.Checkpoint, error) {
	return s.q.GetLatestCheckpoint(ctx, projectID)
}

// GetLatestChapterCheckpoint retrieves the latest checkpoint for a specific chapter.
func (s *Storage) GetLatestChapterCheckpoint(ctx context.Context, projectID string, chapterIndex int64) (sqlc.Checkpoint, error) {
	return s.q.GetLatestChapterCheckpoint(ctx, sqlc.GetLatestChapterCheckpointParams{
		ProjectID:    projectID,
		ChapterIndex: chapterIndex,
	})
}

// ClearCheckpointsByProject clears all checkpoints for a project.
func (s *Storage) ClearCheckpointsByProject(ctx context.Context, projectID string) error {
	return s.q.ClearCheckpointsByProject(ctx, projectID)
}

// UnpackCheckpointState deserializes checkpoint state JSON into the target struct.
func UnpackCheckpointState[T any](cp sqlc.Checkpoint, target *T) error {
	return jsonx.Unmarshal([]byte(cp.StateJson), target)
}

// RollbackChapterToCheckpoint rolls back chapter translation text and status to the latest checkpoint.
func (s *Storage) RollbackChapterToCheckpoint(ctx context.Context, projectID string, chapterIndex int64) (*sqlc.Chapter, error) {
	cp, err := s.GetLatestChapterCheckpoint(ctx, projectID, chapterIndex)
	if err != nil {
		return nil, fmt.Errorf("no checkpoint found for chapter %d: %w", chapterIndex, err)
	}

	var state map[string]string
	if err := UnpackCheckpointState(cp, &state); err != nil {
		return nil, fmt.Errorf("unpack checkpoint state: %w", err)
	}

	chID := state["chapter_id"]
	prevContent := state["translated_content"]
	prevStatus := state["status"]
	if prevStatus == "" {
		prevStatus = "pending"
	}

	if chID == "" {
		ch, err := s.GetChapterByIndex(ctx, projectID, chapterIndex)
		if err != nil {
			return nil, fmt.Errorf("get chapter: %w", err)
		}
		chID = ch.ID
	}

	if err := s.UpdateChapterTranslation(ctx, chID, prevContent, prevStatus); err != nil {
		return nil, fmt.Errorf("restore chapter translation: %w", err)
	}

	restoredCh, err := s.GetChapter(ctx, chID)
	if err != nil {
		return nil, fmt.Errorf("get restored chapter: %w", err)
	}

	return &restoredCh, nil
}
