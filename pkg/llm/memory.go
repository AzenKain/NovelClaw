package llm

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"novelclaw/pkg/storage"
)

// MemoryManager coordinates L0 sliding context and L1 episodic summaries.
type MemoryManager struct {
	store *storage.Storage
}

// NewMemoryManager creates a new MemoryManager.
func NewMemoryManager(store *storage.Storage) *MemoryManager {
	return &MemoryManager{store: store}
}

// ExtractL0SlidingContext extracts the trailing boundary sentences from previous translation.
func (m *MemoryManager) ExtractL0SlidingContext(previousTranslation string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = 300
	}
	clean := strings.TrimSpace(previousTranslation)
	if clean == "" {
		return ""
	}

	runes := []rune(clean)
	if len(runes) <= maxRunes {
		return clean
	}

	tail := string(runes[len(runes)-maxRunes:])
	sentenceDelims := []string{".\n", ".\"", "。」", "。", "！", "？", ". ", ".\n\n"}
	minIdx := -1
	for _, delim := range sentenceDelims {
		idx := strings.Index(tail, delim)
		if idx != -1 && (minIdx == -1 || idx < minIdx) {
			minIdx = idx + len(delim)
		}
	}

	if minIdx != -1 && minIdx < len(tail) {
		return strings.TrimSpace(tail[minIdx:])
	}
	return strings.TrimSpace(tail)
}

// GenerateAndSaveL1Summary generates a brief chapter plot summary and persists it.
func (m *MemoryManager) GenerateAndSaveL1Summary(ctx context.Context, client LLMClient, model, projectID string, chapterIndex int64, chapterTitle, fullTranslatedText string) (string, error) {
	textForSummary := fullTranslatedText
	if utf8.RuneCountInString(textForSummary) > 4000 {
		runes := []rune(textForSummary)
		textForSummary = string(runes[:4000]) + "..."
	}

	prompt := fmt.Sprintf(`Please read the translated chapter below and summarize the key plot developments, characters, and events in 3-4 concise sentences:
Title: %s
Content:
%s

Format: Return only the summary text without any preface or postscript.`, chapterTitle, textForSummary)

	req := CompletionRequest{
		Model: model,
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a professional novel editor summarizing key chapter plot developments."},
			{Role: RoleUser, Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   300,
	}

	resp, err := client.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("generate l1 summary: %w", err)
	}

	summary := strings.TrimSpace(resp.Content)
	if summary == "" {
		summary = fmt.Sprintf("Chapter %d: %s (Translation completed)", chapterIndex, chapterTitle)
	}

	err = m.store.UpsertChapterTimeline(ctx, projectID, chapterIndex, summary, nil)
	if err != nil {
		return summary, fmt.Errorf("save l1 timeline to db: %w", err)
	}

	return summary, nil
}
