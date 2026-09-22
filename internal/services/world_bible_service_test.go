package services

import (
	"context"
	"path/filepath"
	"testing"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/skills"
	"novelclaw/pkg/storage"
)

type mockWbLLM struct {
	response string
}

func (m *mockWbLLM) ProviderName() string { return "mock-wb" }
func (m *mockWbLLM) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{Content: m.response}, nil
}
func (m *mockWbLLM) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: m.response, FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func TestWorldBibleService_CRUDAndLearn(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_wb_svc.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	projectID := "proj-svc-test"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "Cyberpunk 2077",
		Author:         "Mike P",
		SourceLang:     "en",
		TargetLang:     "vi",
		OriginalFormat: "txt",
	})

	registry := skills.NewRegistry(store)
	svc := NewWorldBibleService(store, registry, nil, nil, nil)
	svc.SetSkillStorage(skills.NewStorageManager(t.TempDir()))

	// Mock LLM client
	mockJSON := `{
  "rules": ["Tránh dịch Netrunner thành 'người chạy trên mạng'"],
  "few_shots": [{"input": "Netrunner breach", "output": "Netrunner xâm nhập", "note": "Giữ nguyên thuật ngữ"}],
  "summary": "Chuẩn hóa thuật ngữ Cyberpunk.",
  "target_skill_id": "skill_literary_translator"
}`
	SetTestClientWb(svc, &mockWbLLM{response: mockJSON})

	// 1. Upsert Category
	cat, err := svc.UpsertCategory(ctx, dtos.UpsertWorldCategoryRequest{
		ProjectID:    projectID,
		Slug:         "cyberware",
		Name:         "Cấy Ghép Công Nghệ",
		Icon:         "Cpu",
		Description:  "Trang bị công nghệ sinh học",
		DisplayOrder: 1,
	})
	if err != nil {
		t.Fatalf("UpsertCategory failed: %v", err)
	}
	if cat.Slug != "cyberware" {
		t.Errorf("expected slug cyberware, got %s", cat.Slug)
	}

	// 2. List Categories
	cats, err := svc.ListCategories(ctx, projectID)
	if err != nil || len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d, err: %v", len(cats), err)
	}

	// 3. Upsert Entry
	entry, err := svc.UpsertEntry(ctx, dtos.UpsertWorldEntryRequest{
		ProjectID:  projectID,
		CategoryID: cat.ID,
		Name:       "Sandevistan",
		Aliases:    []string{"Sandy", "Hệ thần kinh tăng tốc"},
		Summary:    "Cấy ghép tăng tốc độ phản xạ và dòng thời gian.",
		Attributes: map[string]string{
			"Tier": "Military Grade",
		},
		IsVerified: true,
	})
	if err != nil {
		t.Fatalf("UpsertEntry failed: %v", err)
	}

	// 4. List Entries
	entries, err := svc.ListEntries(ctx, projectID, "")
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d, err: %v", len(entries), err)
	}

	// 5. Verify toggle
	if err := svc.VerifyEntry(ctx, entry.ID, false); err != nil {
		t.Fatalf("VerifyEntry failed: %v", err)
	}

	// 6. Export Markdown & JSON
	md, err := svc.ExportWorldBibleMarkdown(ctx, projectID)
	if err != nil || md == "" {
		t.Fatalf("ExportWorldBibleMarkdown failed: %v", err)
	}

	jsonStr, err := svc.ExportWorldBibleJSON(ctx, projectID)
	if err != nil || jsonStr == "" {
		t.Fatalf("ExportWorldBibleJSON failed: %v", err)
	}

	// 7. Reset skills to default
	if err := svc.ResetSkillsToDefault(ctx, projectID); err != nil {
		t.Fatalf("ResetSkillsToDefault failed: %v", err)
	}

	// 8. Delete Entry & Category
	if err := svc.DeleteEntry(ctx, entry.ID); err != nil {
		t.Fatalf("DeleteEntry failed: %v", err)
	}
	if err := svc.DeleteCategory(ctx, cat.ID); err != nil {
		t.Fatalf("DeleteCategory failed: %v", err)
	}
}
