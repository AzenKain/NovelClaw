package storage

import (
	"context"
	"fmt"
	"novelclaw/internal/gen/sqlc"
)

// CreateProject creates a new book project.
func (s *Storage) CreateProject(ctx context.Context, p sqlc.CreateProjectParams) error {
	return s.q.CreateProject(ctx, p)
}

// GetProject retrieves project details by ID.
func (s *Storage) GetProject(ctx context.Context, id string) (sqlc.Project, error) {
	return s.q.GetProject(ctx, id)
}

// ListProjects lists all projects ordered by most recently updated.
func (s *Storage) ListProjects(ctx context.Context) ([]sqlc.Project, error) {
	return s.q.ListProjects(ctx)
}

// DeleteProject deletes a project and cascades related data including FTS5.
func (s *Storage) DeleteProject(ctx context.Context, id string) error {
	return s.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.ClearProjectFTS(ctx, id); err != nil {
			return err
		}
		return q.DeleteProject(ctx, id)
	})
}

// UpdateProjectProgress updates the total chapters count for a project.
func (s *Storage) UpdateProjectProgress(ctx context.Context, id string, totalChapters int64) error {
	return s.q.UpdateProjectProgress(ctx, sqlc.UpdateProjectProgressParams{
		ID:            id,
		TotalChapters: totalChapters,
	})
}

// UpdateProjectLanguages updates source and target languages for a project.
func (s *Storage) UpdateProjectLanguages(ctx context.Context, id string, sourceLang, targetLang string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE projects
		SET source_lang = ?, target_lang = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, sourceLang, targetLang, id)
	return err
}

// UpdateProjectTitle updates the title of a project.
func (s *Storage) UpdateProjectTitle(ctx context.Context, id string, title string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE projects
		SET title = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, id)
	return err
}

// UpdateProjectStyle updates the style name and style guide of a project.
func (s *Storage) UpdateProjectStyle(ctx context.Context, id string, styleName string, styleGuide string) error {
	return s.q.UpdateProjectStyle(ctx, sqlc.UpdateProjectStyleParams{
		ID:         id,
		StyleName:  styleName,
		StyleGuide: styleGuide,
	})
}


// InheritProjectKnowledge copies entities, relations, and glossary from sourceProjectID into targetProjectID.
func (s *Storage) InheritProjectKnowledge(ctx context.Context, sourceProjectID, targetProjectID string) (int, int, int, error) {
	if sourceProjectID == targetProjectID {
		return 0, 0, 0, fmt.Errorf("source and target project must be different")
	}

	// 1. Entities
	entities, err := s.ListEntitiesByProject(ctx, sourceProjectID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("list source entities: %w", err)
	}
	entCount := 0
	for _, ent := range entities {
		entCopy := ent
		entCopy.ID = fmt.Sprintf("ent_%s_%s", targetProjectID, ent.Name)
		entCopy.ProjectID = targetProjectID
		if err := s.UpsertEntity(ctx, entCopy); err == nil {
			entCount++
		}
	}

	// 2. Relations
	relations, err := s.ListRelationsByProject(ctx, sourceProjectID)
	if err != nil {
		return entCount, 0, 0, fmt.Errorf("list source relations: %w", err)
	}
	relCount := 0
	for _, rel := range relations {
		relParams := sqlc.UpsertRelationParams{
			ID:           fmt.Sprintf("rel_%s_%s_%s", targetProjectID, rel.FromChar, rel.ToChar),
			ProjectID:    targetProjectID,
			FromChar:     rel.FromChar,
			ToChar:       rel.ToChar,
			CallAs:       rel.CallAs,
			SelfCallAs:   rel.SelfCallAs,
			SinceChapter: rel.SinceChapter,
			Tone:         rel.Tone,
			IsLocked:     rel.IsLocked,
		}
		if err := s.UpsertRelation(ctx, relParams); err == nil {
			relCount++
		}
	}

	// 3. Glossary
	glossary, err := s.ListGlossaryByProject(ctx, sourceProjectID)
	if err != nil {
		return entCount, relCount, 0, fmt.Errorf("list source glossary: %w", err)
	}
	glossCount := 0
	for _, g := range glossary {
		gParams := sqlc.UpsertGlossaryTermParams{
			ID:         fmt.Sprintf("glos_%s_%s", targetProjectID, g.SourceTerm),
			ProjectID:  targetProjectID,
			SourceTerm: g.SourceTerm,
			TargetTerm: g.TargetTerm,
			Category:   g.Category,
		}
		if err := s.UpsertGlossaryTerm(ctx, gParams); err == nil {
			glossCount++
		}
	}

	return entCount, relCount, glossCount, nil
}

