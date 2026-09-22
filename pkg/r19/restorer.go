package r19

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var sentenceEndRegex = regexp.MustCompile(`[.?!…“"”\n\-—]\s*$`)

// Restorer handles fuzzy tag resolution and semantic restoration into target language.
type Restorer struct{}

// NewRestorer creates an instance of Restorer.
func NewRestorer() *Restorer {
	return &Restorer{}
}

// Restore substitutes masked tags with their target language translations.
func (r *Restorer) Restore(text string, entries []Entry, isTitle bool) string {
	if len(entries) == 0 || text == "" {
		return text
	}

	result := text
	for _, entry := range entries {
		patternStr := buildFuzzyTagPattern(entry.Tag)
		re, err := regexp.Compile(patternStr)
		if err != nil {
			result = strings.ReplaceAll(result, entry.Tag, entry.Translation)
			continue
		}

		result = re.ReplaceAllStringFunc(result, func(match string) string {
			idx := strings.Index(result, match)
			target := entry.Translation
			if isTitle {
				return toTitleCase(target)
			}
			if idx == 0 || isSentenceStart(result[:idx]) {
				return uppercaseFirst(target)
			}
			return target
		})
	}

	return result
}

func buildFuzzyTagPattern(tag string) string {
	cleaned := strings.Trim(tag, "<>/ ")
	parts := strings.Split(cleaned, ":")
	if len(parts) < 2 {
		return regexp.QuoteMeta(tag)
	}

	prefix := regexp.QuoteMeta(parts[0])
	val := regexp.QuoteMeta(parts[1])

	return fmt.Sprintf(`(?:<|\{|\[|⟦)\s*%s\s*:\s*%s\s*(?:/>|>|\}|\]|⟧)|%s`, prefix, val, regexp.QuoteMeta(tag))
}

func isSentenceStart(prefix string) bool {
	trimmed := strings.TrimRightFunc(prefix, unicode.IsSpace)
	if trimmed == "" {
		return true
	}
	return sentenceEndRegex.MatchString(trimmed)
}

func uppercaseFirst(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = uppercaseFirst(w)
	}
	return strings.Join(words, " ")
}
