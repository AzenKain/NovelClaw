package llmcontract

import (
	"context"
	"errors"
	"strings"
	"testing"

	"novelclaw/pkg/llm"
)

type sequenceMockClient struct {
	responses []string
	callCount int
}

func (m *sequenceMockClient) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.callCount >= len(m.responses) {
		return nil, errors.New("out of responses")
	}
	content := m.responses[m.callCount]
	m.callCount++
	return &llm.CompletionResponse{Content: content}, nil
}

func (m *sequenceMockClient) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	close(ch)
	return ch, nil
}

func (m *sequenceMockClient) ProviderName() string {
	return "mock_sequence"
}

type samplePerson struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestExecuteSuccessDirect(t *testing.T) {
	mock := &sequenceMockClient{
		responses: []string{`{"name": "Alice", "age": 25}`},
	}

	contract := Contract[samplePerson]{
		Name: "sample_person",
		Schema: map[string]any{
			"type": "object",
		},
	}

	res, err := Execute(context.Background(), mock, contract, "sys", "user", ModePromptContract, 2)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}

	if res.Name != "Alice" || res.Age != 25 {
		t.Errorf("unexpected parsed result: %+v", res)
	}
}

func TestExecuteFencedAndSelfHealing(t *testing.T) {
	mock := &sequenceMockClient{
		responses: []string{
			`Here is your JSON: {"name": "", "age": 20}`,
			"```json\n{\"name\": \"Bob\", \"age\": 30}\n```",
		},
	}

	contract := Contract[samplePerson]{
		Name: "sample_person_validation",
		Validate: func(p *samplePerson) error {
			if strings.TrimSpace(p.Name) == "" {
				return errors.New("name cannot be empty")
			}
			return nil
		},
	}

	res, err := Execute(context.Background(), mock, contract, "sys", "user", ModePromptContract, 2)
	if err != nil {
		t.Fatalf("expected self-healing success, got: %v", err)
	}

	if res.Name != "Bob" || res.Age != 30 {
		t.Errorf("unexpected healed result: %+v", res)
	}

	if mock.callCount != 2 {
		t.Errorf("expected 2 calls for healing, got: %d", mock.callCount)
	}
}
