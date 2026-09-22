package benchmark

import (
	"strings"
	"time"

	"novelclaw/pkg/llm"
)

// BenchmarkMetrics records quantitative performance and cost metrics for a translation mode.
type BenchmarkMetrics struct {
	Mode                    llm.TranslationMode `json:"mode"`
	ChaptersCount           int                 `json:"chapters_count"`
	TotalRunes              int                 `json:"total_runes"`
	Duration                time.Duration       `json:"duration"`
	RunesPerSecond          float64             `json:"runes_per_second"`
	PromptTokens            int                 `json:"prompt_tokens"`
	CompletionTokens        int                 `json:"completion_tokens"`
	TotalTokens             int                 `json:"total_tokens"`
	EstimatedCostUSD        float64             `json:"estimated_cost_usd"`
	PronounConsistencyScore float64             `json:"pronoun_consistency_score"`
	RevisedCount            int                 `json:"revised_count"`
}

// CalculatePronounConsistency evaluates the proportion of required address terms correctly used.
func CalculatePronounConsistency(translatedText string, requiredTerms []string) float64 {
	if len(requiredTerms) == 0 {
		return 1.0
	}

	lowerText := strings.ToLower(translatedText)
	matched := 0
	for _, term := range requiredTerms {
		if strings.Contains(lowerText, strings.ToLower(term)) {
			matched++
		}
	}

	return float64(matched) / float64(len(requiredTerms))
}
