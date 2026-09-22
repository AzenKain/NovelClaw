package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/storage"
)

// Preset names for one-click capability configurations.
const (
	PresetEco        = "eco"
	PresetStandard   = "standard"
	PresetPublishing = "publishing"
)

// SkillOverride records project-specific overrides for a skill.
type SkillOverride struct {
	IsEnabled   *bool  `json:"is_enabled,omitempty"`
	CustomRules string `json:"custom_rules,omitempty"`
}

// CapabilityPreset defines a preset bundle of enabled skills and description.
type CapabilityPreset struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	ActiveSkillIDs       []string `json:"active_skill_ids"`
	TotalEstimatedTokens int      `json:"total_estimated_tokens"`
	ActiveSkillCount     int      `json:"active_skill_count"`
}

// Registry manages modular skills, project-specific settings, and dynamic prompt assembly.
type Registry struct {
	store     *storage.Storage
	builtin   map[string]SkillDefinition
	overrides map[string]map[string]SkillOverride // projectID -> skillID -> override
	mu        sync.RWMutex
}

// NewRegistry initializes the modular skill registry with the 6 standard skills.
func NewRegistry(store *storage.Storage) *Registry {
	builtinList := DefaultBuiltinSkills()
	builtinMap := make(map[string]SkillDefinition, len(builtinList))
	for _, s := range builtinList {
		builtinMap[s.ID] = s
	}

	return &Registry{
		store:     store,
		builtin:   builtinMap,
		overrides: make(map[string]map[string]SkillOverride),
	}
}

// GetPresets returns the list of available 1-click capability presets.
func (r *Registry) GetPresets() []CapabilityPreset {
	skills := DefaultBuiltinSkills()
	tokenMap := make(map[string]int)
	for _, s := range skills {
		tokenMap[s.ID] = s.EstimatedTokenWeight
	}

	calcTokens := func(ids []string) int {
		total := 0
		for _, id := range ids {
			total += tokenMap[id]
		}
		return total
	}

	ecoIDs := []string{"skill_literary_translator", "skill_foreign_sanitizer"}
	stdIDs := []string{"skill_style_scout", "skill_entity_extraction", "skill_literary_translator", "skill_agentic_researcher", "skill_foreign_sanitizer"}
	pubIDs := []string{"skill_style_scout", "skill_entity_extraction", "skill_literary_translator", "skill_shadow_critic", "skill_agentic_researcher", "skill_foreign_sanitizer"}

	return []CapabilityPreset{
		{
			ID:                   PresetEco,
			Name:                 "Eco Draft",
			Description:          "Token-optimized fast draft mode bypassing deep critique and heavy tools.",
			ActiveSkillIDs:       ecoIDs,
			TotalEstimatedTokens: calcTokens(ecoIDs),
			ActiveSkillCount:     len(ecoIDs),
		},
		{
			ID:                   PresetStandard,
			Name:                 "Standard Cowork",
			Description:          "Balanced quality and speed with entity extraction, RAG research, and literary translation.",
			ActiveSkillIDs:       stdIDs,
			TotalEstimatedTokens: calcTokens(stdIDs),
			ActiveSkillCount:     len(stdIDs),
		},
		{
			ID:                   PresetPublishing,
			Name:                 "Publishing Masterpiece",
			Description:          "Full autonomous agent stack with shadow critique, deep fact-checking, and literary polish.",
			ActiveSkillIDs:       pubIDs,
			TotalEstimatedTokens: calcTokens(pubIDs),
			ActiveSkillCount:     len(pubIDs),
		},
	}
}

// ensureLoaded loads project overrides from SQLite if not yet loaded in memory.
func (r *Registry) ensureLoaded(ctx context.Context, projectID string) {
	if projectID == "" || r.store == nil {
		return
	}

	r.mu.RLock()
	_, exists := r.overrides[projectID]
	r.mu.RUnlock()
	if exists {
		return
	}

	rawJSON, err := r.store.GetProjectSkillsConfig(ctx, projectID)
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.overrides[projectID]; exists {
		return
	}

	if err != nil || strings.TrimSpace(rawJSON) == "" || rawJSON == "{}" {
		r.overrides[projectID] = make(map[string]SkillOverride)
		return
	}

	var loaded map[string]SkillOverride
	if err := json.Unmarshal([]byte(rawJSON), &loaded); err != nil {
		log.Warn().Err(err).Str("project_id", projectID).Msg("failed to parse skills_config_json, using defaults")
		r.overrides[projectID] = make(map[string]SkillOverride)
		return
	}

	r.overrides[projectID] = loaded
}

// saveProjectConfig persists memory overrides to SQLite.
func (r *Registry) saveProjectConfig(ctx context.Context, projectID string) error {
	if projectID == "" || r.store == nil {
		return nil
	}

	r.mu.RLock()
	b, err := json.Marshal(r.overrides[projectID])
	r.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("marshal skills config: %w", err)
	}

	err = r.store.UpsertProjectSkillsConfig(ctx, projectID, string(b))
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			return nil
		}
		return fmt.Errorf("save project skills config: %w", err)
	}
	return nil
}

// GetSkills returns all skills for the project, applying any project-specific overrides.
func (r *Registry) GetSkills(ctx context.Context, projectID string) []SkillDefinition {
	r.ensureLoaded(ctx, projectID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	defaults := DefaultBuiltinSkills()
	result := make([]SkillDefinition, len(defaults))

	overrides := r.overrides[projectID]
	for i, s := range defaults {
		item := s
		if ov, ok := overrides[s.ID]; ok {
			if ov.IsEnabled != nil {
				item.IsEnabled = *ov.IsEnabled
			}
			if ov.CustomRules != "" {
				item.CustomRules = ov.CustomRules
			}
		}
		result[i] = item
	}

	return result
}

// IsSkillEnabled returns whether a skill is currently enabled for a project.
func (r *Registry) IsSkillEnabled(ctx context.Context, projectID string, skillID string) bool {
	r.ensureLoaded(ctx, projectID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if ov, ok := r.overrides[projectID][skillID]; ok && ov.IsEnabled != nil {
		return *ov.IsEnabled
	}

	if s, ok := r.builtin[skillID]; ok {
		return s.IsEnabled
	}

	return false
}

// ToggleSkill updates the enabled state of a skill for a project.
func (r *Registry) ToggleSkill(ctx context.Context, projectID string, skillID string, enabled bool) error {
	if _, ok := r.builtin[skillID]; !ok {
		return fmt.Errorf("skill '%s' not found", skillID)
	}

	r.ensureLoaded(ctx, projectID)

	r.mu.Lock()
	if r.overrides[projectID] == nil {
		r.overrides[projectID] = make(map[string]SkillOverride)
	}
	ov := r.overrides[projectID][skillID]
	ov.IsEnabled = &enabled
	r.overrides[projectID][skillID] = ov
	r.mu.Unlock()

	return r.saveProjectConfig(ctx, projectID)
}

// UpdateSkillCustomRules modifies the user-defined rules for a skill.
func (r *Registry) UpdateSkillCustomRules(ctx context.Context, projectID string, skillID string, customRules string) error {
	if _, ok := r.builtin[skillID]; !ok {
		return fmt.Errorf("skill '%s' not found", skillID)
	}

	r.ensureLoaded(ctx, projectID)

	r.mu.Lock()
	if r.overrides[projectID] == nil {
		r.overrides[projectID] = make(map[string]SkillOverride)
	}
	ov := r.overrides[projectID][skillID]
	ov.CustomRules = strings.TrimSpace(customRules)
	r.overrides[projectID][skillID] = ov
	r.mu.Unlock()

	return r.saveProjectConfig(ctx, projectID)
}

// ResetSkillToDefault restores a skill's settings to its factory defaults.
func (r *Registry) ResetSkillToDefault(ctx context.Context, projectID string, skillID string) error {
	if _, ok := r.builtin[skillID]; !ok {
		return fmt.Errorf("skill '%s' not found", skillID)
	}

	r.ensureLoaded(ctx, projectID)

	r.mu.Lock()
	if r.overrides[projectID] != nil {
		delete(r.overrides[projectID], skillID)
	}
	r.mu.Unlock()

	return r.saveProjectConfig(ctx, projectID)
}

// ApplyPreset enables/disables skills according to the specified capability preset.
func (r *Registry) ApplyPreset(ctx context.Context, projectID string, presetName string) error {
	var targetIDs []string
	switch strings.ToLower(presetName) {
	case PresetEco:
		targetIDs = []string{"skill_literary_translator", "skill_foreign_sanitizer"}
	case PresetStandard:
		targetIDs = []string{"skill_style_scout", "skill_entity_extraction", "skill_literary_translator", "skill_agentic_researcher", "skill_foreign_sanitizer"}
	case PresetPublishing:
		targetIDs = []string{"skill_style_scout", "skill_entity_extraction", "skill_literary_translator", "skill_shadow_critic", "skill_agentic_researcher", "skill_foreign_sanitizer"}
	default:
		return fmt.Errorf("unknown preset '%s'", presetName)
	}

	activeSet := make(map[string]bool)
	for _, id := range targetIDs {
		activeSet[id] = true
	}

	r.ensureLoaded(ctx, projectID)

	r.mu.Lock()
	if r.overrides[projectID] == nil {
		r.overrides[projectID] = make(map[string]SkillOverride)
	}
	for id := range r.builtin {
		enabled := activeSet[id]
		ov := r.overrides[projectID][id]
		ov.IsEnabled = &enabled
		r.overrides[projectID][id] = ov
	}
	r.mu.Unlock()

	return r.saveProjectConfig(ctx, projectID)
}

// GetActiveSkillPrompts compiles system prompt instructions for all currently enabled skills in a category.
func (r *Registry) GetActiveSkillPrompts(ctx context.Context, projectID string, category string) string {
	skills := r.GetSkills(ctx, projectID)
	var sb strings.Builder

	for _, s := range skills {
		if !s.IsEnabled {
			continue
		}
		if category != "" && s.Category != category {
			continue
		}

		sb.WriteString(fmt.Sprintf("\n[SKILL: %s]\n%s\n", s.Name, s.SystemPromptTemplate))
		if len(s.ProhibitedRules) > 0 {
			sb.WriteString("Prohibited Rules:\n")
			for _, rule := range s.ProhibitedRules {
				sb.WriteString(fmt.Sprintf("- %s\n", rule))
			}
		}
		if len(s.FewShots) > 0 {
			sb.WriteString("Exemplary Literary Patterns (Few-Shots):\n")
			for _, ex := range s.FewShots {
				fmt.Fprintf(&sb, "- Input: %s\n  Output: %s\n", ex.Input, ex.Output)
				if ex.Note != "" {
					fmt.Fprintf(&sb, "  Rationale: %s\n", ex.Note)
				}
			}
		}
		if strings.TrimSpace(s.CustomRules) != "" {
			fmt.Fprintf(&sb, "User Custom Rules:\n%s\n", s.CustomRules)
		}
	}

	return strings.TrimSpace(sb.String())
}

// GetActiveSkillsTokenWeight calculates the sum of estimated tokens across all enabled skills.
func (r *Registry) GetActiveSkillsTokenWeight(ctx context.Context, projectID string) int {
	skills := r.GetSkills(ctx, projectID)
	total := 0
	for _, s := range skills {
		if s.IsEnabled {
			total += s.EstimatedTokenWeight
		}
	}
	return total
}
