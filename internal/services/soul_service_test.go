package services_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"novelclaw/internal/dtos"
	"novelclaw/internal/services"
	"novelclaw/pkg/soul"
)

func TestSoulService_CRUDAndImportExport(t *testing.T) {
	tmpDir := t.TempDir()
	ctrl := soul.NewControllerWithDir(tmpDir)
	svc := services.NewSoulService(ctrl)

	// 1. Initial list should have built-in souls
	initialList := svc.ListAvailableSouls()
	if len(initialList) < 2 {
		t.Fatalf("expected at least 2 souls, got %d", len(initialList))
	}

	// 2. Active soul initially should be default (NovelClaw-chan)
	active := svc.GetSoul()
	if active.ID != "soul_neko_assistant" {
		t.Errorf("expected default active soul 'soul_neko_assistant', got '%s'", active.ID)
	}

	// 3. Save a new custom soul via service
	newSoul := dtos.SoulDTO{
		ID:                "soul_sensei_wise",
		Name:              "Lão Tiên Sinh",
		Avatar:            "🧙‍♂️",
		Title:             "Cố Vấn Văn Học Cổ Điển",
		Archetype:         "Hiền triết thông tuệ",
		Description:       "Biên tập viên lão thành, hành văn uyên bác",
		Greeting:          "Lão phu xin chào các hạ.",
		OnConfused:        "Chỗ này ngữ pháp vi diệu, cần các hạ phân xử.",
		OnSuccess:         "Chương truyện hoàn mỹ, công đức viên mãn.",
		SystemTone:        "wise, solemn, and profound",
		SystemPromptAddon: "# Lão Tiên Sinh Persona\n- Dùng từ ngữ cổ kính uy nghiêm.",
		IsDefault:         false,
	}

	err := svc.SaveCustomSoul(newSoul)
	if err != nil {
		t.Fatalf("SaveCustomSoul failed: %v", err)
	}

	// 4. Verify soul was persisted to disk
	filePath := filepath.Join(tmpDir, "soul_sensei_wise.soul.md")
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("expected soul file '%s' to exist on disk", filePath)
	}

	// 5. Set active soul for project
	projectID := "proj_sensei_101"
	err = svc.SetActiveSoul(projectID, "soul_sensei_wise")
	if err != nil {
		t.Fatalf("SetActiveSoul failed: %v", err)
	}

	projActive := svc.GetActiveSoul(projectID)
	if projActive.ID != "soul_sensei_wise" {
		t.Errorf("expected active soul for project to be 'soul_sensei_wise', got '%s'", projActive.ID)
	}

	// 6. Export soul markdown
	exportedMD, err := svc.ExportSoulMarkdown("soul_sensei_wise")
	if err != nil {
		t.Fatalf("ExportSoulMarkdown failed: %v", err)
	}
	if !strings.Contains(exportedMD, "Lão Tiên Sinh") || !strings.Contains(exportedMD, "soul_sensei_wise") {
		t.Errorf("exported markdown missing soul metadata: %s", exportedMD)
	}

	// 7. Import soul markdown under new ID
	importedMD := strings.ReplaceAll(exportedMD, "soul_sensei_wise", "soul_sensei_clone")
	importedMD = strings.ReplaceAll(importedMD, "Lão Tiên Sinh", "Lão Tiên Sinh (Clone)")

	importedDTO, err := svc.ImportSoulMarkdown(importedMD)
	if err != nil {
		t.Fatalf("ImportSoulMarkdown failed: %v", err)
	}
	if importedDTO.ID != "soul_sensei_clone" || importedDTO.Name != "Lão Tiên Sinh (Clone)" {
		t.Errorf("imported DTO mismatch: %+v", importedDTO)
	}

	// 8. Delete custom soul
	err = svc.DeleteCustomSoul("soul_sensei_clone")
	if err != nil {
		t.Fatalf("DeleteCustomSoul failed: %v", err)
	}

	// 9. Reject delete default soul
	err = svc.DeleteCustomSoul("soul_neko_assistant")
	if err == nil {
		t.Error("expected error when deleting default soul, got nil")
	}
}
