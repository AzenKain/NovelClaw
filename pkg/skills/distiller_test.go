package skills

import (
	"context"
	"testing"

	"novelclaw/pkg/llm"
)

type mockDistillClient struct {
	response string
}

func (m *mockDistillClient) ProviderName() string {
	return "mock"
}

func (m *mockDistillClient) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{
		Content: m.response,
	}, nil
}

func (m *mockDistillClient) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: m.response, FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func TestDistiller_DistillDiff(t *testing.T) {
	ctx := context.Background()
	mockResp := `{
  "rules": [
    "Không dùng đại từ cổ phong 'ngươi' trong bối cảnh hiện đại"
  ],
  "few_shots": [
    {
      "input": "你今天怎么了？",
      "output": "Hôm nay cậu sao thế?",
      "note": "Bạn bè hiện đại"
    }
  ],
  "summary": "Đã chuẩn hóa đại từ xưng hô hiện đại.",
  "target_skill_id": "skill_literary_translator"
}`

	client := &mockDistillClient{response: mockResp}
	distiller := NewDistiller()

	res, err := distiller.DistillDiff(
		ctx,
		client,
		"mock-model",
		"你今天怎么了？",
		"Hôm nay ngươi sao thế?",
		"Hôm nay cậu sao thế?",
		"Đổi ngươi thành cậu cho thân mật",
	)
	if err != nil {
		t.Fatalf("DistillDiff failed: %v", err)
	}

	if len(res.Rules) != 1 || res.Rules[0] != "Không dùng đại từ cổ phong 'ngươi' trong bối cảnh hiện đại" {
		t.Errorf("unexpected rules: %v", res.Rules)
	}
	if len(res.FewShots) != 1 || res.FewShots[0].Output != "Hôm nay cậu sao thế?" {
		t.Errorf("unexpected few-shots: %v", res.FewShots)
	}
	if res.Summary != "Đã chuẩn hóa đại từ xưng hô hiện đại." {
		t.Errorf("unexpected summary: %s", res.Summary)
	}
}
