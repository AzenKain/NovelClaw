package novelclaw

import (
	"context"
	"path/filepath"
	"testing"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

type mockEmitter struct {
	dispatched []dtos.AppActionPayload
}

func (m *mockEmitter) EmitAction(payload dtos.AppActionPayload) {
	m.dispatched = append(m.dispatched, payload)
}

func setupTestStorage(t *testing.T) (*storage.Storage, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_novelclaw.db")
	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}

	projectID := "proj-test-nc"
	err = store.CreateProject(context.Background(), sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Test Project",
		SourceLang: "zh",
		TargetLang: "vi",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	return store, projectID
}

func TestAppTools_Definitions(t *testing.T) {
	defs := GetOmniAppToolDefinitions()
	if len(defs) != 28 {
		t.Fatalf("expected 28 tool definitions, got %d", len(defs))
	}

	toolMap := make(map[string]bool)
	for _, d := range defs {
		if d.Type != "function" {
			t.Errorf("expected tool type function, got %s", d.Type)
		}
		if d.Function.Name == "" {
			t.Errorf("empty function name")
		}
		toolMap[d.Function.Name] = true
	}

	expectedTools := []string{
		// Group 1
		"app_switch_tab", "app_select_volume", "app_select_chapter", "app_open_modal", "app_scroll_to_text",
		// Group 2
		"app_create_character", "app_set_relation_temporal", "app_scan_character_graph",
		// Group 3
		"app_scan_and_build_world", "app_upsert_world_category", "app_upsert_world_entry", "app_delete_world_entry",
		// Group 4
		"app_add_glossary_term", "app_bulk_import_glossary", "app_delete_glossary_term",
		// Group 5
		"app_update_translation_config", "app_configure_llm_provider", "app_set_agent_soul", "app_toggle_skill",
		// Group 6
		"app_start_translation", "app_pause_translation", "app_resume_translation", "app_soft_stop_translation", "app_abort_translation", "app_rollback_checkpoint",
		// Group 7
		"app_export_book", "app_trigger_learn_evolution", "app_run_benchmark",
	}

	for _, expected := range expectedTools {
		if !toolMap[expected] {
			t.Errorf("missing tool definition: %s", expected)
		}
	}
}

func TestAppTools_Execution(t *testing.T) {
	store, projectID := setupTestStorage(t)
	defer store.Close()

	emitter := &mockEmitter{}
	var startedChaps [2]int64
	hooks := AppToolsHooks{
		OnStartTranslation: func(ctx context.Context, pid string, startChap, endChap int64) error {
			startedChaps = [2]int64{startChap, endChap}
			return nil
		},
	}
	exec := NewAppToolsExecutor(store, emitter, hooks)
	ctx := context.Background()

	// 1. Test Navigation: app_switch_tab
	res, payload, err := exec.Execute(ctx, projectID, llm.ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "app_switch_tab",
			Arguments: `{"tab": "graph"}`,
		},
	})
	if err != nil {
		t.Fatalf("app_switch_tab failed: %v", err)
	}
	if payload == nil || payload.Action != "switch_tab" || payload.Data["tab"] != "graph" {
		t.Errorf("unexpected payload: %+v", payload)
	}
	if res == "" {
		t.Errorf("empty result text")
	}

	// 2. Test Character Graph: app_create_character
	res, payload, err = exec.Execute(ctx, projectID, llm.ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "app_create_character",
			Arguments: `{"name": "Tiêu Viêm", "role": "protagonist", "gender": "male"}`,
		},
	})
	if err != nil {
		t.Fatalf("app_create_character failed: %v", err)
	}
	if payload == nil || payload.Action != "character_created" {
		t.Errorf("unexpected payload: %+v", payload)
	}

	// 3. Test Temporal Character Relation: app_set_relation_temporal
	res, payload, err = exec.Execute(ctx, projectID, llm.ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "app_set_relation_temporal",
			Arguments: `{"volume_index": 2, "from_char": "Tiêu Viêm", "to_char": "Dược Lão", "call_as": "Sư phụ", "self_call_as": "Đồ nhi", "since_chapter": 25}`,
		},
	})
	if err != nil {
		t.Fatalf("app_set_relation_temporal failed: %v", err)
	}
	if payload == nil || payload.Action != "relation_updated" {
		t.Errorf("unexpected payload: %+v", payload)
	}

	// Verify in DB
	rels, err := store.ListRelationsByProject(ctx, projectID)
	if err != nil || len(rels) != 1 {
		t.Fatalf("expected 1 relation in DB, got %d (err: %v)", len(rels), err)
	}
	if rels[0].CallAs != "Sư phụ" || rels[0].SinceChapter != 25 {
		t.Errorf("unexpected relation in DB: %+v", rels[0])
	}

	// 4. Test World Bible: app_upsert_world_category and entry
	_, _, err = exec.Execute(ctx, projectID, llm.ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "app_upsert_world_category",
			Arguments: `{"slug": "megacorps", "name": "Tập Đoàn Công Nghệ", "icon": "Building"}`,
		},
	})
	if err != nil {
		t.Fatalf("app_upsert_world_category failed: %v", err)
	}

	_, _, err = exec.Execute(ctx, projectID, llm.ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "app_upsert_world_entry",
			Arguments: `{"category_slug": "megacorps", "name": "Arasaka", "summary": "Tập đoàn bảo an toàn cầu", "attributes_json": "{\"hq\":\"Night City\"}"}`,
		},
	})
	if err != nil {
		t.Fatalf("app_upsert_world_entry failed: %v", err)
	}

	// 5. Test Glossary: app_add_glossary_term
	_, _, err = exec.Execute(ctx, projectID, llm.ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "app_add_glossary_term",
			Arguments: `{"source_term": "Trúc Cơ", "target_term": "Foundation Establishment", "category": "realm"}`,
		},
	})
	if err != nil {
		t.Fatalf("app_add_glossary_term failed: %v", err)
	}

	terms, err := store.ListGlossaryByProject(ctx, projectID)
	if err != nil || len(terms) != 1 {
		t.Fatalf("expected 1 glossary term in DB, got %d (err: %v)", len(terms), err)
	}
	if terms[0].TargetTerm != "Foundation Establishment" {
		t.Errorf("unexpected glossary term: %+v", terms[0])
	}

	// 6. Test Translation Control: app_start_translation hook
	_, _, err = exec.Execute(ctx, projectID, llm.ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "app_start_translation",
			Arguments: `{"start_chapter": 5, "end_chapter": 10}`,
		},
	})
	if err != nil {
		t.Fatalf("app_start_translation failed: %v", err)
	}
	if startedChaps[0] != 5 || startedChaps[1] != 10 {
		t.Errorf("expected translation started for ch 5-10, got %v", startedChaps)
	}

	// Verify all events were captured by emitter
	if len(emitter.dispatched) < 6 {
		t.Errorf("expected at least 6 dispatched action events, got %d", len(emitter.dispatched))
	}
}

func TestAppTools_AllRemainingGroups(t *testing.T) {
	store, projectID := setupTestStorage(t)
	defer store.Close()
	ctx := context.Background()

	emitter := &mockEmitter{}
	var scanGraphCalled, scanWorldCalled, paused, resumed, softStopped, aborted, rolledBack, exported, learned, benchmarked bool
	hooks := AppToolsHooks{
		OnScanGraph: func(ctx context.Context, pid string, startChap, endChap int64) error {
			scanGraphCalled = true
			return nil
		},
		OnScanWorld: func(ctx context.Context, pid string, startChap, endChap int64) error {
			scanWorldCalled = true
			return nil
		},
		OnPauseTranslation: func(ctx context.Context, pid string) error {
			paused = true
			return nil
		},
		OnResumeTranslation: func(ctx context.Context, pid string) error {
			resumed = true
			return nil
		},
		OnSoftStop: func(ctx context.Context, pid string) error {
			softStopped = true
			return nil
		},
		OnAbort: func(ctx context.Context, pid string) error {
			aborted = true
			return nil
		},
		OnRollback: func(ctx context.Context, pid string, chapIdx int64, cpType string) error {
			rolledBack = true
			return nil
		},
		OnExportBook: func(ctx context.Context, pid string, fmt string, incCover bool, outPath string) error {
			exported = true
			return nil
		},
		OnTriggerLearn: func(ctx context.Context, pid string, notes string) error {
			learned = true
			return nil
		},
		OnRunBenchmark: func(ctx context.Context, pid string, chaps []int64, modes []string) error {
			benchmarked = true
			return nil
		},
	}

	exec := NewAppToolsExecutor(store, emitter, hooks)

	// Group 1: Navigation tools
	toolsToTest := []struct {
		name string
		args string
	}{
		{"app_select_volume", `{"volume_index": 2, "volume_tag": "Tập 2"}`},
		{"app_select_chapter", `{"chapter_index": 3}`},
		{"app_open_modal", `{"modal_name": "settings"}`},
		{"app_scroll_to_text", `{"keyword": "Dược Lão"}`},
		// Group 2 & 3: Scans & Delete
		{"app_scan_character_graph", `{"start_chapter": 1, "end_chapter": 5}`},
		{"app_scan_and_build_world", `{"start_chapter": 1, "end_chapter": 5}`},
		{"app_delete_world_entry", `{"entry_id": "test_ent"}`},
		// Group 4: Bulk glossary & delete
		{"app_bulk_import_glossary", `{"terms_json": "[{\"source_term\":\"A\",\"target_term\":\"B\",\"category\":\"general\"}]"}`},
		{"app_delete_glossary_term", `{"source_term": "A"}`},
		// Group 5: Config & LLM
		{"app_update_translation_config", `{"mode": "ConcurrentDualAgent", "enable_hot_patch": true, "enable_r19": true, "enable_agentic_rag": true, "style_guide": "Kiếm hiệp cổ điển"}`},
		{"app_configure_llm_provider", `{"provider": "gemini", "model_name": "gemini-2.5-pro", "api_key": "test_key", "is_default": true}`},
		{"app_set_agent_soul", `{"soul_id": "strict_editor"}`},
		{"app_toggle_skill", `{"skill_id": "literary_translator", "enabled": true}`},
		// Group 6: Execution
		{"app_pause_translation", `{}`},
		{"app_resume_translation", `{}`},
		{"app_soft_stop_translation", `{}`},
		{"app_abort_translation", `{}`},
		{"app_rollback_checkpoint", `{"chapter_index": 1, "checkpoint_type": "raw"}`},
		// Group 7: Export & Benchmark
		{"app_export_book", `{"format": "epub", "include_cover": true}`},
		{"app_trigger_learn_evolution", `{"notes": "Cập nhật cách xưng hô"}`},
		{"app_run_benchmark", `{"chapter_indices": [1, 2], "modes": ["single_pass"]}`},
	}

	for _, tt := range toolsToTest {
		res, payload, err := exec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      tt.name,
				Arguments: tt.args,
			},
		})
		if err != nil {
			t.Errorf("tool %s failed: %v", tt.name, err)
		}
		if res == "" {
			t.Errorf("tool %s returned empty result text", tt.name)
		}
		if payload == nil {
			t.Errorf("tool %s returned nil action payload", tt.name)
		}
	}

	// Verify hooks were invoked
	if !scanGraphCalled {
		t.Error("OnScanGraph was not called")
	}
	if !scanWorldCalled {
		t.Error("OnScanWorld was not called")
	}
	if !paused {
		t.Error("OnPauseTranslation was not called")
	}
	if !resumed {
		t.Error("OnResumeTranslation was not called")
	}
	if !softStopped {
		t.Error("OnSoftStop was not called")
	}
	if !aborted {
		t.Error("OnAbort was not called")
	}
	if !rolledBack {
		t.Error("OnRollback was not called")
	}
	if !exported {
		t.Error("OnExportBook was not called")
	}
	if !learned {
		t.Error("OnTriggerLearn was not called")
	}
	if !benchmarked {
		t.Error("OnRunBenchmark was not called")
	}

	// Verify LLM config was persisted in storage
	cfg, err := store.GetDefaultLLMConfig(ctx)
	if err != nil {
		t.Fatalf("GetDefaultLLMConfig failed: %v", err)
	}
	if cfg.ModelName != "gemini-2.5-pro" || cfg.ProviderName != "gemini" {
		t.Errorf("unexpected LLM config in store: %+v", cfg)
	}
}

