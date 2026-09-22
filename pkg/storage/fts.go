package storage

import (
	"context"
	"fmt"
	"novelclaw/internal/gen/sqlc"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ftsSpecialCharsRegex = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
	htmlTagRegex         = regexp.MustCompile(`<[^>]*>`)
)

// StripHTML strips all HTML/XML tags leaving only clean text for full-text indexing.
func StripHTML(input string) string {
	cleaned := htmlTagRegex.ReplaceAllString(input, " ")
	return strings.Join(strings.Fields(cleaned), " ")
}

// SanitizeFTS5Query sanitizes search queries to prevent FTS5 syntax errors.
func SanitizeFTS5Query(query string) string {
	cleaned := ftsSpecialCharsRegex.ReplaceAllString(query, " ")
	words := strings.Fields(cleaned)
	if len(words) == 0 {
		return `""`
	}
	var quoted []string
	for _, w := range words {
		if strings.TrimSpace(w) != "" {
			quoted = append(quoted, `"`+w+`"`)
		}
	}
	return strings.Join(quoted, " ")
}

// IndexChapterFTS indexes a chapter into FTS5 virtual table.
func (s *Storage) IndexChapterFTS(ctx context.Context, arg sqlc.IndexChapterFTSParams) error {
	arg.RawContent = StripHTML(arg.RawContent)
	arg.TranslatedContent = StripHTML(arg.TranslatedContent)
	return s.q.IndexChapterFTS(ctx, arg)
}

// SearchBookContext searches context across all chapters without temporal bounds.
func (s *Storage) SearchBookContext(ctx context.Context, projectID string, query string, limit int64) ([]sqlc.ChaptersFt, error) {
	return s.SearchBookContextTemporal(ctx, projectID, query, 0, limit)
}

// SearchBookContextTemporal searches context respecting temporal horizon (chapter_index <= maxChapterIndex) and applies multi-factor reranking.
func (s *Storage) SearchBookContextTemporal(ctx context.Context, projectID string, query string, maxChapterIndex int64, limit int64) ([]sqlc.ChaptersFt, error) {
	if limit <= 0 {
		limit = 10
	}
	clean := strings.TrimSpace(query)
	if clean == "" {
		return nil, nil
	}

	candidateLimit := limit * 3
	if candidateLimit < 20 {
		candidateLimit = 20
	}

	var candidates []sqlc.ChaptersFt
	runeCount := utf8.RuneCountInString(clean)
	if runeCount < 3 {
		likePattern := "%" + clean + "%"
		rows, err := s.q.SearchBookContextLike(ctx, sqlc.SearchBookContextLikeParams{
			ProjectID:       projectID,
			MaxChapterIndex: maxChapterIndex,
			Pattern:         likePattern,
			Limit:           candidateLimit,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range rows {
			candidates = append(candidates, sqlc.ChaptersFt{
				ChapterID:         r.ChapterID,
				ProjectID:         r.ProjectID,
				ChapterIndex:      r.ChapterIndex,
				Title:             r.Title,
				RawContent:        r.RawContent,
				TranslatedContent: r.TranslatedContent,
			})
		}
	} else {
		safeQuery := SanitizeFTS5Query(query)
		rows, err := s.q.SearchBookContextFTS(ctx, sqlc.SearchBookContextFTSParams{
			Query:           safeQuery,
			ProjectID:       projectID,
			MaxChapterIndex: maxChapterIndex,
			Limit:           candidateLimit,
		})
		if err != nil {
			return nil, err
		}
		candidates = rows
	}

	reranked := RerankChapters(candidates, clean, maxChapterIndex)
	if int64(len(reranked)) > limit {
		reranked = reranked[:limit]
	}
	return reranked, nil
}

// SearchSnippet searches context snippets across all chapters without temporal bounds.
func (s *Storage) SearchSnippet(ctx context.Context, projectID string, query string, limit int64) ([]sqlc.SearchSnippetFTSRow, error) {
	return s.SearchSnippetTemporal(ctx, projectID, query, 0, limit)
}

// SearchSnippetTemporal searches context snippets respecting temporal horizon and applies multi-factor reranking.
func (s *Storage) SearchSnippetTemporal(ctx context.Context, projectID string, query string, maxChapterIndex int64, limit int64) ([]sqlc.SearchSnippetFTSRow, error) {
	if limit <= 0 {
		limit = 10
	}
	clean := strings.TrimSpace(query)
	if clean == "" {
		return nil, nil
	}

	candidateLimit := limit * 3
	if candidateLimit < 20 {
		candidateLimit = 20
	}

	var candidates []sqlc.SearchSnippetFTSRow
	runeCount := utf8.RuneCountInString(clean)
	if runeCount < 3 {
		likePattern := "%" + clean + "%"
		rows, err := s.q.SearchBookContextLike(ctx, sqlc.SearchBookContextLikeParams{
			ProjectID:       projectID,
			MaxChapterIndex: maxChapterIndex,
			Pattern:         likePattern,
			Limit:           candidateLimit,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range rows {
			snippet := createManualSnippet(r.RawContent, clean)
			candidates = append(candidates, sqlc.SearchSnippetFTSRow{
				ChapterID:    r.ChapterID,
				ProjectID:    r.ProjectID,
				ChapterIndex: r.ChapterIndex,
				Title:        r.Title,
				Snippet:      snippet,
			})
		}
	} else {
		safeQuery := SanitizeFTS5Query(query)
		rows, err := s.q.SearchSnippetFTS(ctx, sqlc.SearchSnippetFTSParams{
			Query:           safeQuery,
			ProjectID:       projectID,
			MaxChapterIndex: maxChapterIndex,
			Limit:           candidateLimit,
		})
		if err != nil {
			return nil, err
		}
		candidates = rows
	}

	reranked := RerankSnippets(candidates, clean, maxChapterIndex)
	if int64(len(reranked)) > limit {
		reranked = reranked[:limit]
	}
	return reranked, nil
}

// createManualSnippet creates a preview snippet for LIKE fallback results.
func createManualSnippet(content, keyword string) string {
	idx := strings.Index(content, keyword)
	if idx == -1 {
		return content
	}

	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + len(keyword) + 40
	if end > len(content) {
		end = len(content)
	}

	prefix := "... "
	if start == 0 {
		prefix = ""
	}
	suffix := " ..."
	if end == len(content) {
		suffix = ""
	}

	sub := content[start:end]
	highlighted := strings.Replace(sub, keyword, fmt.Sprintf("[MARK]%s[/MARK]", keyword), 1)
	return prefix + highlighted + suffix
}

// DeleteChapterFTS removes FTS5 index for a chapter.
func (s *Storage) DeleteChapterFTS(ctx context.Context, chapterID string) error {
	return s.q.DeleteChapterFTS(ctx, chapterID)
}

// ClearProjectFTS removes all FTS5 indices for a project.
func (s *Storage) ClearProjectFTS(ctx context.Context, projectID string) error {
	return s.q.ClearProjectFTS(ctx, projectID)
}
