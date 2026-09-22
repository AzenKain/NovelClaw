package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"novelclaw/pkg/jsonx"
	"novelclaw/pkg/llm"
)

// DistillationResult holds extracted rules and examples inferred from user edits.
type DistillationResult struct {
	Rules         []string         `json:"rules"`
	FewShots      []FewShotExample `json:"few_shots"`
	Summary       string           `json:"summary"`
	TargetSkillID string           `json:"target_skill_id"` // defaults to "skill_literary_translator"
}

// Distiller analyzes diffs between LLM draft and user edited text to extract translation patterns.
type Distiller struct{}

// NewDistiller creates a new Distiller.
func NewDistiller() *Distiller {
	return &Distiller{}
}

// DistillDiff calls LLM to analyze the diff between original draft and user-edited version.
func (d *Distiller) DistillDiff(
	ctx context.Context,
	client llm.LLMClient,
	modelName string,
	rawSource string,
	originalDraft string,
	userEdited string,
	userNotes string,
) (*DistillationResult, error) {
	if client == nil {
		return nil, fmt.Errorf("llm client is required for distillation")
	}

	if strings.TrimSpace(originalDraft) == "" || strings.TrimSpace(userEdited) == "" {
		return nil, fmt.Errorf("both original draft and user edited text are required")
	}

	systemPrompt := `You are a senior Reflexion Distiller & Cognitive Engineer within a novel translation system.
Your task is to analyze the DIFFERENCES between:
1. The AI's original draft translation (Original Draft)
2. The human translator's refined edit (User Edited)
(Along with the raw source and the translator's notes, when provided.)

From the human's edits, distill:
1. "rules": An array of prohibited patterns or mandatory translation guidelines (e.g. "Do not use archaic pronouns in modern friendship contexts; prefer natural colloquial address instead").
2. "few_shots": A list of canonical paired examples (each with "input", "output", "note").
   - input: The source sentence or passage.
   - output: The canonical translation refined by the translator.
   - note: The rationale for choosing this rendering.
3. "summary": A 1-2 sentence summary of the core lesson learned from this editing pass.
4. "target_skill_id": The skill best suited to receive this rule (usually "skill_literary_translator" or "skill_foreign_sanitizer").

You MUST respond with pure JSON (RFC 8259): no markdown code fences, no explanation outside the JSON:
{
  "rules": [
    "Do not use archaic pronouns in modern school settings",
    "Preserve established domain terminology instead of overly literal translations"
  ],
  "few_shots": [
    {
      "input": "How are you doing today?",
      "output": "Refined canonical translation...",
      "note": "Natural conversational tone between peers"
    }
  ],
  "summary": "Standardized modern address pronouns while preserving domain terminology.",
  "target_skill_id": "skill_literary_translator"
}`

	var userPrompt strings.Builder
	if rawSource != "" {
		userPrompt.WriteString("--- RAW SOURCE ---\n")
		if len(rawSource) > 3000 {
			userPrompt.WriteString(rawSource[:3000])
			userPrompt.WriteString("...")
		} else {
			userPrompt.WriteString(rawSource)
		}
		userPrompt.WriteString("\n\n")
	}

	userPrompt.WriteString("--- AI ORIGINAL DRAFT ---\n")
	if len(originalDraft) > 4000 {
		userPrompt.WriteString(originalDraft[:4000])
		userPrompt.WriteString("...")
	} else {
		userPrompt.WriteString(originalDraft)
	}
	userPrompt.WriteString("\n\n")

	userPrompt.WriteString("--- TRANSLATOR EDIT (USER EDITED) ---\n")
	if len(userEdited) > 4000 {
		userPrompt.WriteString(userEdited[:4000])
		userPrompt.WriteString("...")
	} else {
		userPrompt.WriteString(userEdited)
	}
	userPrompt.WriteString("\n\n")

	if userNotes != "" {
		userPrompt.WriteString(fmt.Sprintf("--- ADDITIONAL TRANSLATOR NOTES ---\n%s\n\n", userNotes))
	}

	userPrompt.WriteString("Distill the rules and canonical examples into JSON.")

	resp, err := client.Generate(ctx, llm.CompletionRequest{
		Model: modelName,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: userPrompt.String()},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("call llm for distillation: %w", err)
	}

	cleanJSON := jsonx.ExtractJSONObject(resp.Content)

	var result DistillationResult
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse distillation response JSON: %w (raw: %s)", err, resp.Content)
	}

	if result.TargetSkillID == "" {
		result.TargetSkillID = "skill_literary_translator"
	}

	return &result, nil
}
