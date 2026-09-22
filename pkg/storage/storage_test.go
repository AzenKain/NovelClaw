package storage_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

func setupTestDB(t *testing.T) *storage.Storage {
	t.Helper()
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory storage: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func TestOpenStorage_InMemory(t *testing.T) {
	store := setupTestDB(t)
	if store.DB() == nil {
		t.Fatal("expected db to be non-nil")
	}
	if store.Queries() == nil {
		t.Fatal("expected queries to be non-nil")
	}
}

func TestProjectCRUD(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	// 1. Create Project
	pParams := sqlc.CreateProjectParams{
		ID:             "proj-001",
		Title:          "That Time I Got Reincarnated as a Slime",
		Author:         "Fuse",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  10,
	}
	err := store.CreateProject(ctx, pParams)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// 2. Get Project
	proj, err := store.GetProject(ctx, "proj-001")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if proj.Title != pParams.Title || proj.Author != pParams.Author {
		t.Errorf("expected title %s, author %s; got %s, %s", pParams.Title, pParams.Author, proj.Title, proj.Author)
	}

	// 3. Update Project Progress
	err = store.UpdateProjectProgress(ctx, "proj-001", 12)
	if err != nil {
		t.Fatalf("UpdateProjectProgress failed: %v", err)
	}
	projUpdated, err := store.GetProject(ctx, "proj-001")
	if err != nil {
		t.Fatalf("GetProject after update failed: %v", err)
	}
	if projUpdated.TotalChapters != 12 {
		t.Errorf("expected total chapters 12, got %d", projUpdated.TotalChapters)
	}

	// 4. List Projects
	projects, err := store.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	// 5. Delete Project
	err = store.DeleteProject(ctx, "proj-001")
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	projects, err = store.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects after delete failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestChapterAndFTS5(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	// Setup Project
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             "proj-tensura",
		Title:          "Tensura",
		Author:         "Fuse",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  3,
	})

	// Add 3 Chapters with Atomic FTS indexing
	c1 := sqlc.CreateChapterParams{
		ID:           "chap-1",
		ProjectID:    "proj-tensura",
		ChapterIndex: 1,
		Title:        "Prologue: Death and Rebirth",
		ContentPath:  "text/c1.xhtml",
		RawContent:   "Rimuru Tempest was once a human named Satoru Mikami. He died saving his junior colleague and was reincarnated as a slime in a sealed cave with Great Sage skill.",
		Status:       "pending",
	}
	c2 := sqlc.CreateChapterParams{
		ID:           "chap-2",
		ProjectID:    "proj-tensura",
		ChapterIndex: 2,
		Title:        "Chapter 1: The Storm Dragon Veldora",
		ContentPath:  "text/c2.xhtml",
		RawContent:   "Deep inside the cave, Rimuru met the legendary Storm Dragon Veldora. Veldora had been sealed by a Hero for over 300 years.",
		Status:       "pending",
	}
	c3 := sqlc.CreateChapterParams{
		ID:           "chap-3",
		ProjectID:    "proj-tensura",
		ChapterIndex: 3,
		Title:        "Chapter 2: Goblin Village",
		ContentPath:  "text/c3.xhtml",
		RawContent:   "Leaving the cave, Rimuru encountered a tribe of weak goblins terrorized by the Direwolves.",
		Status:       "pending",
	}

	for _, c := range []sqlc.CreateChapterParams{c1, c2, c3} {
		if err := store.SaveChapterAndIndex(ctx, c); err != nil {
			t.Fatalf("SaveChapterAndIndex failed for %s: %v", c.ID, err)
		}
	}

	// 1. Check Chapter retrieval
	chap, err := store.GetChapterByIndex(ctx, "proj-tensura", 2)
	if err != nil {
		t.Fatalf("GetChapterByIndex failed: %v", err)
	}
	if chap.Title != "Chapter 1: The Storm Dragon Veldora" {
		t.Errorf("expected title Chapter 1, got %s", chap.Title)
	}

	// 2. FTS5 Search for "Veldora"
	results, err := store.SearchBookContext(ctx, "proj-tensura", "Veldora", 10)
	if err != nil {
		t.Fatalf("SearchBookContext failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'Veldora', got %d", len(results))
	}
	if results[0].ChapterID != "chap-2" {
		t.Errorf("expected chapter chap-2, got %s", results[0].ChapterID)
	}

	// 3. FTS5 Search Snippet for "Rimuru"
	snippets, err := store.SearchSnippet(ctx, "proj-tensura", "Rimuru", 10)
	if err != nil {
		t.Fatalf("SearchSnippet failed: %v", err)
	}
	// Rimuru appears in all 3 chapters
	if len(snippets) != 3 {
		t.Fatalf("expected 3 snippets for 'Rimuru', got %d", len(snippets))
	}
	for _, s := range snippets {
		if s.Snippet == "" {
			t.Errorf("expected non-empty snippet for chapter %s", s.ChapterID)
		}
	}

	// 4. Test Query Sanitization with weird/malformed symbols
	malformedQueries := []string{
		`"unclosed quote`,
		`Rimuru (AND Veldora*`,
		`NOT OR AND`,
		`Special @#$%^&* chars!`,
	}
	for _, mq := range malformedQueries {
		_, err := store.SearchBookContext(ctx, "proj-tensura", mq, 5)
		if err != nil {
			t.Errorf("SearchBookContext crashed on query %q: %v", mq, err)
		}
	}

	// 5. Update Chapter Translation
	err = store.UpdateChapterTranslation(ctx, "chap-1", "Rimuru Tempest từng là một con người...", "completed")
	if err != nil {
		t.Fatalf("UpdateChapterTranslation failed: %v", err)
	}
	chap1, err := store.GetChapter(ctx, "chap-1")
	if err != nil {
		t.Fatalf("GetChapter failed: %v", err)
	}
	if chap1.Status != "completed" || chap1.TranslatedContent == "" {
		t.Errorf("chapter 1 translation not updated properly: %+v", chap1)
	}

	// 6. Count Status
	statusCounts, err := store.CountChaptersByStatus(ctx, "proj-tensura")
	if err != nil {
		t.Fatalf("CountChaptersByStatus failed: %v", err)
	}
	countMap := make(map[string]int64)
	for _, sc := range statusCounts {
		countMap[sc.Status] = sc.Count
	}
	if countMap["completed"] != 1 || countMap["pending"] != 2 {
		t.Errorf("unexpected status counts: %+v", countMap)
	}
}

type MockTranslationState struct {
	StepIndex        int    `json:"step_index"`
	ActiveAgent      string `json:"active_agent"`
	LastSummary      string `json:"last_summary"`
	TokensConsumed   int64  `json:"tokens_consumed"`
}

func TestCheckpointManagement(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             "proj-cp",
		Title:          "CP Test",
		Author:         "Author",
		SourceLang:     "en",
		TargetLang:     "vi",
		OriginalFormat: "txt",
		TotalChapters:  5,
	})

	state1 := MockTranslationState{
		StepIndex:      1,
		ActiveAgent:    "PrimaryTranslator",
		LastSummary:    "Translated paragraph 1-10",
		TokensConsumed: 1540,
	}

	err := store.SaveCheckpoint(ctx, "proj-cp", 1, "step_translation", state1)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	state2 := MockTranslationState{
		StepIndex:      2,
		ActiveAgent:    "CriticReviewer",
		LastSummary:    "Reviewed pronoun consistency",
		TokensConsumed: 2100,
	}
	err = store.SaveCheckpoint(ctx, "proj-cp", 1, "step_critic", state2)
	if err != nil {
		t.Fatalf("SaveCheckpoint 2 failed: %v", err)
	}

	// Fetch latest checkpoint for project
	latestCP, err := store.GetLatestCheckpoint(ctx, "proj-cp")
	if err != nil {
		t.Fatalf("GetLatestCheckpoint failed: %v", err)
	}
	if latestCP.StepName != "step_critic" {
		t.Errorf("expected latest step_critic, got %s", latestCP.StepName)
	}

	var restoredState MockTranslationState
	if err := storage.UnpackCheckpointState(latestCP, &restoredState); err != nil {
		t.Fatalf("UnpackCheckpointState failed: %v", err)
	}
	if restoredState.TokensConsumed != 2100 || restoredState.ActiveAgent != "CriticReviewer" {
		t.Errorf("restored state mismatch: %+v", restoredState)
	}

	// Clear checkpoints
	if err := store.ClearCheckpointsByProject(ctx, "proj-cp"); err != nil {
		t.Fatalf("ClearCheckpointsByProject failed: %v", err)
	}
	_, err = store.GetLatestCheckpoint(ctx, "proj-cp")
	if err == nil {
		t.Error("expected error when getting checkpoint from empty project, got nil")
	}
}

func TestRelationsAndGlossary(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             "proj-rel",
		Title:          "Relation Test",
		Author:         "Author",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  10,
	})

	// 1. Add Relation: Rimuru -> Veldora at chapter 1
	rel1 := sqlc.UpsertRelationParams{
		ID:           "rel-rimuru-veldora-1",
		ProjectID:    "proj-rel",
		FromChar:     "Rimuru",
		ToChar:       "Veldora",
		CallAs:       "Huynh",
		SelfCallAs:   "Đệ",
		SinceChapter: 1,
		Tone:         "Thân thiết, kính trọng",
		IsLocked:     0,
	}
	if err := store.UpsertRelation(ctx, rel1); err != nil {
		t.Fatalf("UpsertRelation 1 failed: %v", err)
	}

	// Add updated Relation at chapter 5: Rimuru -> Veldora (Bằng hữu)
	rel2 := sqlc.UpsertRelationParams{
		ID:           "rel-rimuru-veldora-5",
		ProjectID:    "proj-rel",
		FromChar:     "Rimuru",
		ToChar:       "Veldora",
		CallAs:       "Bạn hiền",
		SelfCallAs:   "Tôi",
		SinceChapter: 5,
		Tone:         "Bình đẳng, chí cốt",
		IsLocked:     0,
	}
	if err := store.UpsertRelation(ctx, rel2); err != nil {
		t.Fatalf("UpsertRelation 2 failed: %v", err)
	}

	// Query relation at chapter 3 -> should return "Huynh"
	rAt3, err := store.GetRelationBetween(ctx, "proj-rel", "Rimuru", "Veldora", 3)
	if err != nil {
		t.Fatalf("GetRelationBetween at chapter 3 failed: %v", err)
	}
	if rAt3.CallAs != "Huynh" {
		t.Errorf("expected CallAs 'Huynh' at chap 3, got '%s'", rAt3.CallAs)
	}

	// Query relation at chapter 7 -> should return "Bạn hiền"
	rAt7, err := store.GetRelationBetween(ctx, "proj-rel", "Rimuru", "Veldora", 7)
	if err != nil {
		t.Fatalf("GetRelationBetween at chapter 7 failed: %v", err)
	}
	if rAt7.CallAs != "Bạn hiền" {
		t.Errorf("expected CallAs 'Bạn hiền' at chap 7, got '%s'", rAt7.CallAs)
	}

	// Lock relation
	if err := store.LockRelation(ctx, rAt7.ID); err != nil {
		t.Fatalf("LockRelation failed: %v", err)
	}

	// List relations
	relations, err := store.ListRelationsByProject(ctx, "proj-rel")
	if err != nil {
		t.Fatalf("ListRelationsByProject failed: %v", err)
	}
	if len(relations) != 2 {
		t.Errorf("expected 2 relations, got %d", len(relations))
	}

	// 2. Glossary Test
	g1 := sqlc.UpsertGlossaryTermParams{
		ID:         "glo-1",
		ProjectID:  "proj-rel",
		SourceTerm: "大賢者",
		TargetTerm: "Đại Hiền Giả",
		Category:   "Skill",
		Notes:      "Kỹ năng tối thượng",
	}
	if err := store.UpsertGlossaryTerm(ctx, g1); err != nil {
		t.Fatalf("UpsertGlossaryTerm failed: %v", err)
	}

	// Test ON CONFLICT DO UPDATE: cập nhật lại thuật ngữ cũ
	g1Updated := sqlc.UpsertGlossaryTermParams{
		ID:         "glo-1-new-id",
		ProjectID:  "proj-rel",
		SourceTerm: "大賢者",
		TargetTerm: "Đại Hiền Triết",
		Category:   "Skill",
		Notes:      "Đã được cập nhật nghĩa mới",
	}
	if err := store.UpsertGlossaryTerm(ctx, g1Updated); err != nil {
		t.Fatalf("UpsertGlossaryTerm update on conflict failed: %v", err)
	}

	glossaryList, err := store.ListGlossaryByProject(ctx, "proj-rel")
	if err != nil {
		t.Fatalf("ListGlossaryByProject failed: %v", err)
	}
	if len(glossaryList) != 1 || glossaryList[0].TargetTerm != "Đại Hiền Triết" {
		t.Errorf("glossary not updated properly: %+v", glossaryList)
	}

	if err := store.DeleteGlossaryTerm(ctx, glossaryList[0].ID); err != nil {
		t.Fatalf("DeleteGlossaryTerm failed: %v", err)
	}
}

func TestTimelineAndSettings(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             "proj-tl-set",
		Title:          "Timeline & Settings",
		Author:         "Author",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  5,
	})

	// 1. Timeline L1 test
	milestones := []string{"Rimuru gặp Veldora", "Đặt tên cho nhau"}
	err := store.UpsertChapterTimeline(ctx, "proj-tl-set", 1, "Chương 1 mở đầu câu chuyện chuyển sinh...", milestones)
	if err != nil {
		t.Fatalf("UpsertChapterTimeline failed: %v", err)
	}

	tl, err := store.GetChapterTimeline(ctx, "proj-tl-set", 1)
	if err != nil {
		t.Fatalf("GetChapterTimeline failed: %v", err)
	}
	if tl.SummaryText == "" || tl.ChapterIndex != 1 {
		t.Errorf("unexpected timeline: %+v", tl)
	}

	// 2. Settings test (Model Routing & Style Guide)
	modelRouting := map[string]string{
		"stage1_style_scout": "gemini-2.5-flash",
		"stage2_translate":   "deepseek-v3",
		"stage3_critic":      "claude-3-7-sonnet",
	}
	styleGuide := map[string]any{
		"tone": "văn phong kiếm hiệp / kỳ ảo",
	}
	err = store.UpsertProjectSettings(ctx, "proj-tl-set", styleGuide, modelRouting, "soul_pro_neko")
	if err != nil {
		t.Fatalf("UpsertProjectSettings failed: %v", err)
	}

	settings, err := store.GetProjectSettings(ctx, "proj-tl-set")
	if err != nil {
		t.Fatalf("GetProjectSettings failed: %v", err)
	}
	if settings.SoulID != "soul_pro_neko" {
		t.Errorf("expected soul_pro_neko, got %s", settings.SoulID)
	}
}

func TestCJK_FTS5_Search(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             "proj-cjk",
		Title:          "CJK Real Test",
		Author:         "Author",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  2,
	})

	// Chapter 1: Tiếng Nhật nguyên bản không có khoảng trắng
	jaText := `リムル＝テンペストはかつて三上悟という名の人間であった。通り魔に刺されて死んだ後、異世界でスライムとして転生した。スキル「大賢者」の導きにより、封印された暴風竜ヴェルドラと出会う。`
	c1 := sqlc.CreateChapterParams{
		ID:           "chap-ja",
		ProjectID:    "proj-cjk",
		ChapterIndex: 1,
		Title:        "第1章 死亡と転生",
		ContentPath:  "text/1.xhtml",
		RawContent:   jaText,
		Status:       "pending",
	}
	if err := store.SaveChapterAndIndex(ctx, c1); err != nil {
		t.Fatalf("SaveChapterAndIndex JA failed: %v", err)
	}

	// Chapter 2: Tiếng Trung nguyên bản không có khoảng trắng
	zhText := `利姆鲁·特恩佩斯特曾是一名普通的人类三上悟。为了保护后辈被歹徒刺死，转生到了异世界的洞窟中，成为了一只史莱姆。在大贤者的指引下，他遇到了被封印的暴风龙维鲁德拉。`
	c2 := sqlc.CreateChapterParams{
		ID:           "chap-zh",
		ProjectID:    "proj-cjk",
		ChapterIndex: 2,
		Title:        "第一章 死亡与转生",
		ContentPath:  "text/2.xhtml",
		RawContent:   zhText,
		Status:       "pending",
	}
	if err := store.SaveChapterAndIndex(ctx, c2); err != nil {
		t.Fatalf("SaveChapterAndIndex ZH failed: %v", err)
	}

	testQueries := []struct {
		query       string
		expectedIDs []string
	}{
		// Tìm kiếm Kanji tiếng Nhật trong câu viết liền
		{"三上悟", []string{"chap-ja", "chap-zh"}},
		{"大賢者", []string{"chap-ja"}},
		{"暴風竜", []string{"chap-ja"}},
		// Tìm kiếm Katakana tiếng Nhật
		{"スライム", []string{"chap-ja"}},
		{"ヴェルドラ", []string{"chap-ja"}},
		// Tìm kiếm tiếng Trung giản thể
		{"大贤者", []string{"chap-zh"}},
		{"暴风龙", []string{"chap-zh"}},
		{"利姆鲁", []string{"chap-zh"}},
		// Query ngắn (2 ký tự)
		{"人間", []string{"chap-ja"}},
		{"人类", []string{"chap-zh"}},
	}

	for _, tq := range testQueries {
		results, err := store.SearchBookContext(ctx, "proj-cjk", tq.query, 10)
		if err != nil {
			t.Errorf("SearchBookContext for %q failed: %v", tq.query, err)
			continue
		}
		if len(results) != len(tq.expectedIDs) {
			t.Errorf("query %q expected %d results, got %d", tq.query, len(tq.expectedIDs), len(results))
		}
	}
}

func BenchmarkFTS5_Search(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "neko_bench_*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dbFile := filepath.Join(tempDir, "bench.db")
	store, err := storage.OpenStorage(dbFile)
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             "bench-proj",
		Title:          "Bench Project",
		Author:         "Bench",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  100,
	})

	// Pre-populate 100 chương thực tế (~10KB mỗi chương, tổng ~1MB dữ liệu hỗn hợp CJK/EN/VI)
	baseParagraphJA := "リムル＝テンペストはかつて三上悟という名の人間であった。通り魔に刺されて死んだ後、異世界でスライムとして転生した。スキル「大賢者」の導きにより、封印された暴風竜ヴェルドラと出会う。無限牢獄を解析しながら、洞窟を出てゴブリンの村へ向かう。"
	baseParagraphZH := "利姆鲁·特恩佩斯特曾是一名普通的人类三上悟。为了保护后辈被歹徒刺死，转生到了异世界的洞窟中，成为了一只史莱姆。在大贤者的指引下，他遇到了被封印的暴风龙维鲁德拉。击败了牙狼族之后，成为了魔物联邦的领袖。"
	baseParagraphVI := "Sau khi được Đại Hiền Giả hỗ trợ thôn phệ Bạo Phong Long Veldora, Rimuru bắt đầu hành trình xây dựng quốc gia quái vật Tempest tại đại sâm lâm Jura, quy tụ yêu tinh, trư đầu tộc, quỷ tộc và giao hảo với các quốc gia nhân loại lân cận."

	for i := 1; i <= 100; i++ {
		// Nhân bản thành chương dài khoảng 1,500 từ
		var bodyBuilder strings.Builder
		for rep := 0; rep < 10; rep++ {
			bodyBuilder.WriteString(fmt.Sprintf("<p>Phần %d: %s %s %s</p>\n", rep+1, baseParagraphJA, baseParagraphZH, baseParagraphVI))
		}
		c := sqlc.CreateChapterParams{
			ID:           fmt.Sprintf("chap-%d", i),
			ProjectID:    "bench-proj",
			ChapterIndex: int64(i),
			Title:        fmt.Sprintf("Chương %d: Cuộc phiêu lưu tại đại sâm lâm Jura - 第%d章 転生物語", i, i),
			ContentPath:  fmt.Sprintf("text/%d.xhtml", i),
			RawContent:   bodyBuilder.String(),
			Status:       "pending",
		}
		if err := store.SaveChapterAndIndex(ctx, c); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark tìm kiếm thực chiến hỗn hợp cả CJK và Latin
		_, err := store.SearchSnippet(ctx, "bench-proj", "暴風竜", 5)
		if err != nil {
			b.Fatalf("SearchSnippet benchmark error JA: %v", err)
		}
		_, err = store.SearchSnippet(ctx, "bench-proj", "大贤者", 5)
		if err != nil {
			b.Fatalf("SearchSnippet benchmark error ZH: %v", err)
		}
	}
}

// TestSearchBookContextTemporal_SpoilerAndRerank tests temporal horizon enforcement and reranking.
func TestSearchBookContextTemporal_SpoilerAndRerank(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()
	projectID := "temporal-proj"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Temporal Test Book",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	chapters := []struct {
		idx     int64
		title   string
		content string
	}{
		{1, "Chapter 1: Origin", "羽島伊月と可児那由多が初めて出会った。那由多は天才作家である。"},
		{2, "Chapter 2: Development", "伊月は締め切りに追われていた。"},
		{3, "Chapter 3: Midpoint", "白川京がアパートに遊びに来た。"},
		{10, "Chapter 10: Future Climax", "可児那由多が重大な告白をする未来の結末。秘密コード999。"},
	}

	for _, ch := range chapters {
		err := store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
			ID:           fmt.Sprintf("t-chap-%d", ch.idx),
			ProjectID:    projectID,
			ChapterIndex: ch.idx,
			Title:        ch.title,
			RawContent:   ch.content,
			Status:       "pending",
		})
		if err != nil {
			t.Fatalf("SaveChapterAndIndex chap %d failed: %v", ch.idx, err)
		}
	}

	// 1. When translating Chapter 2 (currentChapter = 2), search for "那由多"
	// Chapter 10 (Future) MUST NOT appear in the results
	pastHits, err := store.SearchBookContextTemporal(ctx, projectID, "那由多", 2, 5)
	if err != nil {
		t.Fatalf("SearchBookContextTemporal failed: %v", err)
	}
	if len(pastHits) == 0 {
		t.Fatalf("expected hits in past chapters, got 0")
	}
	for _, h := range pastHits {
		cIdx, _ := strconv.ParseInt(h.ChapterIndex, 10, 64)
		if cIdx > 2 {
			t.Fatalf("SPOILER BUG DETECTED: Got future chapter %d when max allowed was 2: %s", cIdx, h.Title)
		}
	}

	// 2. Searching for future secret "秘密コード999" when at Chapter 2 MUST return 0 results
	spoilerSecretHits, err := store.SearchBookContextTemporal(ctx, projectID, "秘密コード999", 2, 5)
	if err != nil {
		t.Fatalf("SearchBookContextTemporal for secret failed: %v", err)
	}
	if len(spoilerSecretHits) != 0 {
		t.Fatalf("expected 0 hits for future secret, got %d", len(spoilerSecretHits))
	}

	// 3. When at Chapter 10, future secret SHOULD be found
	climaxHits, err := store.SearchBookContextTemporal(ctx, projectID, "秘密コード999", 10, 5)
	if err != nil {
		t.Fatalf("SearchBookContextTemporal at chap 10 failed: %v", err)
	}
	if len(climaxHits) != 1 {
		t.Fatalf("expected 1 hit at chap 10, got %d", len(climaxHits))
	}

	// 4. Test Reranking on Snippets
	snippetHits, err := store.SearchSnippetTemporal(ctx, projectID, "那由多", 2, 5)
	if err != nil {
		t.Fatalf("SearchSnippetTemporal failed: %v", err)
	}
	if len(snippetHits) == 0 {
		t.Fatalf("expected snippet hits, got 0")
	}
	if snippetHits[0].ChapterIndex != "1" {
		t.Errorf("expected top snippet to be chapter 1 (origin bonus), got %s", snippetHits[0].ChapterIndex)
	}
}

