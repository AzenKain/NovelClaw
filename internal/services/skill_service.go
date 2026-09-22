package services

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/skills"
)

// SkillService exposes modular skill management, presets, and token footprint analytics to Wails v3.
type SkillService struct {
	registry *skills.Registry
}

// NewSkillService creates a new SkillService instance.
func NewSkillService(registry *skills.Registry) *SkillService {
	return &SkillService{
		registry: registry,
	}
}

// Registry returns the underlying Skill Registry.
func (s *SkillService) Registry() *skills.Registry {
	return s.registry
}

// ListSkills returns all available modular skills and their current state for a project.
func (s *SkillService) ListSkills(ctx context.Context, projectID string) ([]dtos.SkillDTO, error) {
	if s.registry == nil {
		return nil, fmt.Errorf("skill registry not initialized")
	}

	skillDefs := s.registry.GetSkills(ctx, projectID)
	result := make([]dtos.SkillDTO, len(skillDefs))
	for i, def := range skillDefs {
		result[i] = dtos.ToSkillDTO(def)
	}
	return result, nil
}

// ToggleSkill switches a specific skill on or off for a project.
func (s *SkillService) ToggleSkill(ctx context.Context, projectID string, skillID string, enabled bool) error {
	if s.registry == nil {
		return fmt.Errorf("skill registry not initialized")
	}

	if err := s.registry.ToggleSkill(ctx, projectID, skillID, enabled); err != nil {
		return err
	}

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("skill:updated", map[string]any{
			"project_id": projectID,
			"skill_id":   skillID,
			"enabled":    enabled,
		})
	}
	return nil
}

// UpdateSkillCustomRules saves user-defined prompt rules for a skill.
func (s *SkillService) UpdateSkillCustomRules(ctx context.Context, projectID string, skillID string, customRules string) error {
	if s.registry == nil {
		return fmt.Errorf("skill registry not initialized")
	}

	if err := s.registry.UpdateSkillCustomRules(ctx, projectID, skillID, customRules); err != nil {
		return err
	}

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("skill:updated", map[string]any{
			"project_id": projectID,
			"skill_id":   skillID,
		})
	}
	return nil
}

// ResetSkillToDefault restores a skill's rules and state to its built-in values.
func (s *SkillService) ResetSkillToDefault(ctx context.Context, projectID string, skillID string) error {
	if s.registry == nil {
		return fmt.Errorf("skill registry not initialized")
	}

	if err := s.registry.ResetSkillToDefault(ctx, projectID, skillID); err != nil {
		return err
	}

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("skill:updated", map[string]any{
			"project_id": projectID,
			"skill_id":   skillID,
		})
	}
	return nil
}

// ApplyCapabilityPreset sets the skills according to a 1-click bundle (Eco, Standard, Publishing).
func (s *SkillService) ApplyCapabilityPreset(ctx context.Context, projectID string, presetName string) error {
	if s.registry == nil {
		return fmt.Errorf("skill registry not initialized")
	}

	if err := s.registry.ApplyPreset(ctx, projectID, presetName); err != nil {
		return err
	}

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("skill:preset_applied", map[string]any{
			"project_id":  projectID,
			"preset_name": presetName,
		})
	}
	return nil
}

// GetSkillPresets returns all available capability presets.
func (s *SkillService) GetSkillPresets() []dtos.SkillPresetDTO {
	if s.registry == nil {
		return nil
	}

	presets := s.registry.GetPresets()
	dtosList := make([]dtos.SkillPresetDTO, len(presets))
	for i, p := range presets {
		dtosList[i] = dtos.ToSkillPresetDTO(p)
	}
	return dtosList
}

// GetActiveSkillsTokenWeight calculates the sum of estimated tokens for all active skills of a project.
func (s *SkillService) GetActiveSkillsTokenWeight(ctx context.Context, projectID string) int {
	if s.registry == nil {
		return 0
	}
	return s.registry.GetActiveSkillsTokenWeight(ctx, projectID)
}
