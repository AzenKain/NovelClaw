package llmcontract

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Mode represents how structured output schema is enforced.
type Mode string

const (
	ModeNativeJSON     Mode = "native_json"
	ModePromptContract Mode = "prompt_contract"
)

// Contract defines a strongly-typed structured output schema and validator.
type Contract[T any] struct {
	Name        string
	Description string
	Schema      map[string]any
	Validate    func(*T) error
}

// PrepareSystemPrompt appends the output schema contract to prompt if in prompt contract mode.
func (c *Contract[T]) PrepareSystemPrompt(basePrompt string, mode Mode) (string, error) {
	if mode == ModeNativeJSON {
		return basePrompt, nil
	}

	schemaJSON, err := json.Marshal(c.Schema)
	if err != nil {
		return "", fmt.Errorf("marshal contract schema %s: %w", c.Name, err)
	}

	instruction := fmt.Sprintf("\n\n## OUTPUT CONTRACT\nYou must output strictly valid JSON matching the following JSON Schema with no Markdown fences, preambles, or conversational commentary.\n<output_json_schema>\n%s\n</output_json_schema>", string(schemaJSON))
	return strings.TrimSpace(basePrompt) + instruction, nil
}
