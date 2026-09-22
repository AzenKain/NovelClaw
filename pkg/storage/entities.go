package storage

import (
	"context"
	"fmt"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/jsonx"
)

// Entity represents an entity (character, faction, location, item) in the progressive graph.
type Entity struct {
	ID               string         `json:"id"`
	ProjectID        string         `json:"project_id"`
	Name             string         `json:"name"`
	Aliases          []string       `json:"aliases"`
	Category         string         `json:"category"`
	Gender           string         `json:"gender"`
	Role             string         `json:"role"`
	FirstSeenChapter int64          `json:"first_seen_chapter"`
	Metadata         map[string]any `json:"metadata"`
}

// UpsertEntity inserts or updates an entity node in the project graph.
func (s *Storage) UpsertEntity(ctx context.Context, e Entity) error {
	if e.ID == "" {
		e.ID = fmt.Sprintf("ent_%s_%s", e.ProjectID, e.Name)
	}
	if e.Category == "" {
		e.Category = "character"
	}
	if e.Gender == "" {
		e.Gender = "unknown"
	}
	if e.FirstSeenChapter <= 0 {
		e.FirstSeenChapter = 1
	}

	aliasesJSON := "[]"
	if len(e.Aliases) > 0 {
		b, err := jsonx.Marshal(e.Aliases)
		if err != nil {
			return fmt.Errorf("marshal aliases: %w", err)
		}
		aliasesJSON = string(b)
	}

	metadataJSON := "{}"
	if e.Metadata != nil {
		b, err := jsonx.Marshal(e.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		metadataJSON = string(b)
	}

	return s.q.UpsertEntity(ctx, sqlc.UpsertEntityParams{
		ID:               e.ID,
		ProjectID:        e.ProjectID,
		Name:             e.Name,
		AliasesJson:      aliasesJSON,
		Category:         e.Category,
		Gender:           e.Gender,
		Role:             e.Role,
		FirstSeenChapter: e.FirstSeenChapter,
		MetadataJson:     metadataJSON,
	})
}

// GetEntityByName retrieves an entity by project and exact name.
func (s *Storage) GetEntityByName(ctx context.Context, projectID, name string) (*Entity, error) {
	row, err := s.q.GetEntityByName(ctx, sqlc.GetEntityByNameParams{
		ProjectID: projectID,
		Name:      name,
	})
	if err != nil {
		return nil, err
	}
	return toDomainEntity(row), nil
}

// ListEntitiesByProject returns all entities defined for a project.
func (s *Storage) ListEntitiesByProject(ctx context.Context, projectID string) ([]Entity, error) {
	rows, err := s.q.ListEntitiesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	out := make([]Entity, 0, len(rows))
	for _, r := range rows {
		out = append(out, *toDomainEntity(r))
	}
	return out, nil
}

// ListActiveEntitiesAtChapter returns entities that have been introduced at or before chapterIndex.
func (s *Storage) ListActiveEntitiesAtChapter(ctx context.Context, projectID string, chapterIndex int64) ([]Entity, error) {
	rows, err := s.q.ListActiveEntitiesAtChapter(ctx, sqlc.ListActiveEntitiesAtChapterParams{
		ProjectID:        projectID,
		FirstSeenChapter: chapterIndex,
	})
	if err != nil {
		return nil, err
	}

	out := make([]Entity, 0, len(rows))
	for _, r := range rows {
		out = append(out, *toDomainEntity(r))
	}
	return out, nil
}

// DeleteEntity deletes an entity by ID.
func (s *Storage) DeleteEntity(ctx context.Context, id string) error {
	return s.q.DeleteEntity(ctx, id)
}

// toDomainEntity converts sqlc.Entity database row to domain Entity struct.
func toDomainEntity(row sqlc.Entity) *Entity {
	var aliases []string
	if row.AliasesJson != "" && row.AliasesJson != "[]" {
		_ = jsonx.Unmarshal([]byte(row.AliasesJson), &aliases)
	}

	var meta map[string]any
	if row.MetadataJson != "" && row.MetadataJson != "{}" {
		_ = jsonx.Unmarshal([]byte(row.MetadataJson), &meta)
	}

	return &Entity{
		ID:               row.ID,
		ProjectID:        row.ProjectID,
		Name:             row.Name,
		Aliases:          aliases,
		Category:         row.Category,
		Gender:           row.Gender,
		Role:             row.Role,
		FirstSeenChapter: row.FirstSeenChapter,
		Metadata:         meta,
	}
}
