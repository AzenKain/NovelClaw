package llm_test

import (
	"testing"

	"novelclaw/pkg/llm"
)

// TestMemoryManager_ExtractL0SlidingContext verifies sliding context extraction from previous translations.
func TestMemoryManager_ExtractL0SlidingContext(t *testing.T) {
	mm := llm.NewMemoryManager(nil)

	prev := "First sentence. Second sentence. Third sentence. Fourth sentence concluding the section."
	ctx := mm.ExtractL0SlidingContext(prev, 50)

	if ctx == "" {
		t.Fatalf("expected non-empty context")
	}

	if len([]rune(ctx)) > 55 {
		t.Errorf("extracted context exceeds target length: %d", len([]rune(ctx)))
	}
}
