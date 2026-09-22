package dtos

import (
	"time"

	"novelclaw/pkg/benchmark"
	"novelclaw/pkg/llm"
)

// TranslationOptionsDTO defines frontend options for selecting translation modes.
type TranslationOptionsDTO struct {
	Mode              string `json:"mode"`
	Concurrency       int    `json:"concurrency"`
	CriticModel       string `json:"critic_model"`
	PolishModel       string `json:"polish_model"`
	TimeoutSeconds    int64  `json:"timeout_seconds"`
	MaxRetries        int    `json:"max_retries"`
	EnableHotPatch    bool   `json:"enable_hot_patch"`
	EnableR19         bool   `json:"enable_r19"`
	EnableAgenticRAG  bool   `json:"enable_agentic_rag"`
	MaxToolIterations int    `json:"max_tool_iterations"`
	DecisionMode      string `json:"decision_mode"`
	StyleGuide        string `json:"style_guide,omitempty"`
	ThinkingEffort    string `json:"thinking_effort,omitempty"`
}

// ExecutionTelemetryDTO represents quantitative telemetry after translation completes.
type ExecutionTelemetryDTO struct {
	Mode             string  `json:"mode"`
	DurationMs       int64   `json:"duration_ms"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	ReasoningTokens  int     `json:"reasoning_tokens"`
	RevisedCount     int     `json:"revised_count"`
	ToolCallsCount   int     `json:"tool_calls_count"`
	TotalRunes       int     `json:"total_runes"`
	RunesPerSecond   float64 `json:"runes_per_second"`
	TokensPerSecond  float64 `json:"tokens_per_second"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// BenchmarkMetricsDTO represents performance data for a specific translation mode.
type BenchmarkMetricsDTO struct {
	Mode                    string  `json:"mode"`
	ChaptersCount           int     `json:"chapters_count"`
	TotalRunes              int     `json:"total_runes"`
	DurationMs              int64   `json:"duration_ms"`
	RunesPerSecond          float64 `json:"runes_per_second"`
	TokensPerSecond         float64 `json:"tokens_per_second"`
	PromptTokens            int     `json:"prompt_tokens"`
	CompletionTokens        int     `json:"completion_tokens"`
	TotalTokens             int     `json:"total_tokens"`
	ReasoningTokens         int     `json:"reasoning_tokens"`
	EstimatedCostUSD        float64 `json:"estimated_cost_usd"`
	PronounConsistencyScore float64 `json:"pronoun_consistency_score"`
	RevisedCount            int     `json:"revised_count"`
}

// BenchmarkReportDTO summarizes comparative benchmarks across modes.
type BenchmarkReportDTO struct {
	ProjectID      string                `json:"project_id"`
	Timestamp      string                `json:"timestamp"`
	Results        []BenchmarkMetricsDTO `json:"results"`
	MarkdownReport string                `json:"markdown_report"`
}

// StartTranslationRequest contains parameters to begin translating a chapter.
type StartTranslationRequest struct {
	ProjectID    string                `json:"project_id"`
	ChapterIndex int64                 `json:"chapter_index"`
	Options      TranslationOptionsDTO `json:"options"`
}

// StartBenchmarkRequest contains parameters to run a comparative benchmark.
type StartBenchmarkRequest struct {
	ProjectID      string   `json:"project_id"`
	ChapterIndices []int64  `json:"chapter_indices"`
	Modes          []string `json:"modes"`
}

// ToDomainOptions converts TranslationOptionsDTO to llm.TranslationOptions.
func (d TranslationOptionsDTO) ToDomainOptions() llm.TranslationOptions {
	timeout := time.Duration(d.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	mode := llm.NormalizeTranslationMode(d.Mode)
	maxIterations := d.MaxToolIterations
	if maxIterations <= 0 {
		maxIterations = 2
	}
	decMode := d.DecisionMode
	if decMode == "" {
		decMode = "manual"
	}
	effort := d.ThinkingEffort
	if effort == "" {
		effort = "off"
	}
	return llm.TranslationOptions{
		Mode:              mode,
		Concurrency:       d.Concurrency,
		CriticModel:       d.CriticModel,
		PolishModel:       d.PolishModel,
		TimeoutPerChunk:   timeout,
		MaxRetries:        d.MaxRetries,
		EnableHotPatch:    d.EnableHotPatch,
		EnableR19:         d.EnableR19,
		EnableAgenticRAG:  d.EnableAgenticRAG,
		MaxToolIterations: maxIterations,
		DecisionMode:      decMode,
		StyleGuide:        d.StyleGuide,
		ThinkingEffort:    effort,
	}
}

// ToTelemetryDTO converts llm.ExecutionTelemetry to ExecutionTelemetryDTO.
func ToTelemetryDTO(t *llm.ExecutionTelemetry) *ExecutionTelemetryDTO {
	if t == nil {
		return nil
	}
	return &ExecutionTelemetryDTO{
		Mode:             string(t.Mode),
		DurationMs:       t.Duration.Milliseconds(),
		PromptTokens:     t.PromptTokens,
		CompletionTokens: t.CompletionTokens,
		TotalTokens:      t.TotalTokens,
		ReasoningTokens:  t.ReasoningTokens,
		RevisedCount:     t.RevisedCount,
		ToolCallsCount:   t.ToolCallsCount,
		TotalRunes:       t.TotalRunes,
		RunesPerSecond:   t.RunesPerSecond,
		TokensPerSecond:  t.TokensPerSecond,
		EstimatedCostUSD: t.EstimatedCostUSD,
	}
}

// ToBenchmarkReportDTO converts benchmark.BenchmarkReport to BenchmarkReportDTO.
func ToBenchmarkReportDTO(r *benchmark.BenchmarkReport) *BenchmarkReportDTO {
	if r == nil {
		return nil
	}
	results := make([]BenchmarkMetricsDTO, 0, len(r.Results))
	for _, m := range r.Results {
		results = append(results, BenchmarkMetricsDTO{
			Mode:                    string(m.Mode),
			ChaptersCount:           m.ChaptersCount,
			TotalRunes:              m.TotalRunes,
			DurationMs:              m.Duration.Milliseconds(),
			RunesPerSecond:          m.RunesPerSecond,
			PromptTokens:            m.PromptTokens,
			CompletionTokens:        m.CompletionTokens,
			TotalTokens:             m.TotalTokens,
			EstimatedCostUSD:        m.EstimatedCostUSD,
			PronounConsistencyScore: m.PronounConsistencyScore,
			RevisedCount:            m.RevisedCount,
		})
	}
	return &BenchmarkReportDTO{
		ProjectID:      r.ProjectID,
		Timestamp:      r.Timestamp.Format(time.RFC3339),
		Results:        results,
		MarkdownReport: r.GenerateMarkdown(),
	}
}
