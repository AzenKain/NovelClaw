package llm

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/websearch"
)

// TestToolRegistry_Execute verifies tool definitions and execution for book search and web lookup.
func TestToolRegistry_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tools_test.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	projectID := "proj_tool_test"
	err = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Test Novel",
		SourceLang: "ja",
		TargetLang: "vi",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	err = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "chap_1",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "Chapter 1",
		RawContent:   "羽島伊月は妹バカの小説家である。白木屋でビールを飲むのが好きだ。",
		Status:       "pending",
	})
	if err != nil {
		t.Fatalf("SaveChapterAndIndex failed: %v", err)
	}

	researchEng := websearch.NewResearchEngine(10 * time.Second)
	toolReg := NewToolRegistry(store, researchEng)

	defs := toolReg.GetAvailableTools()
	if len(defs) < 2 {
		t.Fatalf("expected at least 2 tools, got %d", len(defs))
	}

	searchCall := ToolCall{
		ID:   "call_1",
		Type: "function",
	}
	searchCall.Function.Name = "search_book_context"
	searchCall.Function.Arguments = `{"query": "妹バカ"}`

	res, err := toolReg.Execute(ctx, projectID, 1, searchCall)
	if err != nil {
		t.Fatalf("Execute search_book_context failed: %v", err)
	}
	if !strings.Contains(res, "Chapter 1") {
		t.Errorf("search_book_context output missing Chapter 1: %s", res)
	}

	// Test Spoiler Barrier: Add future chapter 5 with unique keyword
	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "chap_future_5",
		ProjectID:    projectID,
		ChapterIndex: 5,
		Title:        "Chapter 5 (Future)",
		RawContent:   "千尋の秘密が明かされる未来の秘密コードXYZ",
		Status:       "completed",
	})

	spoilerCall := ToolCall{
		ID:   "call_spoiler",
		Type: "function",
	}
	spoilerCall.Function.Name = "search_book_context"
	spoilerCall.Function.Arguments = `{"query": "XYZ"}`

	// When current chapter is 1, future chapter 5 MUST NOT be returned (No Spoilers)
	spoilerRes, err := toolReg.Execute(ctx, projectID, 1, spoilerCall)
	if err != nil {
		t.Fatalf("Execute spoiler check failed: %v", err)
	}
	if strings.Contains(spoilerRes, "Chapter 5") || strings.Contains(spoilerRes, "未来の秘密コードXYZ") {
		t.Fatalf("CRITICAL SPOILER LEAK: Chapter 1 received future Chapter 5 context: %s", spoilerRes)
	}

	// When current chapter is 5 or later, chapter 5 CAN be returned
	validRes, err := toolReg.Execute(ctx, projectID, 5, spoilerCall)
	if err != nil {
		t.Fatalf("Execute valid chapter 5 check failed: %v", err)
	}
	if !strings.Contains(validRes, "Chapter 5") {
		t.Errorf("expected Chapter 5 to be found when current chapter is 5: %s", validRes)
	}

	webCall := ToolCall{
		ID:   "call_2",
		Type: "function",
	}
	webCall.Function.Name = "web_lookup"
	webCall.Function.Arguments = `{"keyword": "Light novel", "language": "ja"}`

	webRes, err := toolReg.Execute(ctx, projectID, 1, webCall)
	if err != nil {
		t.Fatalf("Execute web_lookup failed: %v", err)
	}
	if !strings.Contains(webRes, "Light novel") {
		t.Errorf("web_lookup output missing keyword: %s", webRes)
	}
}

func TestToolRegistry_CharacterRelationAndWorldLore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tools_relation_lore_test.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_lore_test"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Cyber Cultivation",
		SourceLang: "zh",
		TargetLang: "vi",
	})

	// 1. Setup Entities and Relations for lookup_character_relation
	_ = store.UpsertEntity(ctx, storage.Entity{
		ID:        "ent_tieu_viem",
		ProjectID: projectID,
		Name:      "Tiêu Viêm",
		Gender:    "nam",
		Role:      "protagonist",
	})
	_ = store.UpsertEntity(ctx, storage.Entity{
		ID:        "ent_nha_phi",
		ProjectID: projectID,
		Name:      "Nhã Phi",
		Gender:    "nữ",
		Role:      "elder",
	})
	_ = store.UpsertRelation(ctx, sqlc.UpsertRelationParams{
		ID:           "rel_tv_np",
		ProjectID:    projectID,
		FromChar:     "Tiêu Viêm",
		ToChar:       "Nhã Phi",
		CallAs:       "Nhã Phi tỷ",
		SelfCallAs:   "đệ đệ",
		SinceChapter: 1,
		Tone:         "kính trọng, thân thiết",
	})

	// 2. Setup World Category and Entry for lookup_world_lore
	cat, _ := store.UpsertWorldCategory(ctx, sqlc.UpsertWorldCategoryParams{
		ID:          "cat_megacorps",
		ProjectID:   projectID,
		Slug:        "megacorps",
		Name:        "Tập Đoàn Siêu Cấp",
		Icon:        "Cpu",
		Description: "Các tập đoàn cai trị thế giới",
	})
	_, _ = store.UpsertWorldEntry(ctx, sqlc.UpsertWorldEntryParams{
		ID:                 "entry_arasaka",
		ProjectID:          projectID,
		CategoryID:         cat.ID,
		Name:               "Arasaka Corp",
		AliasesJson:        `["Tập Đoàn Arasaka", "Hoang Bản"]`,
		Summary:            "Tập đoàn quân sự và an ninh mạng thống trị bầu trời",
		FullDescription:    "Tập đoàn được sáng lập bởi Saburo Arasaka, độc quyền công nghệ linh hồn Soulkiller",
		AttributesJson:     `{"tier":"S-Rank","leader":"Saburo Arasaka"}`,
		DiscoveredBy:       "manual",
		SourceChapterIndex: 1,
		IsVerified:         1,
	})

	toolReg := NewToolRegistry(store, nil)

	// Test lookup_character_relation
	relCall := ToolCall{
		ID:   "call_rel",
		Type: "function",
	}
	relCall.Function.Name = "lookup_character_relation"
	relCall.Function.Arguments = `{"character_a": "Tiêu Viêm", "character_b": "Nhã Phi"}`

	relRes, err := toolReg.Execute(ctx, projectID, 5, relCall)
	if err != nil {
		t.Fatalf("Execute lookup_character_relation failed: %v", err)
	}
	if !strings.Contains(relRes, "Nhã Phi tỷ") || !strings.Contains(relRes, "đệ đệ") {
		t.Errorf("lookup_character_relation missing address terms: %s", relRes)
	}
	if !strings.Contains(relRes, "protagonist") || !strings.Contains(relRes, "elder") {
		t.Errorf("lookup_character_relation missing character roles: %s", relRes)
	}

	// Test lookup_world_lore
	loreCall := ToolCall{
		ID:   "call_lore",
		Type: "function",
	}
	loreCall.Function.Name = "lookup_world_lore"
	loreCall.Function.Arguments = `{"query": "Arasaka"}`

	loreRes, err := toolReg.Execute(ctx, projectID, 5, loreCall)
	if err != nil {
		t.Fatalf("Execute lookup_world_lore failed: %v", err)
	}
	if !strings.Contains(loreRes, "Arasaka Corp") {
		t.Errorf("lookup_world_lore missing entity name: %s", loreRes)
	}
	if !strings.Contains(loreRes, "Tập Đoàn Siêu Cấp") {
		t.Errorf("lookup_world_lore missing category name: %s", loreRes)
	}
	if !strings.Contains(loreRes, "S-Rank") {
		t.Errorf("lookup_world_lore missing attributes: %s", loreRes)
	}
}

