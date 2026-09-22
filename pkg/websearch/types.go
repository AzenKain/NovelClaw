package websearch

import (
	"context"
)

// SearchResult represents a web search hit.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"`
}

// ResearchBrief encapsulates aggregated research findings for the agent.
type ResearchBrief struct {
	Query       string         `json:"query"`
	Summary     string         `json:"summary"`
	DeepContent string         `json:"deep_content"`
	SourceURL   string         `json:"source_url"`
	Results     []SearchResult `json:"results"`
}

// WebSearcher defines the searcher interface.
type WebSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	Name() string
}
