package llmcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"novelclaw/pkg/llm"
)

var fencedJSONRegex = regexp.MustCompile("(?s)```(?:json)?\\s*([\\{\\[].*?[\\}\\]])\\s*```")

// ExtractJSON attempts to find and return the outermost JSON string from raw text.
func ExtractJSON(raw string) string {
	trimmed := strings.TrimSpace(raw)
	match := fencedJSONRegex.FindStringSubmatch(trimmed)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	start := strings.IndexAny(trimmed, "{[")
	if start == -1 {
		return trimmed
	}
	end := strings.LastIndexAny(trimmed, "}]")
	if end == -1 || end <= start {
		return trimmed
	}
	return strings.TrimSpace(trimmed[start : end+1])
}

// Execute enforces structured output contract with automatic self-healing feedback.
func Execute[T any](
	ctx context.Context,
	client llm.LLMClient,
	contract Contract[T],
	systemPrompt string,
	userPayload string,
	mode Mode,
	maxRetries int,
) (*T, error) {
	if maxRetries <= 0 {
		maxRetries = 2
	}

	preparedSystem, err := contract.PrepareSystemPrompt(systemPrompt, mode)
	if err != nil {
		return nil, err
	}

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: preparedSystem},
		{Role: llm.RoleUser, Content: userPayload},
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		req := llm.CompletionRequest{
			Messages:    messages,
			Temperature: 0.1,
		}

		resp, err := client.Generate(ctx, req)
		if err != nil {
			lastErr = fmt.Errorf("llm call failed: %w", err)
			continue
		}

		rawOutput := strings.TrimSpace(resp.Content)
		jsonCandidate := ExtractJSON(rawOutput)

		var target T
		if err := json.Unmarshal([]byte(jsonCandidate), &target); err != nil {
			lastErr = fmt.Errorf("invalid json format: %w", err)
			log.Warn().
				Str("contract", contract.Name).
				Int("attempt", attempt+1).
				Err(err).
				Msg("contract json parsing failed, requesting self-healing retry")

			correctionMsg := fmt.Sprintf("Your previous output failed JSON parsing: %v. Please correct and output ONLY a valid JSON object.", err)
			messages = append(messages,
				llm.Message{Role: llm.RoleAssistant, Content: rawOutput},
				llm.Message{Role: llm.RoleUser, Content: correctionMsg},
			)
			continue
		}

		if contract.Validate != nil {
			if err := contract.Validate(&target); err != nil {
				lastErr = fmt.Errorf("business validation failed: %w", err)
				log.Warn().
					Str("contract", contract.Name).
					Int("attempt", attempt+1).
					Err(err).
					Msg("contract business validation failed, requesting self-healing retry")

				correctionMsg := fmt.Sprintf("Your JSON output was structurally valid but failed business constraint validation: %v. Please fix the values and output ONLY the corrected JSON object.", err)
				messages = append(messages,
					llm.Message{Role: llm.RoleAssistant, Content: rawOutput},
					llm.Message{Role: llm.RoleUser, Content: correctionMsg},
				)
				continue
			}
		}

		return &target, nil
	}

	return nil, fmt.Errorf("contract %s execution failed after %d retries: %w", contract.Name, maxRetries, lastErr)
}
