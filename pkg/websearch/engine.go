package websearch

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Yiling-J/theine-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

// ResearchEngine orchestrates multi-source, any-to-any web research with caching.
type ResearchEngine struct {
	timeout        time.Duration
	ddg            *DuckDuckGoSearcher
	deepReader     *DeepReader
	wikiMu         sync.RWMutex
	wikiSearchers  map[string]*WikipediaSearcher
	cache          *theine.Cache[string, *ResearchBrief]
	sfGroup        singleflight.Group
}

// NewResearchEngine creates a new ResearchEngine supporting arbitrary languages (Any-to-Any).
func NewResearchEngine(timeout time.Duration) *ResearchEngine {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	c, _ := theine.NewBuilder[string, *ResearchBrief](1000).Build()

	return &ResearchEngine{
		timeout:       timeout,
		ddg:           NewDuckDuckGoSearcher(timeout),
		deepReader:    NewDeepReader(timeout, 5000),
		wikiSearchers: make(map[string]*WikipediaSearcher),
		cache:         c,
	}
}

// getWikiSearcher dynamically returns or creates a Wikipedia searcher for any ISO language code.
func (re *ResearchEngine) getWikiSearcher(lang string) *WikipediaSearcher {
	langCode := normalizeLangCode(lang)

	re.wikiMu.RLock()
	searcher, ok := re.wikiSearchers[langCode]
	re.wikiMu.RUnlock()
	if ok {
		return searcher
	}

	re.wikiMu.Lock()
	defer re.wikiMu.Unlock()
	if searcher, ok = re.wikiSearchers[langCode]; ok {
		return searcher
	}

	searcher = NewWikipediaSearcher(langCode, re.timeout)
	re.wikiSearchers[langCode] = searcher
	return searcher
}

// normalizeLangCode normalizes language name or code into ISO 639-1 format.
func normalizeLangCode(lang string) string {
	clean := strings.ToLower(strings.TrimSpace(lang))
	switch clean {
	case "japanese", "jp", "ja":
		return "ja"
	case "chinese", "zh-cn", "zh-tw", "zh":
		return "zh"
	case "vietnamese", "vi":
		return "vi"
	case "korean", "ko":
		return "ko"
	case "french", "fr":
		return "fr"
	case "german", "de":
		return "de"
	case "spanish", "es":
		return "es"
	case "russian", "ru":
		return "ru"
	case "italian", "it":
		return "it"
	case "portuguese", "pt":
		return "pt"
	case "english", "en":
		return "en"
	default:
		if len(clean) == 2 {
			return clean
		}
		return "en"
	}
}

// Research conducts deep research for cultural terms, slang, or names across any source language.
func (re *ResearchEngine) Research(ctx context.Context, query string, preferredLang string) (*ResearchBrief, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	langCode := normalizeLangCode(preferredLang)
	cacheKey := fmt.Sprintf("%s:%s", langCode, query)

	if re.cache != nil {
		if entry, ok := re.cache.Get(cacheKey); ok && entry != nil {
			log.Debug().Str("query", query).Str("lang", langCode).Msg("research cache hit")
			return entry, nil
		}
	}

	res, err, _ := re.sfGroup.Do(cacheKey, func() (interface{}, error) {
		brief := &ResearchBrief{
			Query: query,
		}

		var allResults []SearchResult

		primaryWiki := re.getWikiSearcher(langCode)
		wikiRes, err := primaryWiki.Search(ctx, query, 3)
		if err == nil && len(wikiRes) > 0 {
			allResults = append(allResults, wikiRes...)
		}

		if len(allResults) == 0 && langCode != "en" {
			fallbackWiki := re.getWikiSearcher("en")
			enRes, err := fallbackWiki.Search(ctx, query, 3)
			if err == nil && len(enRes) > 0 {
				allResults = append(allResults, enRes...)
			}
		}

		ddgRes, err := re.ddg.Search(ctx, query, 5)
		if err == nil && len(ddgRes) > 0 {
			allResults = append(allResults, ddgRes...)
		}

		brief.Results = allResults

		if len(allResults) == 0 {
			brief.Summary = "No information found on free public sources."
			return brief, nil
		}

		topResult := allResults[0]
		brief.SourceURL = topResult.URL
		if topResult.Snippet != "" {
			brief.Summary = topResult.Snippet
		} else {
			brief.Summary = topResult.Title
		}

		if topResult.URL != "" && strings.HasPrefix(topResult.URL, "http") {
			deepContent, err := re.deepReader.ReadURL(ctx, topResult.URL)
			if err == nil && deepContent != "" {
				brief.DeepContent = deepContent
			}
		}

		if re.cache != nil {
			_ = re.cache.SetWithTTL(cacheKey, brief, 1, 1*time.Hour)
		}

		log.Info().
			Str("query", query).
			Str("lang", langCode).
			Int("results", len(allResults)).
			Str("top_url", brief.SourceURL).
			Msg("any-to-any research completed and cached")

		return brief, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*ResearchBrief), nil
}

// FormatForPrompt formats research findings into Markdown suitable for LLM injection.
func (brief *ResearchBrief) FormatForPrompt() string {
	if brief == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### [RESEARCH FINDINGS FOR '%s']\n", brief.Query))
	if brief.SourceURL != "" {
		sb.WriteString(fmt.Sprintf("- Source: %s\n", brief.SourceURL))
	}
	if brief.Summary != "" {
		sb.WriteString(fmt.Sprintf("- Summary: %s\n", brief.Summary))
	}
	if brief.DeepContent != "" {
		sb.WriteString("\n- Extracted Article Body:\n```text\n")
		sb.WriteString(brief.DeepContent)
		sb.WriteString("\n```\n")
	}

	return sb.String()
}
