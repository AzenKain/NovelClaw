package worldbible

import (
	"context"
	"path/filepath"
	"testing"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

func TestWorldBible_CRUDAndMatcher(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_wb.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	// 1. Create mock project
	projectID := "proj-wb-test-1"
	err = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "Test Novel World",
		Author:         "Author Cyber",
		SourceLang:     "zh",
		TargetLang:     "vi",
		OriginalFormat: "txt",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	wbService := NewService(store)

	// 2. Add Category
	cat, err := wbService.UpsertCategory(ctx, dtos.UpsertWorldCategoryRequest{
		ProjectID:    projectID,
		Slug:         "factions",
		Name:         "Tập Đoàn Công Nghệ",
		Icon:         "Building2",
		Description:  "Các thế lực tập đoàn công nghệ kiểm soát xã hội",
		DisplayOrder: 1,
	})
	if err != nil {
		t.Fatalf("UpsertCategory failed: %v", err)
	}
	if cat.Name != "Tập Đoàn Công Nghệ" {
		t.Errorf("expected category name 'Tập Đoàn Công Nghệ', got %s", cat.Name)
	}

	// 3. Add World Entry
	entry, err := wbService.UpsertEntry(ctx, dtos.UpsertWorldEntryRequest{
		ProjectID:  projectID,
		CategoryID: cat.ID,
		Name:       "Tập đoàn Arasaka",
		Aliases:    []string{"Arasaka", "Arasaka Corp"},
		Summary:    "Tập đoàn vũ khí và an ninh lớn nhất thành phố.",
		Attributes: map[string]string{
			"Cấp bậc":  "Megacorp",
			"Trụ sở": "Night City",
		},
		SourceChapterIndex: 1,
		IsVerified:         true,
	})
	if err != nil {
		t.Fatalf("UpsertEntry failed: %v", err)
	}
	if entry.Name != "Tập đoàn Arasaka" {
		t.Errorf("expected entry name 'Tập đoàn Arasaka', got %s", entry.Name)
	}

	// 4. Test Matcher
	matcher := NewMatcher(store)
	text := "Đêm đó, quân lính của Arasaka tiến vào khu ổ chuột."
	matches := matcher.MatchContext(ctx, projectID, text)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Name != "Tập đoàn Arasaka" {
		t.Errorf("expected matched entry name 'Tập đoàn Arasaka', got %s", matches[0].Name)
	}

	promptBlock := matcher.FormatContextPrompt(matches)
	if promptBlock == "" || !matcherContains(promptBlock, "Tập đoàn Arasaka") {
		t.Errorf("expected formatted prompt block to contain 'Tập đoàn Arasaka', got %s", promptBlock)
	}

	// 5. Test Export Markdown & JSON
	cats, err := wbService.ListCategories(ctx, projectID)
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	entries, err := wbService.ListEntries(ctx, projectID, "")
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(cats) != 1 || len(entries) != 1 {
		t.Fatalf("expected 1 cat and 1 entry, got %d cats, %d entries", len(cats), len(entries))
	}

	md := ExportMarkdown("Test Novel World", cats, entries)
	if !matcherContains(md, "WORLD BIBLE ENCYCLOPEDIA") {
		t.Errorf("expected markdown to have title, got %s", md)
	}

	jsonStr, err := ExportJSON(projectID, cats, entries)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// 6. Test Import JSON to new project
	projectID2 := "proj-wb-test-2"
	err = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID2,
		Title:          "Test Novel 2",
		Author:         "Author Cyber 2",
		SourceLang:     "zh",
		TargetLang:     "vi",
		OriginalFormat: "txt",
	})
	if err != nil {
		t.Fatalf("CreateProject 2 failed: %v", err)
	}
	catsImported, entriesImported, err := ImportJSON(ctx, store, projectID2, jsonStr)
	if err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}
	if catsImported != 1 || entriesImported != 1 {
		t.Errorf("expected 1 cat and 1 entry imported, got %d cats, %d entries", catsImported, entriesImported)
	}

	// 7. Verify Entry Toggle
	if err := wbService.VerifyEntry(ctx, entry.ID, false); err != nil {
		t.Fatalf("VerifyEntry failed: %v", err)
	}
	reloaded, err := store.GetWorldEntryByID(ctx, entry.ID)
	if err != nil {
		t.Fatalf("GetWorldEntryByID failed: %v", err)
	}
	if reloaded.IsVerified != 0 {
		t.Errorf("expected is_verified = 0, got %d", reloaded.IsVerified)
	}

	// 8. Delete Entry
	if err := wbService.DeleteEntry(ctx, entry.ID); err != nil {
		t.Fatalf("DeleteEntry failed: %v", err)
	}
	remainingEntries, _ := wbService.ListEntries(ctx, projectID, "")
	if len(remainingEntries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(remainingEntries))
	}
}

func matcherContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && len(s) > 0 && stringContains(s, sub)))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
