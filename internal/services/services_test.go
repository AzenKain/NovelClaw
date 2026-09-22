package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/internal/services"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

func TestWailsServices_Stage3Integration(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open memory storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	projectSvc := services.NewProjectService(store)
	graphSvc := services.NewGraphService(store)
	glossarySvc := services.NewGlossaryService(store)
	llmSvc := services.NewLLMService(store)

	projectID := "proj_wails_test"

	err = projectSvc.CreateProject(ctx, dtos.CreateProjectRequest{
		ID:         projectID,
		Title:      "A Sister's All You Need (Wails Test)",
		SourceLang: "ja",
		TargetLang: "vi",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	err = llmSvc.SaveConfig(ctx, dtos.SaveLLMConfigRequest{
		ID:             "cfg_gemini",
		ProviderName:   "google_gemini",
		ApiURL:         "https://generativelanguage.googleapis.com",
		Token:          "secret-key-123456",
		ModelName:      "gemini-2.0-flash",
		IsActive:       true,
		IsDefault:      true,
		MaxRetries:     3,
		TimeoutSeconds: 120,
	})
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	cfgDTO, err := llmSvc.GetDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("GetDefaultConfig failed: %v", err)
	}
	if cfgDTO.MaskedToken == "secret-key-123456" || !strings.Contains(cfgDTO.MaskedToken, "...") {
		t.Errorf("token should be masked, got: %s", cfgDTO.MaskedToken)
	}
	if !cfgDTO.HasToken {
		t.Errorf("expected HasToken to be true")
	}

	err = llmSvc.SaveConfig(ctx, dtos.SaveLLMConfigRequest{
		ID:             "cfg_gemini",
		ProviderName:   "google_gemini",
		ApiURL:         "https://generativelanguage.googleapis.com",
		Token:          cfgDTO.MaskedToken,
		ModelName:      "gemini-2.0-flash-updated",
		IsActive:       true,
		IsDefault:      true,
		MaxRetries:     3,
		TimeoutSeconds: 120,
	})
	if err != nil {
		t.Fatalf("SaveConfig with masked token failed: %v", err)
	}

	cfgPreserved, err := store.GetLLMConfigByID(ctx, "cfg_gemini")
	if err != nil {
		t.Fatalf("GetLLMConfigByID failed: %v", err)
	}
	if cfgPreserved.Token != "secret-key-123456" {
		t.Errorf("expected original secret token preserved, got: %s", cfgPreserved.Token)
	}

	err = graphSvc.UpsertEntity(ctx, dtos.UpsertEntityRequest{
		ID:               "ent_itsuki",
		ProjectID:        projectID,
		Name:             "羽島伊月",
		Aliases:          []string{"伊月", "Itsuki"},
		Category:         "character",
		Gender:           "male",
		Role:             "Author",
		FirstSeenChapter: 1,
	})
	if err != nil {
		t.Fatalf("UpsertEntity failed: %v", err)
	}

	err = graphSvc.UpsertEntity(ctx, dtos.UpsertEntityRequest{
		ID:               "ent_nayuta",
		ProjectID:        projectID,
		Name:             "可児那由多",
		Aliases:          []string{"那由多", "Nayuta"},
		Category:         "character",
		Gender:           "female",
		Role:             "Author",
		FirstSeenChapter: 1,
	})
	if err != nil {
		t.Fatalf("UpsertEntity Nayuta failed: %v", err)
	}

	err = graphSvc.UpsertRelation(ctx, dtos.UpsertRelationRequest{
		ID:           "rel_nayuta_itsuki_c1",
		ProjectID:    projectID,
		FromChar:     "可児那由多",
		ToChar:       "羽島伊月",
		CallAs:       "先輩",
		SelfCallAs:   "わたし",
		SinceChapter: 1,
	})
	if err != nil {
		t.Fatalf("UpsertRelation failed: %v", err)
	}

	tsv := "TRPG\tGame nhập vai trên bàn\tgame\tNhập vai đổ xúc xắc\n"
	imported, err := glossarySvc.ImportTSV(ctx, projectID, tsv)
	if err != nil || imported != 1 {
		t.Fatalf("ImportTSV failed: imported=%d, err=%v", imported, err)
	}

	sceneChunk := "伊月が部屋で原稿を書いていると、那由多がTRPGのダイスを持って入ってきた。"
	selCtx, err := graphSvc.BuildSelectiveContext(ctx, projectID, 2, sceneChunk)
	if err != nil {
		t.Fatalf("BuildSelectiveContext failed: %v", err)
	}

	if len(selCtx.ActiveEntities) != 2 {
		t.Errorf("expected 2 active entities, got %d", len(selCtx.ActiveEntities))
	}
	if !strings.Contains(selCtx.FormattedPrompt, "先輩") {
		t.Errorf("expected address term '先輩' in prompt: %s", selCtx.FormattedPrompt)
	}
	if !strings.Contains(selCtx.FormattedPrompt, "Game nhập vai trên bàn") {
		t.Errorf("expected glossary term in prompt: %s", selCtx.FormattedPrompt)
	}
}

type mockScanClient struct{}

func (m *mockScanClient) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	jsonPayload := `{
		"characters": [
			{"name": "Itsuki Hashima", "gender": "male", "role": "protagonist", "aliases": ["Itsuki"]},
			{"name": "Nayuta Kani", "gender": "female", "role": "heroine", "aliases": ["Nayu", "Nayuta"]}
		],
		"relationships": [
			{"from_char": "Nayuta Kani", "to_char": "Itsuki Hashima", "call_as": "Tiền bối", "self_call_as": "Em", "tone": "thân mật"}
		]
	}`
	return &llm.CompletionResponse{
		Content: jsonPayload,
	}, nil
}

func (m *mockScanClient) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	return nil, nil
}

func (m *mockScanClient) ProviderName() string {
	return "mock"
}

func TestAutoScanEntities(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_autoscan.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_scan_test"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "A Sister's All You Need",
		Author:         "Yomi Hirasaka",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "txt",
		TotalChapters:  1,
	})

	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "ch1_scan",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "Concerning Nayuta Kani",
		ContentPath:  "ch1.txt",
		RawContent:   "Itsuki Hashima is a novelist. Nayuta Kani loves Itsuki and calls him Senpai.",
	})

	graphSvc := services.NewGraphService(store)
	services.SetGraphTestClient(graphSvc, &mockScanClient{})

	res, err := graphSvc.AutoScanEntities(ctx, dtos.AutoScanRequest{
		ProjectID:    projectID,
		ChapterIndex: 1,
		ScanMode:     "current_chapter",
	})
	if err != nil {
		t.Fatalf("AutoScanEntities failed: %v", err)
	}

	if !res.Success {
		t.Fatalf("AutoScanEntities returned failure: %s", res.Message)
	}
	if res.EntitiesFound != 2 {
		t.Errorf("expected 2 entities found, got %d", res.EntitiesFound)
	}
	if res.RelationsFound != 1 {
		t.Errorf("expected 1 relation found, got %d", res.RelationsFound)
	}

	// Verify database persistence
	ents, _ := store.ListEntitiesByProject(ctx, projectID)
	if len(ents) != 2 {
		t.Errorf("expected 2 entities in db, got %d", len(ents))
	}
	rels, _ := store.ListRelationsByProject(ctx, projectID)
	if len(rels) != 1 {
		t.Errorf("expected 1 relation in db, got %d", len(rels))
	}
	if rels[0].CallAs != "Tiền bối" {
		t.Errorf("expected CallAs 'Tiền bối', got '%s'", rels[0].CallAs)
	}
}

func TestAutoScanEntities_PreScanFilter(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_prescan.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_prescan_test"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "Imouto Test",
		Author:         "Yomi Hirasaka",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  3,
	})

	// Chapter 1 is a Cover with minimal text
	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "ch1_cover",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "Cover",
		ContentPath:  "ch1.xhtml",
		RawContent:   `<p><img src="cover.jpg" /></p>`,
	})

	// Chapter 2 is Table of Contents
	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "ch2_toc",
		ProjectID:    projectID,
		ChapterIndex: 2,
		Title:        "Table of Contents",
		ContentPath:  "ch2.xhtml",
		RawContent:   `<p>Contents: Chapter 1, Chapter 2</p>`,
	})

	// Chapter 3 is actual story chapter
	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "ch3_story",
		ProjectID:    projectID,
		ChapterIndex: 3,
		Title:        "The Novelist and His Friend",
		ContentPath:  "ch3.xhtml",
		RawContent:   `Itsuki Hashima was writing at his desk when Nayuta Kani arrived. Nayuta smiled happily and called him Senpai as usual. Chihiro Hashima brought them tea.`,
	})

	graphSvc := services.NewGraphService(store)
	services.SetGraphTestClient(graphSvc, &mockScanClient{})

	res, err := graphSvc.AutoScanEntities(ctx, dtos.AutoScanRequest{
		ProjectID: projectID,
		ScanMode:  "pre_scan",
	})
	if err != nil {
		t.Fatalf("AutoScanEntities pre_scan failed: %v", err)
	}
	if !res.Success {
		t.Fatalf("AutoScanEntities pre_scan returned failure: %s", res.Message)
	}
	if res.EntitiesFound != 2 {
		t.Errorf("expected 2 entities found from story chapter, got %d", res.EntitiesFound)
	}
}

func TestAppendBookToProject(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_append.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_append_test"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "My Light Novel Series",
		Author:         "Author",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "series",
		TotalChapters:  1,
	})

	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "proj_append_test_ch1",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "[vol1] Chapter 1",
		ContentPath:  "vol1/ch1.txt",
		RawContent:   "First volume content",
	})

	// Create a dummy txt file for vol2
	vol2Path := filepath.Join(tmpDir, "vol2.txt")
	if err := os.WriteFile(vol2Path, []byte("Chapter 2\nThis is volume 2 chapter 1."), 0644); err != nil {
		t.Fatalf("create vol2 file failed: %v", err)
	}

	projectSvc := services.NewProjectService(store)
	updatedProj, err := projectSvc.AppendBookToProject(ctx, dtos.AppendBookRequest{
		ProjectID: projectID,
		FilePath:  vol2Path,
		VolName:   "vol2",
	})
	if err != nil {
		t.Fatalf("AppendBookToProject failed: %v", err)
	}

	if updatedProj.TotalChapters < 2 {
		t.Errorf("expected at least 2 chapters, got %d", updatedProj.TotalChapters)
	}

	chaps, _ := store.ListChaptersByProject(ctx, projectID)
	if len(chaps) < 2 {
		t.Fatalf("expected at least 2 chapters in db, got %d", len(chaps))
	}
	if !strings.HasPrefix(chaps[1].Title, "[vol2]") {
		t.Errorf("expected chapter 2 title to have [vol2] prefix, got: %s", chaps[1].Title)
	}
}

func TestVolumeScopedAssetIsolation(t *testing.T) {
	projID := "proj_asset_test"
	baseDir := filepath.Join("data", "assets", projID)
	defer os.RemoveAll("data")

	vol1Dir := filepath.Join(baseDir, "vol1")
	vol2Dir := filepath.Join(baseDir, "vol2")
	if err := os.MkdirAll(vol1Dir, 0755); err != nil {
		t.Fatalf("mkdir vol1 failed: %v", err)
	}
	if err := os.MkdirAll(vol2Dir, 0755); err != nil {
		t.Fatalf("mkdir vol2 failed: %v", err)
	}

	vol1Cover := []byte("COVER_OF_VOLUME_1")
	vol2Cover := []byte("COVER_OF_VOLUME_2")

	if err := os.WriteFile(filepath.Join(vol1Dir, "cover.jpg"), vol1Cover, 0644); err != nil {
		t.Fatalf("write vol1 cover: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vol2Dir, "cover.jpg"), vol2Cover, 0644); err != nil {
		t.Fatalf("write vol2 cover: %v", err)
	}

	fallbackHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	handler := services.NewAssetHandler(nil, fallbackHandler)

	// Test 1: Request Volume 1 Cover
	req1 := httptest.NewRequest("GET", "/api/v1/reader/"+projID+"/asset/vol1/cover.jpg", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 for vol1 cover, got %d", rec1.Code)
	}
	if rec1.Body.String() != string(vol1Cover) {
		t.Errorf("expected vol1 cover data, got: %s", rec1.Body.String())
	}

	// Test 2: Request Volume 2 Cover (same filename, isolated path)
	req2 := httptest.NewRequest("GET", "/api/v1/reader/"+projID+"/asset/vol2/cover.jpg", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for vol2 cover, got %d", rec2.Code)
	}
	if rec2.Body.String() != string(vol2Cover) {
		t.Errorf("expected vol2 cover data, got: %s", rec2.Body.String())
	}

	// Test 3: Fallback discovery if volume prefix is omitted
	req3 := httptest.NewRequest("GET", "/api/v1/reader/"+projID+"/asset/cover.jpg", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200 for fallback search, got %d", rec3.Code)
	}
}

type mockVolLLM struct {
	response string
}

func (m *mockVolLLM) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{
		Content: m.response,
	}, nil
}

func (m *mockVolLLM) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: m.response}
	close(ch)
	return ch, nil
}

func (m *mockVolLLM) ProviderName() string {
	return "mock"
}


func TestGraphService_VolumeAwareScanAndEffectiveRelations(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open memory storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projID := "proj_multivol_graph"

	// 1. Create project
	err = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projID,
		Title:      "Imouto Sae Ireba Ii",
		SourceLang: "ja",
		TargetLang: "vi",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	// 2. Add chapters for Volume 1
	vol1Chaps := []struct {
		idx     int64
		title   string
		content string
	}{
		{1, "[vol1] Cover", "Image only"},
		{2, "[vol1] The Novelist Meets the Sister", "Nayuta mỉm cười chào: 'Senpai, anh có rảnh không?' Itsuki gật đầu: 'Anh đang bận viết thảo.'"},
		{3, "[vol1] Nayuta and Chihiro", "Chihiro chuẩn bị bữa tối ấm áp: 'Mời anh Itsuki và chị Nayuta dùng bữa.'"},
	}
	for _, c := range vol1Chaps {
		err := store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
			ID:           projID + "_ch" + string(rune('0'+c.idx)),
			ProjectID:    projID,
			ChapterIndex: c.idx,
			Title:        c.title,
			RawContent:   c.content,
		})
		if err != nil {
			t.Fatalf("save ch %d: %v", c.idx, err)
		}
	}

	// 3. Add chapters for Volume 2
	vol2Chaps := []struct {
		idx     int64
		title   string
		content string
	}{
		{4, "[vol2] Cover", "Image only"},
		{5, "[vol2] A New Relationship", "Nayuta ôm chầm lấy Itsuki: 'Darling! Em yêu anh nhất trần đời!' Itsuki đỏ mặt xoa đầu em."},
	}
	for _, c := range vol2Chaps {
		err := store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
			ID:           projID + "_ch" + string(rune('0'+c.idx)),
			ProjectID:    projID,
			ChapterIndex: c.idx,
			Title:        c.title,
			RawContent:   c.content,
		})
		if err != nil {
			t.Fatalf("save ch %d: %v", c.idx, err)
		}
	}

	graphSvc := services.NewGraphService(store)

	// 4. Test Volume 1 Auto-Scan
	vol1MockReply := `{
		"characters": [
			{"name": "Itsuki", "gender": "male", "role": "protagonist", "description": "Tác giả cuồng em gái", "facts": ["Thích em gái"]},
			{"name": "Nayuta", "gender": "female", "role": "heroine", "description": "Tác giả thiên tài", "facts": ["Đoạt giải tác giả mới"]}
		],
		"relationships": [
			{"from_char": "Nayuta", "to_char": "Itsuki", "relation": "Hậu bối", "call_as": "Tiền bối", "self_call_as": "Em", "tone": "Thân mật, ngưỡng mộ"}
		]
	}`
	services.SetGraphTestClient(graphSvc, &mockVolLLM{response: vol1MockReply})

	res1, err := graphSvc.AutoScanEntities(ctx, dtos.AutoScanRequest{
		ProjectID: projID,
		ScanMode:  "volume",
		Volume:    "vol1",
	})
	if err != nil {
		t.Fatalf("vol1 scan err: %v", err)
	}
	if !res1.Success {
		t.Fatalf("vol1 scan expected success, got msg: %s", res1.Message)
	}
	if res1.EffectiveChapter != 2 {
		t.Errorf("expected vol1 effective chapter to be 2 (first story chapter), got %d", res1.EffectiveChapter)
	}
	if res1.VolumeScanned != "vol1" {
		t.Errorf("expected volume scanned to be vol1, got %s", res1.VolumeScanned)
	}

	// 5. Test Volume 2 Auto-Scan (Relationships Evolve!)
	vol2MockReply := `{
		"characters": [
			{"name": "Itsuki", "gender": "male", "role": "protagonist", "description": "Tác giả cuồng em gái", "facts": ["Đã xuất bản 5 tập"]},
			{"name": "Nayuta", "gender": "female", "role": "heroine", "description": "Tác giả thiên tài", "facts": ["Bạn gái chính thức của Itsuki"]}
		],
		"relationships": [
			{"from_char": "Nayuta", "to_char": "Itsuki", "relation": "Người yêu", "call_as": "Darling", "self_call_as": "Em", "tone": "Yêu say đắm"}
		]
	}`
	services.SetGraphTestClient(graphSvc, &mockVolLLM{response: vol2MockReply})

	res2, err := graphSvc.AutoScanEntities(ctx, dtos.AutoScanRequest{
		ProjectID: projID,
		ScanMode:  "volume",
		Volume:    "vol2",
	})
	if err != nil {
		t.Fatalf("vol2 scan err: %v", err)
	}
	if !res2.Success {
		t.Fatalf("vol2 scan expected success, got msg: %s", res2.Message)
	}
	if res2.EffectiveChapter != 5 {
		t.Errorf("expected vol2 effective chapter to be 5, got %d", res2.EffectiveChapter)
	}

	// 6. Test Temporal Query: ListEffectiveRelationsAtChapter
	// Query at Chapter 2 (Vol 1): Nayuta must call Itsuki "Senpai"
	relsAtVol1, err := graphSvc.ListEffectiveRelationsAtChapter(ctx, projID, 2)
	if err != nil {
		t.Fatalf("list rels at vol1: %v", err)
	}
	if len(relsAtVol1) == 0 {
		t.Fatalf("expected at least 1 relation in Vol 1")
	}
	if relsAtVol1[0].CallAs != "Tiền bối" {
		t.Errorf("expected Nayuta to call Itsuki 'Tiền bối' in Vol 1, got '%s'", relsAtVol1[0].CallAs)
	}

	// Query at Chapter 5 (Vol 2): Nayuta must call Itsuki "Darling"
	relsAtVol2, err := graphSvc.ListEffectiveRelationsAtChapter(ctx, projID, 5)
	if err != nil {
		t.Fatalf("list rels at vol2: %v", err)
	}
	if len(relsAtVol2) == 0 {
		t.Fatalf("expected at least 1 relation in Vol 2")
	}
	if relsAtVol2[0].CallAs != "Darling" {
		t.Errorf("expected Nayuta to call Itsuki 'Darling' in Vol 2, got '%s'", relsAtVol2[0].CallAs)
	}

	// 7. Verify Facts Were Merged In Entity
	ents, err := graphSvc.ListEntities(ctx, projID)
	if err != nil {
		t.Fatalf("list entities: %v", err)
	}
	var nayutaEnt *dtos.EntityDTO
	for _, e := range ents {
		if e.Name == "Nayuta" {
			nayutaEnt = &e
			break
		}
	}
	if nayutaEnt == nil {
		t.Fatalf("Nayuta not found in entities")
	}
	facts, ok := nayutaEnt.Metadata["facts"].([]any)
	if !ok || len(facts) < 2 {
		t.Errorf("expected at least 2 merged facts for Nayuta, got: %v", nayutaEnt.Metadata["facts"])
	}

	// 8. Test DeleteRelation
	relToDelete := relsAtVol1[0].ID
	err = graphSvc.DeleteRelation(ctx, relToDelete)
	if err != nil {
		t.Fatalf("delete relation failed: %v", err)
	}

	// 9. Test CheckGraphReadiness
	readinessVol2, err := graphSvc.CheckGraphReadiness(ctx, projID, 5)
	if err != nil {
		t.Fatalf("check readiness vol 2: %v", err)
	}
	if !readinessVol2.IsReady || !readinessVol2.HasVolumeScan {
		t.Errorf("expected Vol 2 to be ready with volume scan, got: %+v", readinessVol2)
	}
	if readinessVol2.VolumeTag != "vol2" {
		t.Errorf("expected volume tag 'vol2', got '%s'", readinessVol2.VolumeTag)
	}
}


