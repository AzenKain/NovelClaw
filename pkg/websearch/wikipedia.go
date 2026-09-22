package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WikipediaSearcher searches multilingual Wikipedia entries via OpenSearch and Query API.
type WikipediaSearcher struct {
	client   *http.Client
	langCode string
}

// NewWikipediaSearcher creates a new WikipediaSearcher for the specified language code.
func NewWikipediaSearcher(langCode string, timeout time.Duration) *WikipediaSearcher {
	if langCode == "" {
		langCode = "ja"
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &WikipediaSearcher{
		client:   &http.Client{Timeout: timeout},
		langCode: langCode,
	}
}

// Name returns the searcher identifier.
func (w *WikipediaSearcher) Name() string {
	return "wikipedia_" + w.langCode
}

// Search searches Wikipedia articles for the query.
func (w *WikipediaSearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 3
	}

	searchURL := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=opensearch&search=%s&limit=%d&namespace=0&format=json",
		w.langCode, url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "NekoNovel-ResearchEngine/1.0")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var raw []any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if len(raw) < 4 {
		return nil, nil
	}

	titles, _ := raw[1].([]any)
	descriptions, _ := raw[2].([]any)
	urls, _ := raw[3].([]any)

	var results []SearchResult
	for i := 0; i < len(titles) && i < limit; i++ {
		t, _ := titles[i].(string)
		desc, _ := descriptions[i].(string)
		u, _ := urls[i].(string)

		if t != "" {
			results = append(results, SearchResult{
				Title:   t,
				URL:     u,
				Snippet: desc,
				Source:  "wikipedia_" + w.langCode,
			})
		}
	}

	if len(results) > 0 && results[0].Snippet == "" {
		extract, err := w.FetchExtract(ctx, results[0].Title)
		if err == nil && extract != "" {
			results[0].Snippet = extract
		}
	}

	return results, nil
}

// FetchExtract fetches the lead paragraph extract of a Wikipedia article.
func (w *WikipediaSearcher) FetchExtract(ctx context.Context, title string) (string, error) {
	extractURL := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=query&prop=extracts&exintro&explaintext&titles=%s&format=json",
		w.langCode, url.QueryEscape(title))

	req, err := http.NewRequestWithContext(ctx, "GET", extractURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "NekoNovel-ResearchEngine/1.0")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var raw struct {
		Query struct {
			Pages map[string]struct {
				Title   string `json:"title"`
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", err
	}

	for _, page := range raw.Query.Pages {
		if page.Extract != "" {
			return strings.TrimSpace(page.Extract), nil
		}
	}

	return "", nil
}
