package benchmark

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/jsonx"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

// BenchmarkReport summarizes the multi-mode comparative evaluation.
type BenchmarkReport struct {
	ProjectID string             `json:"project_id"`
	Timestamp time.Time          `json:"timestamp"`
	Results   []BenchmarkMetrics `json:"results"`
}

// GenerateMarkdown creates a structured Markdown comparison table from benchmark results.
func (r *BenchmarkReport) GenerateMarkdown() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Benchmark Evaluation Report: Project %s\n", r.ProjectID))
	sb.WriteString(fmt.Sprintf("*Generated at: %s*\n\n", r.Timestamp.Format(time.RFC3339)))
	sb.WriteString("| Mode | Chapters | Duration | Speed (runes/s) | Tokens (In/Out/Total) | Revisions | Pronoun Consistency | Cost (USD) |\n")
	sb.WriteString("| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |\n")

	for _, m := range r.Results {
		tokensStr := fmt.Sprintf("%d / %d / %d", m.PromptTokens, m.CompletionTokens, m.TotalTokens)
		sb.WriteString(fmt.Sprintf("| **%s** | %d | %v | %.1f | %s | %d | %.1f%% | $%.4f |\n",
			m.Mode,
			m.ChaptersCount,
			m.Duration.Round(time.Millisecond),
			m.RunesPerSecond,
			tokensStr,
			m.RevisedCount,
			m.PronounConsistencyScore*100.0,
			m.EstimatedCostUSD,
		))
	}

	return sb.String()
}

// GenerateJSON serializes the benchmark report to JSON.
func (r *BenchmarkReport) GenerateJSON() (string, error) {
	b, err := jsonx.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Runner coordinates benchmark execution across translation modes.
type Runner struct {
	store  *storage.Storage
	client llm.LLMClient
	model  string
}

// NewRunner creates a new benchmark Runner.
func NewRunner(store *storage.Storage, client llm.LLMClient, model string) *Runner {
	return &Runner{
		store:  store,
		client: client,
		model:  model,
	}
}

// RunBenchmark executes comparative translations on specified chapters across selected modes.
func (r *Runner) RunBenchmark(
	ctx context.Context,
	projectID string,
	chapterIndices []int64,
	modes []llm.TranslationMode,
) (*BenchmarkReport, error) {
	if len(modes) == 0 {
		modes = []llm.TranslationMode{
			llm.ModeSinglePass,
			llm.ModeHierarchical3Pass,
			llm.ModeConcurrentDualAgent,
		}
	}

	report := &BenchmarkReport{
		ProjectID: projectID,
		Timestamp: time.Now(),
		Results:   make([]BenchmarkMetrics, 0, len(modes)),
	}

	var sourceBuilder strings.Builder
	for _, cIdx := range chapterIndices {
		if chap, err := r.store.GetChapterByIndex(ctx, projectID, cIdx); err == nil {
			sourceBuilder.WriteString(chap.RawContent)
			sourceBuilder.WriteString(" ")
		}
	}
	combinedSource := sourceBuilder.String()

	relations, _ := r.store.ListRelationsByProject(ctx, projectID)
	var expectedTerms []string
	for _, rel := range relations {
		if rel.CallAs != "" {
			if combinedSource == "" || strings.Contains(combinedSource, rel.FromChar) || strings.Contains(combinedSource, rel.ToChar) {
				expectedTerms = append(expectedTerms, rel.CallAs)
			}
		}
	}

	orch := llm.NewTranslationOrchestrator(r.client, r.model, r.store, nil)
	swarm := llm.NewSwarmArcOrchestrator(orch)

	for _, mode := range modes {
		log.Info().
			Str("mode", string(mode)).
			Int("chapters", len(chapterIndices)).
			Msg("starting benchmark run for mode")

		modeStart := time.Now()
		var combinedText strings.Builder
		totalPromptTokens := 0
		totalCompTokens := 0
		totalTokens := 0
		totalRevisions := 0
		totalRunes := 0

		opts := llm.DefaultTranslationOptions()
		opts.Mode = mode

		if mode == llm.ModeSwarmArcParallel {
			results, err := swarm.TranslateArcChapters(ctx, projectID, chapterIndices, opts, nil)
			if err != nil {
				return nil, fmt.Errorf("swarm arc parallel benchmark failed: %w", err)
			}
			for _, res := range results {
				combinedText.WriteString(res.Translated)
				combinedText.WriteString("\n\n")
				if res.Telemetry != nil {
					totalPromptTokens += res.Telemetry.PromptTokens
					totalCompTokens += res.Telemetry.CompletionTokens
					totalTokens += res.Telemetry.TotalTokens
					totalRevisions += res.Telemetry.RevisedCount
					totalRunes += res.Telemetry.TotalRunes
				}
			}
		} else {
			for _, cIdx := range chapterIndices {
				translated, tele, err := orch.TranslateChapter(ctx, projectID, cIdx, opts, nil)
				if err != nil {
					return nil, fmt.Errorf("benchmark mode %s chapter %d failed: %w", mode, cIdx, err)
				}
				combinedText.WriteString(translated)
				combinedText.WriteString("\n\n")
				if tele != nil {
					totalPromptTokens += tele.PromptTokens
					totalCompTokens += tele.CompletionTokens
					totalTokens += tele.TotalTokens
					totalRevisions += tele.RevisedCount
					totalRunes += tele.TotalRunes
				}
			}
		}

		duration := time.Since(modeStart)
		runesPerSec := 0.0
		if duration.Seconds() > 0 {
			runesPerSec = float64(totalRunes) / duration.Seconds()
		}

		pronounScore := CalculatePronounConsistency(combinedText.String(), expectedTerms)
		costUSD := llm.EstimateTokenCostUSD(r.model, totalPromptTokens, totalCompTokens)

		report.Results = append(report.Results, BenchmarkMetrics{
			Mode:                    mode,
			ChaptersCount:           len(chapterIndices),
			TotalRunes:              totalRunes,
			Duration:                duration,
			RunesPerSecond:          runesPerSec,
			PromptTokens:            totalPromptTokens,
			CompletionTokens:        totalCompTokens,
			TotalTokens:             totalTokens,
			EstimatedCostUSD:        costUSD,
			PronounConsistencyScore: pronounScore,
			RevisedCount:            totalRevisions,
		})
	}

	return report, nil
}
