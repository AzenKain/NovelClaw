package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// DeepReader extracts clean article content from URLs without external dependencies.
type DeepReader struct {
	client    *http.Client
	maxLength int
}

// NewDeepReader creates a new DeepReader.
func NewDeepReader(timeout time.Duration, maxLength int) *DeepReader {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if maxLength <= 0 {
		maxLength = 5000
	}
	return &DeepReader{
		client: &http.Client{
			Timeout: timeout,
		},
		maxLength: maxLength,
	}
}

// ReadURL fetches the HTML from a URL and extracts clean markdown-like text.
func (dr *DeepReader) ReadURL(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ja,zh,vi,en;q=0.9")

	resp, err := dr.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch url failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	lr := io.LimitReader(resp.Body, 2*1024*1024)
	return dr.ExtractContent(lr)
}

// ExtractContent parses HTML reader and strips boilerplate into clean readable text.
func (dr *DeepReader) ExtractContent(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", fmt.Errorf("parse html failed: %w", err)
	}

	cleanDom(doc)

	mainNode := findMainContentNode(doc)
	if mainNode == nil {
		mainNode = doc
	}

	var sb strings.Builder
	extractCleanText(mainNode, &sb, dr.maxLength)

	result := strings.TrimSpace(sb.String())
	result = cleanExcessNewlines(result)

	if len([]rune(result)) > dr.maxLength {
		runes := []rune(result)
		result = string(runes[:dr.maxLength]) + "\n\n[...Content truncated...]"
	}

	return result, nil
}

// cleanDom removes non-content nodes such as script, style, nav, and footers.
func cleanDom(n *html.Node) {
	var toRemove []*html.Node

	var inspect func(*html.Node)
	inspect = func(node *html.Node) {
		if node == nil {
			return
		}

		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			switch tag {
			case "script", "style", "noscript", "nav", "footer", "header", "aside",
				"iframe", "svg", "form", "button", "input", "select", "textarea":
				toRemove = append(toRemove, node)
				return
			}

			for _, attr := range node.Attr {
				key := strings.ToLower(attr.Key)
				if key == "class" || key == "id" {
					val := strings.ToLower(attr.Val)
					if strings.Contains(val, "sidebar") ||
						strings.Contains(val, "comment") ||
						strings.Contains(val, "advertisement") ||
						strings.Contains(val, "ad-") ||
						strings.Contains(val, "banner") ||
						strings.Contains(val, "cookie") ||
						strings.Contains(val, "nav") ||
						strings.Contains(val, "menu") {
						toRemove = append(toRemove, node)
						return
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			inspect(c)
		}
	}

	inspect(n)

	for _, node := range toRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// findMainContentNode locates the primary content container element.
func findMainContentNode(n *html.Node) *html.Node {
	var mainNode *html.Node

	var find func(*html.Node)
	find = func(node *html.Node) {
		if node == nil || mainNode != nil {
			return
		}
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if tag == "article" || tag == "main" {
				mainNode = node
				return
			}
			for _, attr := range node.Attr {
				if attr.Key == "role" && attr.Val == "main" {
					mainNode = node
					return
				}
				if attr.Key == "id" || attr.Key == "class" {
					val := strings.ToLower(attr.Val)
					if strings.Contains(val, "article") || strings.Contains(val, "content") || strings.Contains(val, "post-body") {
						mainNode = node
						return
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}

	find(n)
	return mainNode
}

// extractCleanText traverses the node tree and writes simplified Markdown format.
func extractCleanText(n *html.Node, sb *strings.Builder, maxRunes int) {
	if n == nil || sb.Len() >= maxRunes*4 {
		return
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
		return
	}

	if n.Type == html.ElementNode {
		tag := strings.ToLower(n.Data)
		isHeading := false
		switch tag {
		case "h1":
			sb.WriteString("\n\n# ")
			isHeading = true
		case "h2":
			sb.WriteString("\n\n## ")
			isHeading = true
		case "h3":
			sb.WriteString("\n\n### ")
			isHeading = true
		case "h4", "h5", "h6":
			sb.WriteString("\n\n#### ")
			isHeading = true
		case "p":
			sb.WriteString("\n\n")
		case "li":
			sb.WriteString("\n- ")
		case "br":
			sb.WriteString("\n")
		case "blockquote":
			sb.WriteString("\n> ")
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractCleanText(c, sb, maxRunes)
		}

		if isHeading || tag == "p" || tag == "blockquote" {
			sb.WriteString("\n")
		}
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractCleanText(c, sb, maxRunes)
	}
}

// cleanExcessNewlines condenses multiple consecutive blank lines.
func cleanExcessNewlines(s string) string {
	lines := strings.Split(s, "\n")
	var cleaned []string
	emptyCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			emptyCount++
			if emptyCount <= 2 {
				cleaned = append(cleaned, "")
			}
		} else {
			emptyCount = 0
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}
