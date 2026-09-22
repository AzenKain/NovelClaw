package dtos

import "novelclaw/pkg/skills"

// FewShotExampleDTO holds a demonstration pair for a skill.
type FewShotExampleDTO struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Note   string `json:"note,omitempty"`
}

// SkillDTO transfers a modular skill's status, rules, and token consumption to the frontend.
type SkillDTO struct {
	ID                   string              `json:"id"`
	Name                 string              `json:"name"`
	Description          string              `json:"description"`
	Category             string              `json:"category"`
	EstimatedTokenWeight int                 `json:"estimated_token_weight"`
	WeightBadge          string              `json:"weight_badge"`
	IsEnabled            bool                `json:"is_enabled"`
	SystemPromptTemplate string              `json:"system_prompt_template"`
	ProhibitedRules      []string            `json:"prohibited_rules"`
	CustomRules          string              `json:"custom_rules"`
	FewShots             []FewShotExampleDTO `json:"few_shots"`
}

// SkillPresetDTO represents a pre-configured bundle of skills for quick 1-click toggling.
type SkillPresetDTO struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	ActiveSkillIDs       []string `json:"active_skill_ids"`
	TotalEstimatedTokens int      `json:"total_estimated_tokens"`
	ActiveSkillCount     int      `json:"active_skill_count"`
}

// ToSkillDTO converts a domain SkillDefinition into a transport DTO.
func ToSkillDTO(s skills.SkillDefinition) SkillDTO {
	fewShots := make([]FewShotExampleDTO, len(s.FewShots))
	for i, f := range s.FewShots {
		fewShots[i] = FewShotExampleDTO{
			Input:  f.Input,
			Output: f.Output,
			Note:   f.Note,
		}
	}

	return SkillDTO{
		ID:                   s.ID,
		Name:                 s.Name,
		Description:          s.Description,
		Category:             s.Category,
		EstimatedTokenWeight: s.EstimatedTokenWeight,
		WeightBadge:          s.WeightBadge,
		IsEnabled:            s.IsEnabled,
		SystemPromptTemplate: s.SystemPromptTemplate,
		ProhibitedRules:      s.ProhibitedRules,
		CustomRules:          s.CustomRules,
		FewShots:             fewShots,
	}
}

// ToSkillPresetDTO converts a CapabilityPreset to transport DTO.
func ToSkillPresetDTO(p skills.CapabilityPreset) SkillPresetDTO {
	return SkillPresetDTO{
		ID:                   p.ID,
		Name:                 p.Name,
		Description:          p.Description,
		ActiveSkillIDs:       p.ActiveSkillIDs,
		TotalEstimatedTokens: p.TotalEstimatedTokens,
		ActiveSkillCount:     p.ActiveSkillCount,
	}
}
