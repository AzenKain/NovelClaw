package storage

import (
	"context"
	"novelclaw/internal/gen/sqlc"
)

// UpsertGlossaryTerm inserts or updates a terminology item in the project glossary.
func (s *Storage) UpsertGlossaryTerm(ctx context.Context, arg sqlc.UpsertGlossaryTermParams) error {
	return s.q.UpsertGlossaryTerm(ctx, arg)
}

// ListGlossaryByProject retrieves all glossary terms for a project.
func (s *Storage) ListGlossaryByProject(ctx context.Context, projectID string) ([]sqlc.Glossary, error) {
	return s.q.ListGlossaryByProject(ctx, projectID)
}

// DeleteGlossaryTerm deletes a glossary term by ID.
func (s *Storage) DeleteGlossaryTerm(ctx context.Context, id string) error {
	return s.q.DeleteGlossaryTerm(ctx, id)
}
