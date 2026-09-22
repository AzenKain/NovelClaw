package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// KeyPoolManager manages a pool of API keys with round-robin rotation and cooldown.
type KeyPoolManager struct {
	mu           sync.RWMutex
	keys         []string
	currentIndex uint64
	disabledKeys map[string]time.Time
	maxRetries   int
}

// NewKeyPoolManager creates a new KeyPoolManager.
func NewKeyPoolManager(keys []string, maxRetries int) *KeyPoolManager {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	cleanKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		if trimmed := strings.TrimSpace(k); trimmed != "" {
			cleanKeys = append(cleanKeys, trimmed)
		}
	}
	return &KeyPoolManager{
		keys:         cleanKeys,
		disabledKeys: make(map[string]time.Time),
		maxRetries:   maxRetries,
	}
}

// NextKey returns the next available key skipping rate-limited ones.
func (p *KeyPoolManager) NextKey() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for k, exp := range p.disabledKeys {
		if now.After(exp) {
			delete(p.disabledKeys, k)
		}
	}

	if len(p.keys) == 0 {
		return "", errors.New("no api keys configured in key pool")
	}

	n := uint64(len(p.keys))
	for i := uint64(0); i < n; i++ {
		idx := (atomic.AddUint64(&p.currentIndex, 1) - 1) % n
		k := p.keys[idx]
		if _, disabled := p.disabledKeys[k]; !disabled {
			return k, nil
		}
	}

	return "", errors.New("all api keys are currently rate-limited (429), please wait")
}

// MarkRateLimited sets a cooldown for a rate-limited key.
func (p *KeyPoolManager) MarkRateLimited(key string, cooldown time.Duration) {
	if cooldown <= 0 {
		cooldown = 30 * time.Second
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.disabledKeys[key] = time.Now().Add(cooldown)
}

// IsRateLimitError checks whether an error indicates rate limiting.
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "status 429") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "quota exceeded") ||
		strings.Contains(msg, "too many requests")
}

// ExecuteWithRetry executes a call with exponential backoff on retryable errors.
func (p *KeyPoolManager) ExecuteWithRetry(ctx context.Context, fn func(apiKey string) (*CompletionResponse, error)) (*CompletionResponse, error) {
	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		key, err := p.NextKey()
		if err != nil {
			return nil, fmt.Errorf("key pool: %w", err)
		}

		resp, err := fn(key)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if IsRateLimitError(err) {
			p.MarkRateLimited(key, 30*time.Second)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
		}
	}

	return nil, fmt.Errorf("all %d retry attempts failed: %w", p.maxRetries, lastErr)
}
