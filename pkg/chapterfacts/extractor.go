package chapterfacts

import (
	"context"
	"fmt"
	"strings"

	"novelclaw/pkg/llm"
	"novelclaw/pkg/llmcontract"
)

// FactContractSchema defines the expected JSON Schema for chapter fact extraction.
var FactContractSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"title": map[string]any{"type": "string"},
		"summary": map[string]any{"type": "string"},
		"key_events": map[string]any{
			"type": "array",
			"items": map[string]any{"type": "string"},
		},
		"relationship_changes": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from_char":    map[string]any{"type": "string"},
					"to_char":      map[string]any{"type": "string"},
					"call_as":      map[string]any{"type": "string"},
					"self_call_as": map[string]any{"type": "string"},
					"tone":         map[string]any{"type": "string"},
				},
				"required": []string{"from_char", "to_char", "call_as"},
			},
		},
		"state_changes": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"entity":    map[string]any{"type": "string"},
					"field":     map[string]any{"type": "string"},
					"old_value": map[string]any{"type": "string"},
					"new_value": map[string]any{"type": "string"},
					"reason":    map[string]any{"type": "string"},
				},
				"required": []string{"entity", "field", "new_value"},
			},
		},
		"cast_intros": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":   map[string]any{"type": "string"},
					"role":   map[string]any{"type": "string"},
					"gender": map[string]any{"type": "string"},
				},
				"required": []string{"name", "role"},
			},
		},
	},
	"required": []string{"title", "summary", "key_events"},
}

// Extractor handles extracting structured episodic facts from translated chapters.
type Extractor struct {
	contract llmcontract.Contract[ChapterFacts]
}

// NewExtractor creates an initialized Extractor.
func NewExtractor() *Extractor {
	return &Extractor{
		contract: llmcontract.Contract[ChapterFacts]{
			Name:        "chapter_facts",
			Description: "Structured episodic memory and dynamic relationship change extraction",
			Schema:      FactContractSchema,
			Validate: func(f *ChapterFacts) error {
				if strings.TrimSpace(f.Summary) == "" {
					return fmt.Errorf("chapter summary must not be empty")
				}
				return nil
			},
		},
	}
}

// Extract parses a chapter and extracts structured facts using the LLM contract.
func (e *Extractor) Extract(ctx context.Context, client llm.LLMClient, title, content string) (*ChapterFacts, error) {
	systemPrompt := "You are a professional literary continuity editor. Read the translated chapter and extract episodic facts: summary, key events, character relationship changes (how characters address each other), entity state changes, and newly introduced characters."
	userPayload := fmt.Sprintf("Title: %s\n\nChapter Content:\n%s", title, content)

	return llmcontract.Execute(ctx, client, e.contract, systemPrompt, userPayload, llmcontract.ModePromptContract, 2)
}
