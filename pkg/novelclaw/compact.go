package novelclaw

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

// DefaultAutoCompactThreshold is the default context token threshold (250k tokens)
const DefaultAutoCompactThreshold int64 = 250000

// Compactor manages context window compression and markdown distillation.
type Compactor struct {
	store     *storage.Storage
	client    llm.LLMClient
	modelName string
	threshold int64
}

// NewCompactor creates a new Compactor instance.
func NewCompactor(store *storage.Storage, client llm.LLMClient, modelName string, threshold int64) *Compactor {
	if threshold <= 0 {
		threshold = DefaultAutoCompactThreshold
	}
	return &Compactor{
		store:     store,
		client:    client,
		modelName: modelName,
		threshold: threshold,
	}
}

// SetThreshold updates the token threshold for auto-compact.
func (c *Compactor) SetThreshold(threshold int64) {
	if threshold > 0 {
		c.threshold = threshold
	}
}

// GetThreshold returns the current token threshold.
func (c *Compactor) GetThreshold() int64 {
	return c.threshold
}

// NeedsCompact checks if a thread's active tokens meet or exceed the threshold.
func (c *Compactor) NeedsCompact(ctx context.Context, threadID string) (bool, int64, error) {
	tokens, err := c.store.SumActiveThreadTokens(ctx, threadID)
	if err != nil {
		return false, 0, err
	}
	return tokens >= c.threshold, tokens, nil
}

// CompactThread compresses older active messages in a thread into a distilled summary.
func (c *Compactor) CompactThread(ctx context.Context, threadID string, force bool) (*dtos.CompactThreadResponse, error) {
	currentTokens, err := c.store.SumActiveThreadTokens(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("sum thread tokens: %w", err)
	}

	if !force && currentTokens < c.threshold {
		return &dtos.CompactThreadResponse{
			ThreadID:         threadID,
			OriginalTokens:   currentTokens,
			CompactedTokens:  currentTokens,
			ArchivedMessages: 0,
			SummaryNotice:    "Auto-Compact threshold not reached yet.",
		}, nil
	}

	activeMsgs, err := c.store.ListActiveNovelClawMessages(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("list active messages: %w", err)
	}

	if len(activeMsgs) <= 1 {
		return &dtos.CompactThreadResponse{
			ThreadID:         threadID,
			OriginalTokens:   currentTokens,
			CompactedTokens:  currentTokens,
			ArchivedMessages: 0,
			SummaryNotice:    "Conversation history is too short; no compaction needed.",
		}, nil
	}

	projectID := activeMsgs[0].ProjectID

	// Build conversation transcript for LLM distillation
	var transcript strings.Builder
	for _, m := range activeMsgs {
		transcript.WriteString(fmt.Sprintf("[%s - %s]: %s\n", m.Sender, m.Role, m.Content))
		if m.ActionCallJson.Valid && m.ActionCallJson.String != "" {
			transcript.WriteString(fmt.Sprintf("  -> Action: %s\n", m.ActionCallJson.String))
		}
	}

	// Request distillation summary from LLM
	summaryContent := ""
	if c.client != nil {
		systemPrompt := `You are the intelligent Context Distillation Compactor of NovelClaw Studio.
Your task is to distill the entire conversation history into a concise, clearly structured summary document.
The summary MUST preserve, in full:
1. Pronoun/address agreements and character relationships settled per Volume / Chapter.
2. L3 translation glossary terms and agreed translation conventions.
3. World entities, rules, and world-building settings (World Bible / Lorebook).
4. Style Guide directives and feedback from the author/translator.
5. Current translation progress and the next tasks.

Return the summary in Markdown format starting with '# [NOVELCLAW DISTILLED MEMORY CHUNK]'.`

		resp, err := c.client.Generate(ctx, llm.CompletionRequest{
			Model: c.modelName,
			Messages: []llm.Message{
				{Role: llm.RoleSystem, Content: systemPrompt},
				{Role: llm.RoleUser, Content: fmt.Sprintf("Here is the conversation history to compact:\n\n%s", transcript.String())},
			},
			Temperature: 0.2,
		})
		if err == nil && resp.Content != "" {
			summaryContent = resp.Content
		}
	}

	// Fallback rule-based distillation if LLM client is unavailable
	if summaryContent == "" {
		summaryContent = fmt.Sprintf("# [NOVELCLAW DISTILLED MEMORY CHUNK]\n- Automatic summary: archived %d previous messages.\n- Total tokens compacted: %d tokens.",
			len(activeMsgs), currentTokens)
	}

	// 1. Archive all current active messages
	if err := c.store.ArchiveThreadMessagesForCompact(ctx, threadID); err != nil {
		return nil, fmt.Errorf("archive compacted messages: %w", err)
	}

	// 2. Insert distilled memory chunk message
	summaryMsgID := "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	estimatedTokens := int64(len(summaryContent) / 4)
	_, err = c.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
		ID:                summaryMsgID,
		ThreadID:          threadID,
		ProjectID:         projectID,
		Sender:            "novelclaw",
		Role:              "system",
		Content:           summaryContent,
		ThinkingContent:   sql.NullString{Valid: false},
		StepType:          sql.NullString{String: "compact_summary", Valid: true},
		StepStatus:        sql.NullString{String: "completed", Valid: true},
		ActionCallJson:    sql.NullString{Valid: false},
		IsCollapsed:       0,
		IsArchivedCompact: 0,
		TokenCount:        estimatedTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("save compaction summary: %w", err)
	}

	// 3. Insert notification notice message
	noticeMsgID := "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	noticeText := fmt.Sprintf("⚡ NovelClaw auto-compacted the context (freed ~%d tokens, archived %d messages).",
		currentTokens-estimatedTokens, len(activeMsgs))
	_, _ = c.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
		ID:                noticeMsgID,
		ThreadID:          threadID,
		ProjectID:         projectID,
		Sender:            "system",
		Role:              "system",
		Content:           noticeText,
		ThinkingContent:   sql.NullString{Valid: false},
		StepType:          sql.NullString{String: "compact_notice", Valid: true},
		StepStatus:        sql.NullString{String: "completed", Valid: true},
		ActionCallJson:    sql.NullString{Valid: false},
		IsCollapsed:       0,
		IsArchivedCompact: 0,
		TokenCount:        10,
	})

	return &dtos.CompactThreadResponse{
		ThreadID:         threadID,
		OriginalTokens:   currentTokens,
		CompactedTokens:  estimatedTokens + 10,
		ArchivedMessages: len(activeMsgs),
		SummaryNotice:    noticeText,
	}, nil
}
