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

// mockCriticLLM is a simple mock for the LLMClient used by ShadowCritic.
type mockCriticLLM struct {
	response string
	err      error
}

func (m *mockCriticLLM) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CompletionResponse{
		Content: m.response,
	}, nil
}

func (m *mockCriticLLM) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: m.response}
	close(ch)
	return ch, nil
}

func (m *mockCriticLLM) ProviderName() string {
	return "mock_hitl"
}

func TestToolRegistry_SetDecisionMode(t *testing.T) {
	tr := NewToolRegistry(nil, nil)

	// Default is manual
	if tr.decisionMode != "manual" {
		t.Errorf("expected default decisionMode to be manual, got %s", tr.decisionMode)
	}

	tr.SetDecisionMode("auto")
	if tr.decisionMode != "auto" {
		t.Errorf("expected decisionMode to be auto, got %s", tr.decisionMode)
	}

	tr.SetDecisionMode("")
	if tr.decisionMode != "auto" {
		t.Errorf("expected empty string to not override decisionMode, got %s", tr.decisionMode)
	}
}

func setupTestToolRegistry(t *testing.T) *ToolRegistry {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "hitl_test.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	researchEng := websearch.NewResearchEngine(1 * time.Second)
	return NewToolRegistry(store, researchEng)
}

func TestToolRegistry_AskHumanCoworker_ManualMode(t *testing.T) {
	tr := setupTestToolRegistry(t)
	tr.SetDecisionMode("manual")

	called := false
	tr.SetAskHumanHandler(func(ctx context.Context, question string, options []string, contextSnippet string) (string, error) {
		called = true
		return "I prefer option B", nil
	})

	call := ToolCall{
		ID:   "call_ask_human",
		Type: "function",
	}
	call.Function.Name = "ask_human_coworker"
	call.Function.Arguments = `{"question": "A or B?", "options": ["A", "B"], "context_snippet": "..."}`

	res, err := tr.Execute(context.Background(), "proj1", 1, call)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !called {
		t.Error("Expected AskHumanHandler to be called")
	}
	if !strings.Contains(res, "I prefer option B") {
		t.Errorf("Expected result to contain human answer, got: %s", res)
	}
}

func TestToolRegistry_AskHumanCoworker_ManualMode_NoHandler(t *testing.T) {
	tr := setupTestToolRegistry(t)
	tr.SetDecisionMode("manual")

	call := ToolCall{
		ID:   "call_ask_human",
		Type: "function",
	}
	call.Function.Name = "ask_human_coworker"
	call.Function.Arguments = `{"question": "A or B?", "options": ["Option A", "Option B"], "context_snippet": "..."}`

	res, err := tr.Execute(context.Background(), "proj1", 1, call)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(res, "Option A") {
		t.Errorf("Expected result to contain fallback to first option, got: %s", res)
	}
}

func TestToolRegistry_AskHumanCoworker_AutoMode(t *testing.T) {
	tr := setupTestToolRegistry(t)
	tr.SetDecisionMode("auto")

	mockLLM := &mockCriticLLM{
		response: "SELECTED: Option B\nRATIONALE: Because.",
	}
	critic := NewShadowCritic(mockLLM, "dummy-model", 1*time.Second)
	tr.SetShadowCritic(critic)

	call := ToolCall{
		ID:   "call_ask_human",
		Type: "function",
	}
	call.Function.Name = "ask_human_coworker"
	call.Function.Arguments = `{"question": "A or B?", "options": ["Option A", "Option B"], "context_snippet": "..."}`

	res, err := tr.Execute(context.Background(), "proj1", 1, call)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(res, "Option B") {
		t.Errorf("Expected result to contain auto-pilot decision for Option B, got: %s", res)
	}
}

func TestToolRegistry_AskHumanCoworker_AutoMode_NoCritic(t *testing.T) {
	tr := setupTestToolRegistry(t)
	tr.SetDecisionMode("auto")
	// Intentionally not setting shadow critic

	call := ToolCall{
		ID:   "call_ask_human",
		Type: "function",
	}
	call.Function.Name = "ask_human_coworker"
	call.Function.Arguments = `{"question": "A or B?", "options": ["Option A", "Option B"], "context_snippet": "..."}`

	res, err := tr.Execute(context.Background(), "proj1", 1, call)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(res, "Option A") {
		t.Errorf("Expected result to fallback to first option when no critic is set, got: %s", res)
	}
}

func TestToolRegistry_ArbitrateDilemma_NoCritic(t *testing.T) {
	tr := setupTestToolRegistry(t)
	// Critic is nil

	res, err := tr.ArbitrateDilemma(context.Background(), "Q", []string{"Opt1", "Opt2"}, "Ctx")
	if err != nil {
		t.Fatalf("ArbitrateDilemma failed: %v", err)
	}

	if res != "Opt1" {
		t.Errorf("Expected Opt1, got %s", res)
	}
}

func TestToolRegistry_ArbitrateDilemma_NoCriticNoOptions(t *testing.T) {
	tr := setupTestToolRegistry(t)
	// Critic is nil, empty options

	res, err := tr.ArbitrateDilemma(context.Background(), "Q", []string{}, "Ctx")
	if err != nil {
		t.Fatalf("ArbitrateDilemma failed: %v", err)
	}

	if res != "Default recommendation" {
		t.Errorf("Expected 'Default recommendation', got %s", res)
	}
}

// TestToolRegistry_AskHumanCoworker_AutoMode_AutoLearns verifies GAP-1:
// Auto-Pilot decisions made by Shadow Critic are automatically learned into storage.
func TestToolRegistry_AskHumanCoworker_AutoMode_AutoLearns(t *testing.T) {
	tr := setupTestToolRegistry(t)
	tr.SetDecisionMode("auto")

	projectID := "proj_hitl_autolearn"
	_ = tr.store.CreateProject(context.Background(), sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Auto-Learn Test Project",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	mockLLM := &mockCriticLLM{
		response: "SELECTED: Arthur\nRATIONALE: Standard fantasy transliteration.",
	}
	critic := NewShadowCritic(mockLLM, "dummy-model", 1*time.Second)
	tr.SetShadowCritic(critic)

	call := ToolCall{
		ID:   "call_ask_human",
		Type: "function",
	}
	call.Function.Name = "ask_human_coworker"
	call.Function.Arguments = `{"question": "Nên dịch tên \"アーサー\" thành gì?", "options": ["Arthur", "Asa"], "context_snippet": "アーサー đã rút thanh kiếm."}`

	res, err := tr.Execute(context.Background(), projectID, 1, call)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !strings.Contains(res, "Arthur") {
		t.Fatalf("Expected result to contain Arthur, got: %s", res)
	}

	// Verify that the decision was automatically recorded into L3 Glossary
	terms, err := tr.store.ListGlossaryByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("ListGlossaryByProject failed: %v", err)
	}
	if len(terms) == 0 {
		t.Fatalf("CRITICAL: Auto-Pilot decision was NOT learned into Glossary (0 terms found)")
	}
	if terms[0].SourceTerm != "アーサー" || terms[0].TargetTerm != "Arthur" {
		t.Errorf("Expected learned term 'アーサー' -> 'Arthur', got: '%s' -> '%s'", terms[0].SourceTerm, terms[0].TargetTerm)
	}
}

// TestToolRegistry_AutoLearnResolution_Heuristics verifies GAP-4:
// Questions containing options in quotes are NOT misclassified as character relations.
func TestToolRegistry_AutoLearnResolution_Heuristics(t *testing.T) {
	tr := setupTestToolRegistry(t)
	projectID := "proj_heuristic_test"
	_ = tr.store.CreateProject(context.Background(), sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Heuristics Test",
		SourceLang: "zh",
		TargetLang: "vi",
	})

	ctx := context.Background()

	// Scenario 1: Question quotes 3 terms: the source term and both candidate options
	questionWithOptionQuotes := `Nên dịch danh hiệu "魔王" thành "Ma Vương" hay "Chúa Tể"?`
	options := []string{"Ma Vương", "Chúa Tể"}
	answer := "[HITL] Tôi chọn: Ma Vương"

	tr.AutoLearnResolution(ctx, projectID, 5, questionWithOptionQuotes, answer, options)

	// Must be saved to Glossary, NOT Character Relations
	terms, err := tr.store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("ListGlossaryByProject failed: %v", err)
	}
	if len(terms) == 0 {
		t.Fatalf("Expected term '魔王' to be recorded in Glossary, got 0 terms")
	}
	if terms[0].SourceTerm != "魔王" || terms[0].TargetTerm != "Ma Vương" {
		t.Errorf("Expected '魔王' -> 'Ma Vương', got: '%s' -> '%s'", terms[0].SourceTerm, terms[0].TargetTerm)
	}

	rels, err := tr.store.ListRelationsByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("ListRelationsByProject failed: %v", err)
	}
	if len(rels) != 0 {
		t.Errorf("CRITICAL BUG-4: Misclassified glossary question as character relation: %+v", rels)
	}

	// Scenario 2: CleanAnswer helper functionality
	cleaned := CleanAnswer("[HITL] Tôi chọn: \"Arthur\"")
	if cleaned != "Arthur" {
		t.Errorf("Expected CleanAnswer to strip quotes and prefix, got: '%s'", cleaned)
	}
}

// TestToolRegistry_AskHumanCoworker_ContextCanceled verifies GAP-3:
// When context is canceled, Execute propagates the error instead of swallowing it.
func TestToolRegistry_AskHumanCoworker_ContextCanceled(t *testing.T) {
	tr := setupTestToolRegistry(t)
	tr.SetDecisionMode("manual")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled immediately

	tr.SetAskHumanHandler(func(hCtx context.Context, question string, options []string, contextSnippet string) (string, error) {
		return "", hCtx.Err()
	})

	call := ToolCall{
		ID:   "call_cancel",
		Type: "function",
	}
	call.Function.Name = "ask_human_coworker"
	call.Function.Arguments = `{"question": "A or B?", "options": ["A", "B"]}`

	_, err := tr.Execute(ctx, "proj_cancel", 1, call)
	if err == nil {
		t.Fatalf("Expected error when context is canceled, got nil")
	}
	if !strings.Contains(err.Error(), "canceled") {
		t.Errorf("Expected context canceled error, got: %v", err)
	}
}
