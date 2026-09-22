package novelclaw

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/llm"
)

type mockCompactLLM struct {
	summaryText string
}

func (m *mockCompactLLM) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{
		Content: m.summaryText,
	}, nil
}

func (m *mockCompactLLM) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: m.summaryText}
	close(ch)
	return ch, nil
}

func (m *mockCompactLLM) ProviderName() string {
	return "mock-compact"
}

func TestCompactor_ThresholdAndDistillation(t *testing.T) {
	store, projectID := setupTestStorage(t)
	defer store.Close()
	ctx := context.Background()

	thread, err := store.GetOrCreateMainThread(ctx, projectID)
	if err != nil {
		t.Fatalf("GetOrCreateMainThread failed: %v", err)
	}

	mockLLM := &mockCompactLLM{
		summaryText: "# [NOVELCLAW DISTILLED MEMORY CHUNK]\n- Tiêu Viêm gọi Dược Lão là Sư phụ.\n- Trúc Cơ = Foundation Establishment.",
	}

	// Create compactor with low threshold of 100 tokens
	compactor := NewCompactor(store, mockLLM, "test-model", 100)

	// Initially 0 tokens
	needed, tokens, err := compactor.NeedsCompact(ctx, thread.ID)
	if err != nil || needed || tokens != 0 {
		t.Fatalf("unexpected initial compact status: needed=%v, tokens=%d", needed, tokens)
	}

	// Insert several messages totaling 150 tokens
	for i := 0; i < 5; i++ {
		_, err := store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
			ID:                fmt.Sprintf("msg_%d", i),
			ThreadID:          thread.ID,
			ProjectID:         projectID,
			Sender:            "user",
			Role:              "user",
			Content:           fmt.Sprintf("Tin nhắn thảo luận số %d", i),
			ThinkingContent:   sql.NullString{Valid: false},
			StepType:          sql.NullString{Valid: false},
			StepStatus:        sql.NullString{String: "completed", Valid: true},
			ActionCallJson:    sql.NullString{Valid: false},
			IsCollapsed:       0,
			IsArchivedCompact: 0,
			TokenCount:        30,
		})
		if err != nil {
			t.Fatalf("SaveNovelClawMessage failed: %v", err)
		}
	}

	// Now tokens = 150 >= 100 -> Should need compact
	needed, tokens, err = compactor.NeedsCompact(ctx, thread.ID)
	if err != nil || !needed || tokens != 150 {
		t.Fatalf("expected needed=true, tokens=150, got needed=%v, tokens=%d", needed, tokens)
	}

	// Run auto-compact
	resp, err := compactor.CompactThread(ctx, thread.ID, false)
	if err != nil {
		t.Fatalf("CompactThread failed: %v", err)
	}

	if resp.ArchivedMessages != 5 {
		t.Errorf("expected 5 archived messages, got %d", resp.ArchivedMessages)
	}
	if resp.OriginalTokens != 150 {
		t.Errorf("expected original tokens 150, got %d", resp.OriginalTokens)
	}

	// Active messages in thread should now only contain the distilled summary and notice
	activeMsgs, err := store.ListActiveNovelClawMessages(ctx, thread.ID)
	if err != nil {
		t.Fatalf("ListActiveNovelClawMessages failed: %v", err)
	}
	if len(activeMsgs) != 2 {
		t.Fatalf("expected 2 active messages after compact (summary + notice), got %d", len(activeMsgs))
	}
	if activeMsgs[0].StepType.String != "compact_summary" {
		t.Errorf("expected step_type compact_summary, got %s", activeMsgs[0].StepType.String)
	}
}
