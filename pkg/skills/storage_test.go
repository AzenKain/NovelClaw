package skills

import (
	"context"
	"testing"
)

func TestStorageManager_HierarchicalAndReset(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewStorageManager(tempDir)
	projectID := "proj-test-reflexion"
	skillID := "skill_literary_translator"

	// 1. Initial state: no evolved data
	if sm.HasEvolvedData(projectID) {
		t.Errorf("expected no evolved data initially")
	}

	// 2. Save evolved rules
	rules := []string{
		"Không dịch 'ngươi' trong ngữ cảnh bạn bè hiện đại, phải dùng 'cậu' hoặc 'bạn'",
	}
	fewShots := []FewShotExample{
		{
			Input:  "你今天怎么了？",
			Output: "Hôm nay cậu sao thế?",
			Note:   "Văn phong tự nhiên không dùng đại từ cổ phong",
		},
	}

	err := sm.SaveEvolvedRules(projectID, skillID, rules, fewShots)
	if err != nil {
		t.Fatalf("SaveEvolvedRules failed: %v", err)
	}

	if !sm.HasEvolvedData(projectID) {
		t.Errorf("expected evolved data to exist after save")
	}

	// 3. Load evolved rules
	loadedRules, loadedFewShots, err := sm.LoadEvolvedRules(projectID, skillID)
	if err != nil {
		t.Fatalf("LoadEvolvedRules failed: %v", err)
	}
	if len(loadedRules) != 1 || loadedRules[0] != rules[0] {
		t.Errorf("expected 1 rule with exact text, got %v", loadedRules)
	}
	if len(loadedFewShots) != 1 || loadedFewShots[0].Output != "Hôm nay cậu sao thế?" {
		t.Errorf("expected 1 few-shot with output, got %v", loadedFewShots)
	}

	// 4. Save another rule with deduplication
	rules2 := []string{
		"Không dịch 'ngươi' trong ngữ cảnh bạn bè hiện đại, phải dùng 'cậu' hoặc 'bạn'",
		"Không dùng từ tu tiên trong truyện khoa học viễn tưởng",
	}
	err = sm.SaveEvolvedRules(projectID, skillID, rules2, nil)
	if err != nil {
		t.Fatalf("SaveEvolvedRules second batch failed: %v", err)
	}
	reloadedRules, _, _ := sm.LoadEvolvedRules(projectID, skillID)
	if len(reloadedRules) != 2 {
		t.Errorf("expected 2 rules after deduplicated merge, got %d", len(reloadedRules))
	}

	// 5. 1-Click Reset to Factory Defaults
	ctx := context.Background()
	if err := sm.ResetToFactoryDefaults(ctx, projectID); err != nil {
		t.Fatalf("ResetToFactoryDefaults failed: %v", err)
	}

	if sm.HasEvolvedData(projectID) {
		t.Errorf("expected no evolved data after reset")
	}

	postResetRules, postResetShots, err := sm.LoadEvolvedRules(projectID, skillID)
	if err != nil {
		t.Fatalf("LoadEvolvedRules post reset failed: %v", err)
	}
	if len(postResetRules) != 0 || len(postResetShots) != 0 {
		t.Errorf("expected empty rules post reset, got %d rules, %d shots", len(postResetRules), len(postResetShots))
	}
}
