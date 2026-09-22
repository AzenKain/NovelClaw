package llm_test

import (
	"strings"
	"testing"

	"novelclaw/pkg/llm"
)

// TestChunkerPreservesQuotes ensures that chunks never split in the middle of dialogue quotes.
func TestChunkerPreservesQuotes(t *testing.T) {
	text := `Paragraph 1: A peaceful morning in the apartment of Itsuki Hashima.

Paragraph 2: He was engrossed in writing his new light novel manuscript.

Paragraph 3: Suddenly the door flew open, and a girl entered shouting:
「Hey Itsuki, are you staying up all night writing about little sisters again?
Don't you realize how important sleep is, you idiot!」

Paragraph 4: Itsuki sighed, swiveled his chair, and smiled:
"Little sisters are the ultimate truth of this world; I cannot stop!"

Paragraph 5: And thus ended another noisy, passionate morning for the young authors.`

	chunker := llm.NewChunker(120)
	chunks := chunker.ChunkText(text)

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	for i, c := range chunks {
		openCountJA := strings.Count(c.Content, "「")
		closeCountJA := strings.Count(c.Content, "」")
		if openCountJA != closeCountJA {
			t.Errorf("chunk %d split in the middle of Japanese quotes 「」: %s", i+1, c.Content)
		}

		quoteCount := strings.Count(c.Content, "\"")
		if quoteCount%2 != 0 {
			t.Errorf("chunk %d split in the middle of double quotes \"\": %s", i+1, c.Content)
		}
	}
}
