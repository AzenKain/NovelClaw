package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// DuckDuckGoSearcher implements WebSearcher using DuckDuckGo free endpoints.
type DuckDuckGoSearcher struct {
	client *http.Client
}

// NewDuckDuckGoSearcher creates a new DuckDuckGoSearcher.
func NewDuckDuckGoSearcher(timeout time.Duration) *DuckDuckGoSearcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &DuckDuckGoSearcher{
		client: &http.Client{Timeout: timeout},
	}
}

// Name returns the searcher identifier.
func (d *DuckDuckGoSearcher) Name() string {
	return "duckduckgo"
}

// Search queries DuckDuckGo Instant Answer API with HTML fallback.
func (d *DuckDuckGoSearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	results, err := d.searchInstantAnswer(ctx, query)
	if err == nil && len(results) > 0 {
		if len(results) > limit {
			results = results[:limit]
		}
		return results, nil
	}

	return d.searchHTML(ctx, query, limit)
}

// searchInstantAnswer queries the DuckDuckGo JSON API.
func (d *DuckDuckGoSearcher) searchInstantAnswer(ctx context.Context, query string) ([]SearchResult, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "NekoNovel-ResearchEngine/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var raw struct {
		Abstract       string `json:"Abstract"`
		AbstractSource string `json:"AbstractSource"`
		AbstractURL    string `json:"AbstractURL"`
		Heading        string `json:"Heading"`
		RelatedTopics  []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var out []SearchResult
	if raw.Abstract != "" {
		out = append(out, SearchResult{
			Title:   raw.Heading,
			URL:     raw.AbstractURL,
			Snippet: raw.Abstract,
			Source:  "duckduckgo_instant",
		})
	}

	for _, rt := range raw.RelatedTopics {
		if rt.Text != "" && rt.FirstURL != "" {
			out = append(out, SearchResult{
				Title:   rt.Text,
				URL:     rt.FirstURL,
				Snippet: rt.Text,
				Source:  "duckduckgo_instant",
			})
		}
	}

	return out, nil
}

// searchHTML queries DuckDuckGo HTML endpoint as a fallback.
func (d *DuckDuckGoSearcher) searchHTML(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	return parseDuckDuckGoHTML(resp.Body, limit)
}

// parseDuckDuckGoHTML parses search result snippets from DuckDuckGo HTML.
func parseDuckDuckGoHTML(r io.Reader, limit int) ([]SearchResult, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if len(results) >= limit {
			return
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			isSnippet := false
			var href string
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "result__snippet") {
					isSnippet = true
				}
				if attr.Key == "href" {
					href = attr.Val
				}
			}

			if isSnippet && href != "" {
				text := extractText(n)
				if text != "" {
					actualURL := cleanDuckDuckGoURL(href)
					results = append(results, SearchResult{
						Title:   "Web Result",
						URL:     actualURL,
						Snippet: strings.TrimSpace(text),
						Source:  "duckduckgo",
					})
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return results, nil
}

// cleanDuckDuckGoURL unwraps the target destination URL from DuckDuckGo redirect format.
func cleanDuckDuckGoURL(rawHref string) string {
	if strings.Contains(rawHref, "uddg=") {
		parts := strings.Split(rawHref, "uddg=")
		if len(parts) > 1 {
			targetPart := strings.Split(parts[1], "&")[0]
			decoded, err := url.QueryUnescape(targetPart)
			if err == nil {
				return decoded
			}
		}
	}
	if strings.HasPrefix(rawHref, "//") {
		return "https:" + rawHref
	}
	return rawHref
}

// extractText recursively extracts plain text from an HTML node tree.
func extractText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
