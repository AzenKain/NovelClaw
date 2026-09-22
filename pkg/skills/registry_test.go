package skills_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/skills"
	"novelclaw/pkg/storage"
)

func TestSkillRegistry_PresetsAndToggling(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_skills.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	reg := skills.NewRegistry(store)
	projectID := "proj_skill_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Test Novel for Skills",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	// 1. Check default skills
	allSkills := reg.GetSkills(ctx, projectID)
	if len(allSkills) != 6 {
		t.Fatalf("expected 6 default skills, got %d", len(allSkills))
	}

	// All 6 should be enabled by default
	for _, s := range allSkills {
		if !s.IsEnabled {
			t.Errorf("expected skill %s to be enabled by default", s.ID)
		}
	}

	// 2. Apply Eco Preset
	if err := reg.ApplyPreset(ctx, projectID, skills.PresetEco); err != nil {
		t.Fatalf("ApplyPreset eco failed: %v", err)
	}

	if reg.IsSkillEnabled(ctx, projectID, "skill_shadow_critic") {
		t.Error("expected skill_shadow_critic to be disabled in Eco preset")
	}
	if reg.IsSkillEnabled(ctx, projectID, "skill_agentic_researcher") {
		t.Error("expected skill_agentic_researcher to be disabled in Eco preset")
	}
	if !reg.IsSkillEnabled(ctx, projectID, "skill_literary_translator") {
		t.Error("expected skill_literary_translator to be enabled in Eco preset")
	}

	// 3. Update custom rules
	customRule := "Ưu tiên dịch các câu thoại ngắn theo phong cách kiếm hiệp cổ phong."
	if err := reg.UpdateSkillCustomRules(ctx, projectID, "skill_literary_translator", customRule); err != nil {
		t.Fatalf("UpdateSkillCustomRules failed: %v", err)
	}

	prompts := reg.GetActiveSkillPrompts(ctx, projectID, "translation")
	if !strings.Contains(prompts, customRule) {
		t.Errorf("expected active prompt to contain custom rule, got: %s", prompts)
	}

	// 4. Persistence check: create a new registry instance from the same DB
	reg2 := skills.NewRegistry(store)
	if reg2.IsSkillEnabled(ctx, projectID, "skill_shadow_critic") {
		t.Error("expected skill_shadow_critic to remain disabled across registry instances")
	}
	prompts2 := reg2.GetActiveSkillPrompts(ctx, projectID, "translation")
	if !strings.Contains(prompts2, customRule) {
		t.Errorf("expected persisted prompt to contain custom rule in new instance")
	}

	// 5. Reset to default
	if err := reg2.ResetSkillToDefault(ctx, projectID, "skill_literary_translator"); err != nil {
		t.Fatalf("ResetSkillToDefault failed: %v", err)
	}
	prompts3 := reg2.GetActiveSkillPrompts(ctx, projectID, "translation")
	if strings.Contains(prompts3, customRule) {
		t.Error("expected custom rule to be removed after reset to default")
	}

	// 6. Test Standard Preset
	if err := reg2.ApplyPreset(ctx, projectID, skills.PresetStandard); err != nil {
		t.Fatalf("ApplyPreset standard failed: %v", err)
	}
	if reg2.IsSkillEnabled(ctx, projectID, "skill_shadow_critic") {
		t.Error("expected skill_shadow_critic to be disabled in Standard preset")
	}
	if !reg2.IsSkillEnabled(ctx, projectID, "skill_style_scout") {
		t.Error("expected skill_style_scout to be enabled in Standard preset")
	}
	if !reg2.IsSkillEnabled(ctx, projectID, "skill_agentic_researcher") {
		t.Error("expected skill_agentic_researcher to be enabled in Standard preset")
	}
	stdWeight := reg2.GetActiveSkillsTokenWeight(ctx, projectID)
	if stdWeight != 1600 {
		t.Errorf("expected Standard preset weight 1600, got %d", stdWeight)
	}

	// 7. Test Publishing Preset
	if err := reg2.ApplyPreset(ctx, projectID, skills.PresetPublishing); err != nil {
		t.Fatalf("ApplyPreset publishing failed: %v", err)
	}
	pubWeight := reg2.GetActiveSkillsTokenWeight(ctx, projectID)
	if pubWeight != 2450 {
		t.Errorf("expected Publishing preset weight 2450, got %d", pubWeight)
	}
	for _, s := range reg2.GetSkills(ctx, projectID) {
		if !s.IsEnabled {
			t.Errorf("expected skill %s to be enabled in Publishing preset", s.ID)
		}
	}

	// 8. Test Presets metadata
	presets := reg2.GetPresets()
	if len(presets) != 3 {
		t.Fatalf("expected 3 presets, got %d", len(presets))
	}
}

func TestSkillRegistry_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_skills_race.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	reg := skills.NewRegistry(store)
	projectID := "proj_race_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Race Test Novel",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	done := make(chan bool)
	workers := 8
	iterations := 50

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			for i := 0; i < iterations; i++ {
				switch (workerID + i) % 4 {
				case 0:
					_ = reg.GetSkills(ctx, projectID)
				case 1:
					_ = reg.ToggleSkill(ctx, projectID, "skill_shadow_critic", (i%2 == 0))
				case 2:
					_ = reg.GetActiveSkillPrompts(ctx, projectID, "")
				case 3:
					_ = reg.GetActiveSkillsTokenWeight(ctx, projectID)
				}
			}
			done <- true
		}(w)
	}

	for w := 0; w < workers; w++ {
		<-done
	}
}
