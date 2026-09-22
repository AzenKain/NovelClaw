package llm_test

import (
	"context"
	"errors"
	"testing"

	"novelclaw/pkg/llm"
)

// MockClient mocks LLMClient for testing retry and failover logic.
type MockClient struct {
	providerName string
	failCount    int
	totalCalls   int
	response     string
}

// ProviderName returns mock provider name.
func (m *MockClient) ProviderName() string {
	return m.providerName
}

// Generate returns mock error or successful response.
func (m *MockClient) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	m.totalCalls++
	if m.failCount > 0 {
		m.failCount--
		return nil, errors.New("temporary 502 bad gateway")
	}
	return &llm.CompletionResponse{Content: m.response}, nil
}

// Stream returns mock stream channel.
func (m *MockClient) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: m.response}
	close(ch)
	return ch, nil
}

// TestFallbackRouter_AutoFailover tests automatic failover to the secondary route when primary fails.
func TestFallbackRouter_AutoFailover(t *testing.T) {
	client1 := &MockClient{
		providerName: "client_failing",
		failCount:    10,
		response:     "response from 1",
	}

	client2 := &MockClient{
		providerName: "client_backup",
		failCount:    0,
		response:     "response from backup",
	}

	routes := []llm.ModelRoute{
		{Name: "Primary", Client: client1, Model: "model-1", MaxRetries: 2},
		{Name: "Secondary", Client: client2, Model: "model-2", MaxRetries: 2},
	}

	router := llm.NewFallbackRouter(routes)
	resp, err := router.Generate(context.Background(), llm.CompletionRequest{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hello"}},
	})

	if err != nil {
		t.Fatalf("FallbackRouter failed: %v", err)
	}

	if resp.Content != "response from backup" {
		t.Errorf("expected response from backup, got %s", resp.Content)
	}

	if client1.totalCalls != 2 {
		t.Errorf("expected client1 to be called 2 times, got %d", client1.totalCalls)
	}
	if client2.totalCalls != 1 {
		t.Errorf("expected client2 to be called 1 time, got %d", client2.totalCalls)
	}
}
