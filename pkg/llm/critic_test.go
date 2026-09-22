package llm

import (
	"context"
	"testing"
	"time"
)

type mockCriticClient struct {
	response string
}

func (m *mockCriticClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{Content: m.response}, nil
}

func (m *mockCriticClient) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: m.response}
	close(ch)
	return ch, nil
}

func (m *mockCriticClient) ProviderName() string {
	return "mock"
}

// TestShadowCritic_ParsePass verifies that PASS output results in CriticActionPass.
func TestShadowCritic_ParsePass(t *testing.T) {
	client := &mockCriticClient{response: "PASS"}
	critic := NewShadowCritic(client, "test-model", 5*time.Second)

	verdict, err := critic.EvaluateSentence(context.Background(), CriticRequest{
		SourceLang:       "ja",
		TargetLang:       "vi",
		OriginalSentence: "伊月は机に向かった。",
		DraftTranslation: "Itsuki ngồi vào bàn.",
	})
	if err != nil {
		t.Fatalf("EvaluateSentence failed: %v", err)
	}

	if verdict.Action != CriticActionPass {
		t.Errorf("expected PASS, got %s", verdict.Action)
	}
}

// TestShadowCritic_ParseRevise verifies that REVISE with reason and correction parses properly.
func TestShadowCritic_ParseRevise(t *testing.T) {
	rawOutput := "REVISE\nREASON: Nayuta addresses Itsuki as Senior (Senpai) in this scene.\nCORRECTION: Nayuta bước vào và gọi, \"Tiền bối!\""
	client := &mockCriticClient{response: rawOutput}
	critic := NewShadowCritic(client, "test-model", 5*time.Second)

	verdict, err := critic.EvaluateSentence(context.Background(), CriticRequest{
		SourceLang:       "ja",
		TargetLang:       "vi",
		OriginalSentence: "那由多は「先輩！」と呼んだ。",
		DraftTranslation: "Nayuta bước vào và gọi, \"Anh yêu!\"",
	})
	if err != nil {
		t.Fatalf("EvaluateSentence failed: %v", err)
	}

	if verdict.Action != CriticActionRevise {
		t.Fatalf("expected REVISE, got %s", verdict.Action)
	}
	if verdict.Reason == "" {
		t.Errorf("expected reason to be parsed")
	}
	if verdict.Correction != "Nayuta bước vào và gọi, \"Tiền bối!\"" {
		t.Errorf("unexpected correction: %s", verdict.Correction)
	}
}

// TestShadowCritic_ParseReviseMultiline verifies that multiline corrections are fully preserved.
func TestShadowCritic_ParseReviseMultiline(t *testing.T) {
	rawOutput := "REVISE\nREASON: Multiple dialogue lines\nCORRECTION: Dòng thứ nhất.\nDòng thứ hai.\nDòng thứ ba."
	client := &mockCriticClient{response: rawOutput}
	critic := NewShadowCritic(client, "test-model", 5*time.Second)

	verdict, err := critic.EvaluateSentence(context.Background(), CriticRequest{
		SourceLang:       "ja",
		TargetLang:       "vi",
		OriginalSentence: "複数行のテスト。",
		DraftTranslation: "Bản dịch cũ.",
	})
	if err != nil {
		t.Fatalf("EvaluateSentence failed: %v", err)
	}

	if verdict.Action != CriticActionRevise {
		t.Fatalf("expected REVISE, got %s", verdict.Action)
	}
	expectedCorrection := "Dòng thứ nhất.\nDòng thứ hai.\nDòng thứ ba."
	if verdict.Correction != expectedCorrection {
		t.Errorf("expected multiline correction %q, got %q", expectedCorrection, verdict.Correction)
	}
}
