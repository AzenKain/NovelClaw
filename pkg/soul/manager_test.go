package soul_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"novelclaw/pkg/soul"
)

func TestSoulParser_RoundTrip(t *testing.T) {
	original := soul.Soul{
		ID:                "soul_custom_hero",
		Name:              "Anh Hùng",
		Avatar:            "🛡️",
		Title:             "Chiến Binh Biên Tập",
		Archetype:         "Heroic Shounen",
		Description:       "Biên tập viên nhiệt huyết",
		Greeting:          "Cùng nhau chiến đấu nào!",
		OnConfused:        "Chỗ này có biến rồi!",
		OnSuccess:         "Chiến thắng rực rỡ!",
		SystemTone:        "shounen passionate",
		SystemPromptAddon: "# Persona Rules\n- Nhiệt huyết dâng trào\n- Luôn lạc quan",
		IsDefault:         false,
	}

	markdown, err := soul.FormatSoulMarkdown(original)
	if err != nil {
		t.Fatalf("FormatSoulMarkdown failed: %v", err)
	}

	if !strings.HasPrefix(markdown, "---") {
		t.Errorf("expected markdown to start with '---', got: %s", markdown)
	}

	parsed, err := soul.ParseSoulMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseSoulMarkdown failed: %v", err)
	}

	if parsed.ID != original.ID {
		t.Errorf("expected ID '%s', got '%s'", original.ID, parsed.ID)
	}
	if parsed.Name != original.Name {
		t.Errorf("expected Name '%s', got '%s'", original.Name, parsed.Name)
	}
	if parsed.Avatar != original.Avatar {
		t.Errorf("expected Avatar '%s', got '%s'", original.Avatar, parsed.Avatar)
	}
	if parsed.Greeting != original.Greeting {
		t.Errorf("expected Greeting '%s', got '%s'", original.Greeting, parsed.Greeting)
	}
	if !strings.Contains(parsed.SystemPromptAddon, "Nhiệt huyết dâng trào") {
		t.Errorf("expected SystemPromptAddon to contain 'Nhiệt huyết dâng trào', got: %s", parsed.SystemPromptAddon)
	}
}

func TestSoulManager_LifecycleAndPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := soul.NewManager(tmpDir)

	// 1. Check auto-generation of built-in souls
	souls := mgr.List()
	if len(souls) < 2 {
		t.Fatalf("expected at least 2 default souls, got %d", len(souls))
	}

	// 2. Check default soul retrieval
	defaultSoul, ok := mgr.Get("soul_neko_assistant")
	if !ok {
		t.Fatal("expected default soul to exist")
	}
	if defaultSoul.Name != "NovelClaw-chan" {
		t.Errorf("expected name 'NovelClaw-chan', got '%s'", defaultSoul.Name)
	}

	// 3. Save a new custom soul
	customSoul := soul.Soul{
		ID:                "soul_cyber_hacker",
		Name:              "Zero",
		Avatar:            "💻",
		Title:             "Cyberpunk Netrunner",
		Archetype:         "Cool Hacker",
		Description:       "Biên tập viên công nghệ cao",
		Greeting:          "System online. Sẵn sàng biên tập.",
		OnConfused:        "Lỗi cú pháp ngữ nghĩa, cần can thiệp thủ công.",
		OnSuccess:         "Upload hoàn tất 100%.",
		SystemTone:        "cyberpunk cool",
		SystemPromptAddon: "- Dùng thuật ngữ Sci-Fi hiện đại.",
		IsDefault:         false,
	}

	err := mgr.Save(customSoul)
	if err != nil {
		t.Fatalf("Save custom soul failed: %v", err)
	}

	// Verify file was written to disk
	expectedFile := filepath.Join(tmpDir, "soul_cyber_hacker.soul.md")
	if _, err := os.Stat(expectedFile); err != nil {
		t.Errorf("expected file '%s' to exist on disk", expectedFile)
	}

	// 4. Project binding
	projectID := "proj_cyberpunk_1"
	err = mgr.SetProjectSoul(projectID, "soul_cyber_hacker")
	if err != nil {
		t.Fatalf("SetProjectSoul failed: %v", err)
	}

	boundSoul := mgr.GetProjectSoul(projectID)
	if boundSoul.ID != "soul_cyber_hacker" {
		t.Errorf("expected bound soul 'soul_cyber_hacker', got '%s'", boundSoul.ID)
	}

	// Unbound project should fallback to default
	unboundSoul := mgr.GetProjectSoul("proj_unknown")
	if unboundSoul.ID != "soul_neko_assistant" {
		t.Errorf("expected fallback to default soul, got '%s'", unboundSoul.ID)
	}

	// 5. Try deleting default souls -> should both be rejected
	if err := mgr.Delete("soul_neko_assistant"); err == nil {
		t.Error("expected error when deleting soul_neko_assistant, got nil")
	}
	if err := mgr.Delete("soul_tieu_mai_wuxia"); err == nil {
		t.Error("expected error when deleting soul_tieu_mai_wuxia, got nil")
	}

	// 5b. Try saving malicious path traversal ID -> should be rejected
	maliciousSoul := soul.Soul{
		ID:   "../../etc_passwd",
		Name: "Hacker",
	}
	if err := mgr.Save(maliciousSoul); err == nil {
		t.Error("expected error when saving path traversal soul ID, got nil")
	}
	if err := mgr.Delete("../../etc_passwd"); err == nil {
		t.Error("expected error when deleting path traversal soul ID, got nil")
	}

	// 6. Delete custom soul -> should succeed
	err = mgr.Delete("soul_cyber_hacker")
	if err != nil {
		t.Fatalf("Delete custom soul failed: %v", err)
	}

	if _, ok := mgr.Get("soul_cyber_hacker"); ok {
		t.Error("expected custom soul to be deleted from memory")
	}
	if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
		t.Error("expected custom soul file to be deleted from disk")
	}

	// 7. Test RestoreDefaults
	if err := mgr.RestoreDefaults(); err != nil {
		t.Fatalf("RestoreDefaults failed: %v", err)
	}
	neko, ok := mgr.Get("soul_neko_assistant")
	if !ok || neko.Name != "NovelClaw-chan" {
		t.Errorf("RestoreDefaults failed to restore NovelClaw-chan")
	}
	mai, ok := mgr.Get("soul_tieu_mai_wuxia")
	if !ok || (mai.Name != "Xiao Mai" && mai.Name != "Tiểu Mai") {
		t.Errorf("RestoreDefaults failed to restore Xiao Mai")
	}
}
