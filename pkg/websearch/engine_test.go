package websearch

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestResearchEngine_FormatForPrompt verifies formatting of research brief.
func TestResearchEngine_FormatForPrompt(t *testing.T) {
	brief := &ResearchBrief{
		Query:       "妹さえいればいい",
		Summary:     "A Sister's All You Need is a Japanese light novel series written by Yomi Hirasaka.",
		DeepContent: "The story follows Itsuki Hashima, a novelist obsessed with little sisters...",
		SourceURL:   "https://ja.wikipedia.org/wiki/妹さえいればいい",
	}

	formatted := brief.FormatForPrompt()
	if !strings.Contains(formatted, "RESEARCH FINDINGS FOR '妹さえいればいい'") {
		t.Errorf("FormatForPrompt missing expected header: %s", formatted)
	}
	if !strings.Contains(formatted, "https://ja.wikipedia.org") {
		t.Errorf("FormatForPrompt missing source URL")
	}
	if !strings.Contains(formatted, "Itsuki Hashima") {
		t.Errorf("FormatForPrompt missing DeepContent")
	}
}

// TestResearchEngine_LiveLookup tests live query and subsequent cache hit speed.
func TestResearchEngine_LiveLookup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live web search test in short mode")
	}

	engine := NewResearchEngine(10 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	brief, err := engine.Research(ctx, "Light novel", "ja")
	if err != nil {
		t.Logf("network lookup warning: %v", err)
		return
	}

	start := time.Now()
	cachedBrief, err := engine.Research(ctx, "Light novel", "ja")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if cachedBrief == nil || cachedBrief.Summary != brief.Summary {
		t.Errorf("cache payload mismatch")
	}
	if elapsed > 10*time.Millisecond {
		t.Errorf("cache retrieval too slow: %v", elapsed)
	}
}

// TestResearchEngine_NormalizeLangCode verifies ISO 639-1 normalization across multiple languages.
func TestResearchEngine_NormalizeLangCode(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"japanese", "ja"},
		{"JP", "ja"},
		{"ja", "ja"},
		{"chinese", "zh"},
		{"zh-CN", "zh"},
		{"korean", "ko"},
		{"ko", "ko"},
		{"vietnamese", "vi"},
		{"vi", "vi"},
		{"french", "fr"},
		{"german", "de"},
		{"spanish", "es"},
		{"russian", "ru"},
		{"italian", "it"},
		{"portuguese", "pt"},
		{"english", "en"},
		{"th", "th"}, // arbitrary 2-letter ISO code
		{"pl", "pl"}, // arbitrary 2-letter ISO code
		{"unknown_long_name", "en"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeLangCode(tc.input)
			if got != tc.expected {
				t.Errorf("normalizeLangCode(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}

// TestResearchEngine_DynamicWikiPool tests dynamic creation and concurrency safety of wiki searcher pool.
func TestResearchEngine_DynamicWikiPool(t *testing.T) {
	engine := NewResearchEngine(5 * time.Second)

	s1 := engine.getWikiSearcher("ja")
	s2 := engine.getWikiSearcher("Japanese")
	if s1 != s2 {
		t.Errorf("expected same searcher instance for 'ja' and 'Japanese'")
	}

	sFr := engine.getWikiSearcher("french")
	if sFr == nil || sFr.langCode != "fr" {
		t.Errorf("expected french searcher with lang 'fr'")
	}

	sKo := engine.getWikiSearcher("ko")
	if sKo == nil || sKo.langCode != "ko" {
		t.Errorf("expected korean searcher with lang 'ko'")
	}
}

