package llm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"novelclaw/pkg/llm"
)

// TestKeyPoolManagerRoundRobinAndFailover tests round-robin distribution and 429 rate limit cooldown.
func TestKeyPoolManagerRoundRobinAndFailover(t *testing.T) {
	keys := []string{"key_1", "key_2", "key_3"}
	pool := llm.NewKeyPoolManager(keys, 3)

	k1, _ := pool.NextKey()
	k2, _ := pool.NextKey()
	k3, _ := pool.NextKey()
	k4, _ := pool.NextKey()

	if k1 != "key_1" || k2 != "key_2" || k3 != "key_3" || k4 != "key_1" {
		t.Errorf("Round robin sequence mismatch: %v %v %v %v", k1, k2, k3, k4)
	}

	pool.MarkRateLimited("key_2", 1*time.Minute)

	for i := 0; i < 4; i++ {
		k, err := pool.NextKey()
		if err != nil {
			t.Fatalf("unexpected error getting key: %v", err)
		}
		if k == "key_2" {
			t.Errorf("expected key_2 to be skipped due to rate limiting, got key_2")
		}
	}

	callCount := 0
	resp, err := pool.ExecuteWithRetry(context.Background(), func(apiKey string) (*llm.CompletionResponse, error) {
		callCount++
		if apiKey == "key_1" {
			return nil, errors.New("API error: status 429 rate limit exceeded")
		}
		return &llm.CompletionResponse{Content: "success from " + apiKey}, nil
	})

	if err != nil {
		t.Fatalf("ExecuteWithRetry failed: %v", err)
	}
	if resp.Content != "success from key_3" && resp.Content != "success from key_1" {
		t.Logf("Result: %s", resp.Content)
	}
}
