package llm

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

// mockAgenticClient simulates an LLM client that can trigger tool calls and then provide translation.
type mockAgenticClient struct {
	mu           sync.Mutex
	callCount    int
	lastRequest  CompletionRequest
	triggerTool  bool
	failWithTool bool
}

func (m *mockAgenticClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++
	m.lastRequest = req

	if m.failWithTool && len(req.Tools) > 0 {
		return nil, errors.New("model does not support function/tool declarations")
	}

	// First call: if tools are available and triggerTool is set, emit tool call
	if m.triggerTool && len(req.Tools) > 0 && m.callCount == 1 {
		var tc ToolCall
		tc.ID = "call_test_1"
		tc.Type = "function"
		tc.Function.Name = "search_book_context"
		tc.Function.Arguments = `{"query": "Dược Lão"}`

		return &CompletionResponse{
			Model:        req.Model,
			Thought:      "Cần tra cứu xem Dược Lão là ai trong các chương trước",
			ToolCalls:    []ToolCall{tc},
			PromptTokens: 50,
			CompTokens:   20,
			TotalTokens:  70,
		}, nil
	}

	// Subsequent call or direct call: return translated content
	return &CompletionResponse{
		Model:        req.Model,
		Content:      "Dược Lão thở dài một hơi, nhìn Tiêu Viêm mỉm cười.",
		PromptTokens: 80,
		CompTokens:   40,
		TotalTokens:  120,
	}, nil
}

func (m *mockAgenticClient) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: "Dịch..."}
	close(ch)
	return ch, nil
}

func (m *mockAgenticClient) ProviderName() string {
	return "mock_agentic"
}

func TestOrchestrator_AgenticRAG_ToolCalling(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "orch_rag_test.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_rag_test"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Đấu Phá Khung Thương",
		SourceLang: "zh",
		TargetLang: "vi",
	})

	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "chap_1",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "Chương 1",
		RawContent:   "药老叹了一口气，看着萧炎微笑了。",
		Status:       "pending",
	})

	client := &mockAgenticClient{triggerTool: true}
	orch := NewTranslationOrchestrator(client, "mock-model", store, nil)

	var thoughts []string
	var toolCallsReceived []string
	handler := &TranslationEventHandler{
		OnThought: func(thought string) {
			thoughts = append(thoughts, thought)
		},
		OnToolCall: func(name, input, output string) {
			toolCallsReceived = append(toolCallsReceived, name)
		},
	}

	opts := DefaultTranslationOptions()
	opts.Mode = ModeSinglePass
	opts.EnableAgenticRAG = true
	opts.MaxToolIterations = 2

	result, telemetry, err := orch.TranslateChapter(ctx, projectID, 1, opts, handler)
	if err != nil {
		t.Fatalf("TranslateChapter failed: %v", err)
	}

	if !strings.Contains(result, "Dược Lão") {
		t.Fatalf("expected translation containing 'Dược Lão', got: %s", result)
	}

	if len(thoughts) == 0 {
		t.Errorf("expected OnThought to be called")
	}

	if len(toolCallsReceived) == 0 || toolCallsReceived[0] != "search_book_context" {
		t.Errorf("expected OnToolCall for 'search_book_context', got: %v", toolCallsReceived)
	}

	if telemetry.ToolCallsCount != 1 {
		t.Errorf("expected telemetry.ToolCallsCount == 1, got %d", telemetry.ToolCallsCount)
	}

	// 2 calls for chunk translation (1 tool call + 1 completion) + 1 call for L1 chapter summary
	if client.callCount != 3 {
		t.Errorf("expected 3 LLM calls (2 for chunk with tool calling + 1 for L1 summary), got %d", client.callCount)
	}
}

func TestOrchestrator_AgenticRAG_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "orch_rag_disabled_test.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_rag_disabled"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Test",
		SourceLang: "zh",
		TargetLang: "vi",
	})

	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "chap_1",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "Chương 1",
		RawContent:   "你好世界",
		Status:       "pending",
	})

	client := &mockAgenticClient{triggerTool: false}
	orch := NewTranslationOrchestrator(client, "mock-model", store, nil)

	opts := DefaultTranslationOptions()
	opts.Mode = ModeSinglePass
	opts.EnableAgenticRAG = false // DISABLED

	result, telemetry, err := orch.TranslateChapter(ctx, projectID, 1, opts, nil)
	if err != nil {
		t.Fatalf("TranslateChapter failed: %v", err)
	}

	if result == "" {
		t.Fatalf("expected non-empty result")
	}

	if len(client.lastRequest.Tools) > 0 {
		t.Errorf("expected tools to be nil when EnableAgenticRAG = false")
	}

	if telemetry.ToolCallsCount != 0 {
		t.Errorf("expected ToolCallsCount == 0, got %d", telemetry.ToolCallsCount)
	}
}

func TestOrchestrator_ToolCalling_FallbackGraceful(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "orch_fallback_test.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_fallback"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Test",
		SourceLang: "zh",
		TargetLang: "vi",
	})

	_ = store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
		ID:           "chap_1",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "Chương 1",
		RawContent:   "你好世界",
		Status:       "pending",
	})

	// Client will error when tools are sent
	client := &mockAgenticClient{failWithTool: true}
	orch := NewTranslationOrchestrator(client, "mock-model", store, nil)

	opts := DefaultTranslationOptions()
	opts.Mode = ModeSinglePass
	opts.EnableAgenticRAG = true

	result, _, err := orch.TranslateChapter(ctx, projectID, 1, opts, nil)
	if err != nil {
		t.Fatalf("expected graceful fallback without error, got: %v", err)
	}

	if result == "" {
		t.Fatalf("expected translation output after graceful fallback")
	}
}
