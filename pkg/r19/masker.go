package r19

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"novelclaw/pkg/trie"
)

var (
	digitAgeRegex = regexp.MustCompile(`(?i)\b([1-9]|1[0-7])\s*(tuổi|岁|歳|才|살|(?:years?\s*old|yo|-year-old)\b)`)
	cjkMinorRegex = regexp.MustCompile(`(十二|十三|十四|十五|十六|十七|十四三|十四五)\s*(岁|歳|才)`)
	krMinorRegex  = regexp.MustCompile(`(열두|열세|열네|열다섯|열여섯|열일곱)\s*살`)
	enMinorRegex  = regexp.MustCompile(`(?i)\b(twelve|thirteen|fourteen|fifteen|sixteen|seventeen)\s*(?:years?\s*old|yo)\b`)
)

// Masker executes fast multi-lingual Aho-Corasick masking and dynamic age neutralization.
type Masker struct {
	mu      sync.RWMutex
	matcher *trie.Matcher
	terms   map[string]string
}

// NewMasker compiles a new Masker with given term mappings.
func NewMasker(customTerms map[string]string) *Masker {
	m := &Masker{
		matcher: trie.NewMatcher(),
		terms:   make(map[string]string),
	}

	for k, v := range DefaultTerms {
		m.terms[k] = v
	}
	for k, v := range customTerms {
		m.terms[k] = v
	}

	for pattern, translation := range m.terms {
		m.matcher.AddPattern(pattern, translation)
	}
	m.matcher.Build()
	return m
}

// AddTerm registers a new sensitive term and rebuilds the Trie index.
func (m *Masker) AddTerm(source, translation string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleanSrc := strings.TrimSpace(source)
	if cleanSrc == "" {
		return
	}
	m.terms[cleanSrc] = translation
	m.matcher.AddPattern(cleanSrc, translation)
	m.matcher.Build()
}

// LoadFromFile loads external term mappings from a disk file into the Masker.
func (m *Masker) LoadFromFile(path string) error {
	terms, err := LoadTermsFromFile(path)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range terms {
		m.terms[k] = v
		m.matcher.AddPattern(k, v)
	}
	m.matcher.Build()
	return nil
}

// Mask neutralizes sensitive terms and underage age references into robust tags.
func (m *Masker) Mask(text string, cfg Config) MaskResult {
	if !cfg.Enabled || text == "" {
		return MaskResult{MaskedText: text}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var entries []Entry
	nextID := 1
	currentText := text

	if cfg.NeutralizeAges {
		currentText, entries, nextID = m.neutralizeAges(currentText, entries, nextID, cfg.TargetLang)
	}

	matches := m.matcher.FindAll(currentText)
	ageCount := len(entries)
	if len(matches) == 0 {
		return MaskResult{
			MaskedText:          currentText,
			Entries:             entries,
			TermsMaskedCount:    0,
			AgeNeutralizedCount: ageCount,
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Start != matches[j].Start {
			return matches[i].Start < matches[j].Start
		}
		return matches[i].End > matches[j].End
	})

	var nonOverlapping []trie.MatchResult
	lastEnd := 0
	for _, match := range matches {
		if match.Start >= lastEnd {
			nonOverlapping = append(nonOverlapping, match)
			lastEnd = match.End
		}
	}

	normTgt := strings.ToLower(strings.TrimSpace(cfg.TargetLang))

	var sb strings.Builder
	cursor := 0
	for _, match := range nonOverlapping {
		sb.WriteString(currentText[cursor:match.Start])
		tag := fmt.Sprintf("<r:%d/>", nextID)
		translation, _ := match.Payload.(string)
		// If target language is English and pattern is English, preserve English source
		if strings.HasPrefix(normTgt, "en") && isPureASCII(match.Pattern) {
			translation = match.Pattern
		}
		entries = append(entries, Entry{
			ID:          nextID,
			Source:      match.Pattern,
			Translation: translation,
			Tag:         tag,
			Category:    CategorySexual,
		})
		nextID++
		sb.WriteString(tag)
		cursor = match.End
	}
	sb.WriteString(currentText[cursor:])

	return MaskResult{
		MaskedText:          sb.String(),
		Entries:             entries,
		TermsMaskedCount:    len(nonOverlapping),
		AgeNeutralizedCount: ageCount,
	}
}

func isPureASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func (m *Masker) neutralizeAges(text string, entries []Entry, nextID int, targetLang string) (string, []Entry, int) {
	normTgt := strings.ToLower(strings.TrimSpace(targetLang))
	if normTgt == "" {
		normTgt = "vietnamese"
	}

	formatAge := func(numVal int, rawStr string) string {
		switch {
		case strings.HasPrefix(normTgt, "en"):
			return fmt.Sprintf("%d years old", numVal)
		case strings.HasPrefix(normTgt, "ja"):
			return fmt.Sprintf("%d歳", numVal)
		case strings.HasPrefix(normTgt, "zh"):
			return fmt.Sprintf("%d岁", numVal)
		case strings.HasPrefix(normTgt, "ko"):
			return fmt.Sprintf("%d살", numVal)
		case strings.HasPrefix(normTgt, "fr"):
			return fmt.Sprintf("%d ans", numVal)
		case strings.HasPrefix(normTgt, "de"):
			return fmt.Sprintf("%d Jahre alt", numVal)
		case strings.HasPrefix(normTgt, "es"):
			return fmt.Sprintf("%d años", numVal)
		default:
			// Vietnamese
			return fmt.Sprintf("%d tuổi", numVal)
		}
	}

	updatedText := digitAgeRegex.ReplaceAllStringFunc(text, func(match string) string {
		tag := fmt.Sprintf("<r:age_%d/>", nextID)
		sub := digitAgeRegex.FindStringSubmatch(match)
		ageVal := sub[1]
		numVal, _ := strconv.Atoi(ageVal)
		trans := formatAge(numVal, ageVal)
		entries = append(entries, Entry{
			ID:          nextID,
			Source:      match,
			Translation: trans,
			Tag:         tag,
			Category:    CategoryUnderage,
		})
		nextID++
		return tag
	})

	updatedText = cjkMinorRegex.ReplaceAllStringFunc(updatedText, func(match string) string {
		tag := fmt.Sprintf("<r:age_%d/>", nextID)
		sub := cjkMinorRegex.FindStringSubmatch(match)
		cjkAge := sub[1]
		numVal := cjkNumeralToNumber(cjkAge)
		trans := formatAge(numVal, cjkAge)
		entries = append(entries, Entry{
			ID:          nextID,
			Source:      match,
			Translation: trans,
			Tag:         tag,
			Category:    CategoryUnderage,
		})
		nextID++
		return tag
	})

	updatedText = krMinorRegex.ReplaceAllStringFunc(updatedText, func(match string) string {
		tag := fmt.Sprintf("<r:age_%d/>", nextID)
		sub := krMinorRegex.FindStringSubmatch(match)
		krAge := sub[1]
		numVal := krNumeralToNumber(krAge)
		trans := formatAge(numVal, krAge)
		entries = append(entries, Entry{
			ID:          nextID,
			Source:      match,
			Translation: trans,
			Tag:         tag,
			Category:    CategoryUnderage,
		})
		nextID++
		return tag
	})

	updatedText = enMinorRegex.ReplaceAllStringFunc(updatedText, func(match string) string {
		tag := fmt.Sprintf("<r:age_%d/>", nextID)
		sub := enMinorRegex.FindStringSubmatch(match)
		enAge := strings.ToLower(sub[1])
		numVal := enNumeralToNumber(enAge)
		trans := formatAge(numVal, enAge)
		entries = append(entries, Entry{
			ID:          nextID,
			Source:      match,
			Translation: trans,
			Tag:         tag,
			Category:    CategoryUnderage,
		})
		nextID++
		return tag
	})

	return updatedText, entries, nextID
}

func cjkNumeralToNumber(num string) int {
	switch num {
	case "十二":
		return 12
	case "十三":
		return 13
	case "十四":
		return 14
	case "十五":
		return 15
	case "十六":
		return 16
	case "十七":
		return 17
	default:
		return 0
	}
}

func krNumeralToNumber(num string) int {
	switch num {
	case "열두":
		return 12
	case "열세":
		return 13
	case "열네":
		return 14
	case "열다섯":
		return 15
	case "열여섯":
		return 16
	case "열일곱":
		return 17
	default:
		return 0
	}
}

func enNumeralToNumber(num string) int {
	switch strings.ToLower(num) {
	case "twelve":
		return 12
	case "thirteen":
		return 13
	case "fourteen":
		return 14
	case "fifteen":
		return 15
	case "sixteen":
		return 16
	case "seventeen":
		return 17
	default:
		return 0
	}
}

func parseAgeString(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}
