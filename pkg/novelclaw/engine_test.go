package novelclaw

import (
	"context"
	"strings"
	"testing"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/llm"
)

type mockStreamEmitter struct {
	events []struct {
		eventType string
		data      map[string]any
	}
}

func (m *mockStreamEmitter) EmitStreamEvent(eventType string, data map[string]any) {
	m.events = append(m.events, struct {
		eventType string
		data      map[string]any
	}{eventType: eventType, data: data})
}

type mockEngineLLM struct {
	responses []*llm.CompletionResponse
	callIndex int
}

func (m *mockEngineLLM) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.callIndex >= len(m.responses) {
		return &llm.CompletionResponse{Content: "Mặc định hoàn thành"}, nil
	}
	resp := m.responses[m.callIndex]
	m.callIndex++
	return resp, nil
}

func (m *mockEngineLLM) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "Stream response"}
	close(ch)
	return ch, nil
}

func (m *mockEngineLLM) ProviderName() string {
	return "mock-engine-llm"
}

func TestEngine_SlashCommands(t *testing.T) {
	store, projectID := setupTestStorage(t)
	defer store.Close()
	ctx := context.Background()

	thread, err := store.GetOrCreateMainThread(ctx, projectID)
	if err != nil {
		t.Fatalf("GetOrCreateMainThread failed: %v", err)
	}

	emitter := &mockStreamEmitter{}
	engine := NewEngine(Config{
		Store:        store,
		Client:       &mockEngineLLM{},
		ModelName:    "test-model",
		SkillsDir:    "../../skills/default",
		Emitter:      emitter,
		CompactLimit: 250000,
	})

	// 1. Test /help
	msg, err := engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "/help",
	})
	if err != nil {
		t.Fatalf("/help failed: %v", err)
	}
	if !strings.Contains(msg.Content, "NovelClaw Slash Commands") {
		t.Errorf("unexpected /help response: %s", msg.Content)
	}

	// 2. Test /term
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "/term Trúc Cơ = Foundation Establishment realm",
	})
	if err != nil {
		t.Fatalf("/term failed: %v", err)
	}
	if !strings.Contains(msg.Content, "Added glossary term") {
		t.Errorf("unexpected /term response: %s", msg.Content)
	}

	// Verify term exists in DB
	terms, _ := store.ListGlossaryByProject(ctx, projectID)
	if len(terms) != 1 || terms[0].TargetTerm != "Foundation Establishment" {
		t.Fatalf("term not found in DB: %+v", terms)
	}

	// 3. Test /char
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "/char Tiêu Viêm -> Dược Lão: Sư phụ, tự xưng: Đồ nhi",
	})
	if err != nil {
		t.Fatalf("/char failed: %v", err)
	}
	if !strings.Contains(msg.Content, "Configured address relation") {
		t.Errorf("unexpected /char response: %s", msg.Content)
	}

	// 4. Test /steer with double quotes in argument (verifying BUG-01 fix)
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   `/steer Dịch câu "kiếm khí" sắc lạnh và u ám hơn`,
	})
	if err != nil {
		t.Fatalf("/steer failed with quotes: %v", err)
	}
	if !strings.Contains(msg.Content, "Updated translation configuration") {
		t.Errorf("unexpected /steer response: %s", msg.Content)
	}

	// 5. Test /soul
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "/soul strict_editor",
	})
	if err != nil {
		t.Fatalf("/soul failed: %v", err)
	}
	if !strings.Contains(msg.Content, "strict_editor") {
		t.Errorf("unexpected /soul response: %s", msg.Content)
	}

	// 6. Test /rollback
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "/rollback 2",
	})
	if err != nil {
		t.Fatalf("/rollback failed: %v", err)
	}
	if !strings.Contains(msg.Content, "Rolled back chapter") {
		t.Errorf("unexpected /rollback response: %s", msg.Content)
	}

	// 7. Test /pause, /resume, /abort
	for _, cmd := range []string{"/pause", "/resume", "/abort"} {
		msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
			ThreadID:  thread.ID,
			ProjectID: projectID,
			Content:   cmd,
		})
		if err != nil {
			t.Fatalf("%s failed: %v", cmd, err)
		}
		if msg.Content == "" {
			t.Errorf("empty %s response", cmd)
		}
	}

	// 8. Test /learn with quotes
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   `/learn Lưu ý đại từ xưng hô "sư huynh - muội muội"`,
	})
	if err != nil {
		t.Fatalf("/learn failed: %v", err)
	}
	if !strings.Contains(msg.Content, "Reflexion Engine") {
		t.Errorf("unexpected /learn response: %s", msg.Content)
	}

	// 9. Test /compact
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "/compact",
	})
	if err != nil {
		t.Fatalf("/compact failed: %v", err)
	}
	if !strings.Contains(msg.Content, "Successfully compacted") {
		t.Errorf("unexpected /compact response: %s", msg.Content)
	}

	// 10. Test /clear
	msg, err = engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "/clear",
	})
	if err != nil {
		t.Fatalf("/clear failed: %v", err)
	}
	if !strings.Contains(msg.Content, "All messages in this thread have been cleared") {
		t.Errorf("unexpected /clear response: %s", msg.Content)
	}

	// Verify thread is now empty
	msgs, _ := store.ListNovelClawMessages(ctx, thread.ID)
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after /clear, got %d", len(msgs))
	}
}

func TestEngine_ReActToolCallingLoop(t *testing.T) {
	store, projectID := setupTestStorage(t)
	defer store.Close()
	ctx := context.Background()

	thread, err := store.GetOrCreateMainThread(ctx, projectID)
	if err != nil {
		t.Fatalf("GetOrCreateMainThread failed: %v", err)
	}

	emitter := &mockStreamEmitter{}

	// Multi-turn mock LLM:
	// Turn 1: Thinks, and emits a tool call for app_switch_tab
	// Turn 2: Receives tool result, produces final message
	mockLLM := &mockEngineLLM{
		responses: []*llm.CompletionResponse{
			{
				Thought: "Người dùng muốn chuyển sang tab Đồ thị để quan sát quan hệ. Tôi sẽ gọi app_switch_tab.",
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_123",
						Type: "function",
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{
							Name:      "app_switch_tab",
							Arguments: `{"tab": "graph"}`,
						},
					},
				},
			},
			{
				Thought: "Công cụ đã thực thi thành công. Thông báo cho người dùng.",
				Content: "Tôi đã chuyển giao diện sang tab Đồ thị nhân vật cho bạn rồi nhé!",
			},
		},
	}

	engine := NewEngine(Config{
		Store:        store,
		Client:       mockLLM,
		ModelName:    "test-model",
		SkillsDir:    "../../skills/default",
		Emitter:      emitter,
		CompactLimit: 250000,
	})

	resp, err := engine.SendMessage(ctx, dtos.SendNovelClawMessageRequest{
		ThreadID:  thread.ID,
		ProjectID: projectID,
		Content:   "Chuyển sang tab đồ thị hộ tôi với",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if resp.Content != "Tôi đã chuyển giao diện sang tab Đồ thị nhân vật cho bạn rồi nhé!" {
		t.Errorf("unexpected reply content: %s", resp.Content)
	}
	if !strings.Contains(resp.ThinkingContent, "Tôi sẽ gọi app_switch_tab") && !strings.Contains(resp.ThinkingContent, "Thông báo cho người dùng") {
		t.Errorf("unexpected thinking content: %s", resp.ThinkingContent)
	}

	// Verify events were emitted: thinking, step_badge, action_card, message_created
	hasThinking := false
	hasStepBadge := false
	hasActionCard := false
	for _, ev := range emitter.events {
		if ev.eventType == "thinking" {
			hasThinking = true
		}
		if ev.eventType == "step_badge" {
			hasStepBadge = true
		}
		if ev.eventType == "action_card" || ev.eventType == "app:action_dispatched" {
			hasActionCard = true
		}
	}

	if !hasThinking {
		t.Errorf("missing thinking event")
	}
	if !hasStepBadge {
		t.Errorf("missing step_badge event")
	}
	if !hasActionCard {
		t.Errorf("missing action event")
	}
}
