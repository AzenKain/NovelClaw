package storage

import (
	"context"
	"novelclaw/internal/gen/sqlc"
)

// UpsertRelation inserts or updates an address term relationship between two characters.
func (s *Storage) UpsertRelation(ctx context.Context, arg sqlc.UpsertRelationParams) error {
	return s.q.UpsertRelation(ctx, arg)
}

// GetRelationBetween retrieves effective address term relationship at a specific chapter.
func (s *Storage) GetRelationBetween(ctx context.Context, projectID, fromChar, toChar string, sinceChapter int64) (sqlc.CharacterRelation, error) {
	return s.q.GetRelationBetween(ctx, sqlc.GetRelationBetweenParams{
		ProjectID:    projectID,
		FromChar:     fromChar,
		ToChar:       toChar,
		SinceChapter: sinceChapter,
	})
}

// ListRelationsByProject lists all character relationship mappings for a project.
func (s *Storage) ListRelationsByProject(ctx context.Context, projectID string) ([]sqlc.CharacterRelation, error) {
	return s.q.ListRelationsByProject(ctx, projectID)
}

// LockRelation locks a relationship to prevent automatic AI modifications.
func (s *Storage) LockRelation(ctx context.Context, id string) error {
	return s.q.LockRelation(ctx, id)
}

// DeleteRelation deletes a character relationship by id.
func (s *Storage) DeleteRelation(ctx context.Context, id string) error {
	return s.q.DeleteRelation(ctx, id)
}

