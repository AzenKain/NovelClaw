package llm

import (
	"strings"
	"time"
)

// TranslationMode specifies the orchestration strategy for translating chapters.
type TranslationMode string

const (
	ModeSinglePass          TranslationMode = "single_pass"
	ModeHierarchical3Pass   TranslationMode = "hierarchical_3pass"
	ModeConcurrentDualAgent TranslationMode = "concurrent_dual_agent"
	ModeSwarmArcParallel    TranslationMode = "swarm_arc_parallel"
)

// NormalizeTranslationMode maps user-facing / legacy mode aliases onto the
// canonical mode values. The frontend historically sent "swarm_arc" and
// "dual_agent" while the engine expects "swarm_arc_parallel" and
// "concurrent_dual_agent"; without this the mode was silently dropped and
// the run fell back to the default.
func NormalizeTranslationMode(raw string) TranslationMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "concurrent_dual_agent", "dual_agent", "concurrentdualagent", "dual":
		return ModeConcurrentDualAgent
	case "single_pass", "singlepass", "single":
		return ModeSinglePass
	case "hierarchical_3pass", "hierarchical3pass", "hierarchical", "3pass":
		return ModeHierarchical3Pass
	case "swarm_arc", "swarm_arc_parallel", "swarmarc", "swarm", "arc":
		return ModeSwarmArcParallel
	default:
		return TranslationMode(raw)
	}
}

// TranslationOptions configures execution parameters for a translation run.
type TranslationOptions struct {
	Mode              TranslationMode `json:"mode"`
	Concurrency       int             `json:"concurrency"`
	CriticModel       string          `json:"critic_model"`
	PolishModel       string          `json:"polish_model"`
	TimeoutPerChunk   time.Duration   `json:"timeout_per_chunk"`
	MaxRetries        int             `json:"max_retries"`
	EnableHotPatch    bool            `json:"enable_hot_patch"`
	EnableR19         bool            `json:"enable_r19"`
	EnableAgenticRAG  bool            `json:"enable_agentic_rag"`
	MaxToolIterations int             `json:"max_tool_iterations"`
	DecisionMode      string          `json:"decision_mode"` // "manual" | "auto"
	StyleGuide        string          `json:"style_guide,omitempty"`
	WorldBibleContext string          `json:"world_bible_context,omitempty"`
	VoiceContext      string          `json:"voice_context,omitempty"`
	ThinkingEffort    string          `json:"thinking_effort,omitempty"` // "off" | "low" | "medium" | "high" | "xhigh" | "max"
}

// DefaultTranslationOptions returns standard recommended translation options.
func DefaultTranslationOptions() TranslationOptions {
	return TranslationOptions{
		Mode:              ModeConcurrentDualAgent,
		Concurrency:       3,
		TimeoutPerChunk:   120 * time.Second,
		MaxRetries:        3,
		EnableHotPatch:    true,
		EnableAgenticRAG:  true,
		MaxToolIterations: 2,
		DecisionMode:      "manual",
		ThinkingEffort:    "off",
	}
}

// ExecutionTelemetry records runtime performance, token burn, and hot-patches.
type ExecutionTelemetry struct {
	Mode             TranslationMode `json:"mode"`
	Duration         time.Duration   `json:"duration"`
	PromptTokens     int             `json:"prompt_tokens"`
	CompTokens       int             `json:"comp_tokens"`
	CompletionTokens int             `json:"completion_tokens"`
	ReasoningTokens  int             `json:"reasoning_tokens,omitempty"`
	TotalTokens      int             `json:"total_tokens"`
	RevisedCount     int             `json:"revised_count"`
	ToolCallsCount   int             `json:"tool_calls_count"`
	TotalRunes       int             `json:"total_runes"`
	RunesPerSecond   float64         `json:"runes_per_second"`
	TokensPerSecond  float64         `json:"tokens_per_second"`
	EstimatedCostUSD float64         `json:"estimated_cost_usd"`
}

// EstimateTokenCostUSD calculates estimated USD cost based on the model family and token usage.
func EstimateTokenCostUSD(modelName string, promptTokens, completionTokens int) float64 {
	m := strings.ToLower(modelName)
	var promptRate, compRate float64 // per 1M tokens

	switch {
	case strings.Contains(m, "flash"):
		promptRate, compRate = 0.15, 0.60
	case strings.Contains(m, "pro") && strings.Contains(m, "gemini"):
		promptRate, compRate = 1.25, 5.00
	case strings.Contains(m, "gpt-4o-mini"):
		promptRate, compRate = 0.15, 0.60
	case strings.Contains(m, "gpt-4o"):
		promptRate, compRate = 2.50, 10.00
	case strings.Contains(m, "sonnet") || strings.Contains(m, "claude-3-5"):
		promptRate, compRate = 3.00, 15.00
	case strings.Contains(m, "haiku"):
		promptRate, compRate = 0.80, 4.00
	case strings.Contains(m, "deepseek"):
		promptRate, compRate = 0.14, 0.28
	case strings.Contains(m, "ollama") || strings.Contains(m, "local"):
		promptRate, compRate = 0.0, 0.0
	default:
		promptRate, compRate = 0.15, 0.60
	}

	return (float64(promptTokens)*promptRate + float64(completionTokens)*compRate) / 1000000.0
}
