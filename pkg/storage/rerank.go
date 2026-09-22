package storage

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"novelclaw/internal/gen/sqlc"
)

// ScoredChapter represents a chapter result with multi-factor reranking scores.
type ScoredChapter struct {
	Chapter    sqlc.ChaptersFt
	FinalScore float64
}

// ScoredSnippet represents a snippet result with multi-factor reranking scores.
type ScoredSnippet struct {
	Snippet    sqlc.SearchSnippetFTSRow
	FinalScore float64
}

// RerankChapters applies temporal recency, origin bonus, and co-occurrence scoring.
func RerankChapters(items []sqlc.ChaptersFt, query string, maxChapterIndex int64) []sqlc.ChaptersFt {
	if len(items) <= 1 {
		return items
	}

	keywords := extractKeywords(query)
	scored := make([]ScoredChapter, 0, len(items))

	for rankIdx, item := range items {
		cIdx, _ := strconv.ParseInt(item.ChapterIndex, 10, 64)

		baseScore := 1.0 / (float64(rankIdx) + 1.0)

		recencyScore := 0.5
		if maxChapterIndex > 0 && cIdx > 0 {
			distance := math.Abs(float64(maxChapterIndex - cIdx))
			recencyScore = 1.0 / (1.0 + math.Log1p(distance))
		}

		originBonus := 0.0
		if cIdx > 0 && cIdx <= 5 {
			originBonus = 0.35
		}

		coOccurScore := calculateCoOccurrence(item.RawContent, keywords)

		finalScore := 0.40*baseScore + 0.35*recencyScore + 0.15*originBonus + 0.10*coOccurScore

		scored = append(scored, ScoredChapter{
			Chapter:    item,
			FinalScore: finalScore,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].FinalScore > scored[j].FinalScore
	})

	out := make([]sqlc.ChaptersFt, len(scored))
	for i, s := range scored {
		out[i] = s.Chapter
	}
	return out
}

// RerankSnippets applies temporal recency, origin bonus, and co-occurrence scoring to snippets.
func RerankSnippets(items []sqlc.SearchSnippetFTSRow, query string, maxChapterIndex int64) []sqlc.SearchSnippetFTSRow {
	if len(items) <= 1 {
		return items
	}

	keywords := extractKeywords(query)
	scored := make([]ScoredSnippet, 0, len(items))

	for rankIdx, item := range items {
		cIdx, _ := strconv.ParseInt(item.ChapterIndex, 10, 64)

		baseScore := 1.0 / (float64(rankIdx) + 1.0)

		recencyScore := 0.5
		if maxChapterIndex > 0 && cIdx > 0 {
			distance := math.Abs(float64(maxChapterIndex - cIdx))
			recencyScore = 1.0 / (1.0 + math.Log1p(distance))
		}

		originBonus := 0.0
		if cIdx > 0 && cIdx <= 5 {
			originBonus = 0.35
		}

		coOccurScore := calculateCoOccurrence(item.Snippet, keywords)

		finalScore := 0.40*baseScore + 0.35*recencyScore + 0.15*originBonus + 0.10*coOccurScore

		scored = append(scored, ScoredSnippet{
			Snippet:    item,
			FinalScore: finalScore,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].FinalScore > scored[j].FinalScore
	})

	out := make([]sqlc.SearchSnippetFTSRow, len(scored))
	for i, s := range scored {
		out[i] = s.Snippet
	}
	return out
}

// extractKeywords splits query into non-empty keywords.
func extractKeywords(query string) []string {
	fields := strings.Fields(strings.TrimSpace(query))
	var out []string
	for _, f := range fields {
		clean := strings.TrimSpace(f)
		if clean != "" {
			out = append(out, strings.ToLower(clean))
		}
	}
	return out
}

// calculateCoOccurrence scores how many keywords co-occur within the given text.
func calculateCoOccurrence(content string, keywords []string) float64 {
	if len(keywords) <= 1 {
		return 1.0
	}

	lower := strings.ToLower(content)
	matches := 0
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			matches++
		}
	}

	return float64(matches) / float64(len(keywords))
}
