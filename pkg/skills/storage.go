package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"novelclaw/pkg/paths"
)

// StorageManager manages hierarchical skill persistence:
// - Default (Factory) immutable baseline: `skills/default/`
// - Evolved per-project mutations: `skills/evolved/<project_id>/`
type StorageManager struct {
	mu         sync.RWMutex
	baseDir    string
	evolvedDir string
}

// findSkillsDir resolves the canonical skills root. Explicit non-default
// paths win; otherwise the user-writable skills dir of the standard data
// layout is used (Windows: ./skills, mac/Linux: ~/.novelclaw/skills).
func findSkillsDir(dir string) string {
	if dir != "" && dir != "skills" && dir != "./skills" {
		return dir
	}
	return paths.Skills()
}

// NewStorageManager creates a StorageManager with the specified root directory.
func NewStorageManager(rootSkillsDir string) *StorageManager {
	rootSkillsDir = findSkillsDir(rootSkillsDir)
	base := filepath.Join(rootSkillsDir, "default")
	evolved := filepath.Join(rootSkillsDir, "evolved")

	return &StorageManager{
		baseDir:    base,
		evolvedDir: evolved,
	}
}

// ProjectEvolvedDir returns the directory path for a project's evolved skills.
func (sm *StorageManager) ProjectEvolvedDir(projectID string) string {
	return filepath.Join(sm.evolvedDir, projectID)
}

// SaveEvolvedRules saves newly distilled rules and examples for a project skill.
func (sm *StorageManager) SaveEvolvedRules(projectID string, skillID string, newRules []string, fewShots []FewShotExample) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	pDir := sm.ProjectEvolvedDir(projectID)
	if err := os.MkdirAll(pDir, 0755); err != nil {
		return fmt.Errorf("create project evolved dir: %w", err)
	}

	filePath := filepath.Join(pDir, fmt.Sprintf("%s.evolved.json", skillID))

	// Load existing data if present
	var existing struct {
		SkillID         string           `json:"skill_id"`
		ProjectID       string           `json:"project_id"`
		ProhibitedRules []string         `json:"prohibited_rules"`
		FewShotExamples []FewShotExample `json:"few_shot_examples"`
	}

	if data, err := os.ReadFile(filePath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	existing.SkillID = skillID
	existing.ProjectID = projectID

	// Append rules avoiding exact duplicates
	ruleSet := make(map[string]bool)
	for _, r := range existing.ProhibitedRules {
		ruleSet[r] = true
	}
	for _, r := range newRules {
		if !ruleSet[r] {
			existing.ProhibitedRules = append(existing.ProhibitedRules, r)
			ruleSet[r] = true
		}
	}

	// Append few-shots avoiding exact duplicates
	exampleSet := make(map[string]bool)
	for _, ex := range existing.FewShotExamples {
		exampleSet[ex.Input] = true
	}
	for _, ex := range fewShots {
		if !exampleSet[ex.Input] {
			existing.FewShotExamples = append(existing.FewShotExamples, ex)
			exampleSet[ex.Input] = true
		}
	}

	bytes, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal evolved skill: %w", err)
	}

	return os.WriteFile(filePath, bytes, 0644)
}

// LoadEvolvedRules loads any evolved rules and examples for a specific project and skill.
func (sm *StorageManager) LoadEvolvedRules(projectID string, skillID string) ([]string, []FewShotExample, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	filePath := filepath.Join(sm.ProjectEvolvedDir(projectID), fmt.Sprintf("%s.evolved.json", skillID))
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var evolved struct {
		ProhibitedRules []string         `json:"prohibited_rules"`
		FewShotExamples []FewShotExample `json:"few_shot_examples"`
	}
	if err := json.Unmarshal(data, &evolved); err != nil {
		return nil, nil, fmt.Errorf("unmarshal evolved rules: %w", err)
	}

	return evolved.ProhibitedRules, evolved.FewShotExamples, nil
}

// ResetToFactoryDefaults wipes out all evolved skills for a project, safely restoring factory baseline.
func (sm *StorageManager) ResetToFactoryDefaults(ctx context.Context, projectID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	pDir := sm.ProjectEvolvedDir(projectID)
	if _, err := os.Stat(pDir); os.IsNotExist(err) {
		return nil // Already at factory default
	}

	if err := os.RemoveAll(pDir); err != nil {
		return fmt.Errorf("reset to factory defaults failed: %w", err)
	}

	return nil
}

// HasEvolvedData checks whether a project has accumulated any evolved skills.
func (sm *StorageManager) HasEvolvedData(projectID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	pDir := sm.ProjectEvolvedDir(projectID)
	entries, err := os.ReadDir(pDir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}
