package storage

import (
	"context"
	"fmt"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/jsonx"
)

// UpsertProjectSettings saves or updates project configuration including style guide, model routing matrix, and soul ID.
func (s *Storage) UpsertProjectSettings(ctx context.Context, projectID string, styleGuide any, modelRouting any, soulID string) error {
	if soulID == "" {
		soulID = "default_neko"
	}

	var styleGuideJSON string = "{}"
	if styleGuide != nil {
		b, err := jsonx.Marshal(styleGuide)
		if err != nil {
			return fmt.Errorf("marshal style_guide: %w", err)
		}
		styleGuideJSON = string(b)
	}

	var modelRoutingJSON string = "{}"
	if modelRouting != nil {
		b, err := jsonx.Marshal(modelRouting)
		if err != nil {
			return fmt.Errorf("marshal model_routing: %w", err)
		}
		modelRoutingJSON = string(b)
	}

	return s.q.UpsertProjectSettings(ctx, sqlc.UpsertProjectSettingsParams{
		ProjectID:        projectID,
		StyleGuideJson:   styleGuideJSON,
		ModelRoutingJson: modelRoutingJSON,
		SoulID:           soulID,
	})
}

// GetProjectSettings retrieves configuration settings for a project.
func (s *Storage) GetProjectSettings(ctx context.Context, projectID string) (sqlc.ProjectSetting, error) {
	return s.q.GetProjectSettings(ctx, projectID)
}

// DeleteProjectSettings deletes configuration settings for a project.
func (s *Storage) DeleteProjectSettings(ctx context.Context, projectID string) error {
	return s.q.DeleteProjectSettings(ctx, projectID)
}

// UpsertProjectSkillsConfig saves or updates project skills configuration JSON.
func (s *Storage) UpsertProjectSkillsConfig(ctx context.Context, projectID string, skillsConfigJSON string) error {
	if skillsConfigJSON == "" {
		skillsConfigJSON = "{}"
	}
	return s.q.UpsertProjectSkillsConfig(ctx, sqlc.UpsertProjectSkillsConfigParams{
		ProjectID:        projectID,
		SkillsConfigJson: skillsConfigJSON,
	})
}

// GetProjectSkillsConfig retrieves the skills configuration JSON for a project.
func (s *Storage) GetProjectSkillsConfig(ctx context.Context, projectID string) (string, error) {
	return s.q.GetProjectSkillsConfig(ctx, projectID)
}

