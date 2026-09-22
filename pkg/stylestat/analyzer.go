package stylestat

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var sentenceSplitRegex = regexp.MustCompile(`[.?!…\n]+`)

// Analyzer computes deterministic stylistic statistics and detects AI verbal tics.
type Analyzer struct{}

// NewAnalyzer creates an initialized Analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Compute analyzes translated text and produces a comprehensive StyleReport.
func (a *Analyzer) Compute(text string, glossaryTerms map[string]string) StyleReport {
	if strings.TrimSpace(text) == "" {
		return StyleReport{QualityScore: 10.0}
	}

	rawSentences := sentenceSplitRegex.Split(text, -1)
	var sentences []string
	sentenceCounts := make(map[string]int)
	dupSentenceCount := 0

	for _, s := range rawSentences {
		clean := strings.TrimSpace(s)
		if len(clean) > 5 {
			sentences = append(sentences, clean)
			norm := strings.ToLower(clean)
			sentenceCounts[norm]++
			if sentenceCounts[norm] == 2 {
				dupSentenceCount++
			}
		}
	}

	words := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})

	phraseCounts := make(map[string]int)
	for i := 0; i+3 <= len(words); i++ {
		phrase := strings.ToLower(strings.Join(words[i:i+3], " "))
		phraseCounts[phrase]++
	}

	var repetitive []RepetitivePhrase
	for phrase, count := range phraseCounts {
		if count >= 4 {
			repetitive = append(repetitive, RepetitivePhrase{
				Phrase: phrase,
				Count:  count,
			})
		}
	}

	sort.Slice(repetitive, func(i, j int) bool {
		return repetitive[i].Count > repetitive[j].Count
	})
	if len(repetitive) > 5 {
		repetitive = repetitive[:5]
	}

	adherenceRate := 1.0
	if len(glossaryTerms) > 0 {
		matched := 0
		total := 0
		lowerText := strings.ToLower(text)
		for _, expectedVI := range glossaryTerms {
			cleanVI := strings.TrimSpace(strings.ToLower(expectedVI))
			if cleanVI != "" {
				total++
				if strings.Contains(lowerText, cleanVI) {
					matched++
				}
			}
		}
		if total > 0 {
			adherenceRate = float64(matched) / float64(total)
		}
	}

	score := 10.0
	score -= float64(dupSentenceCount) * 1.5
	score -= float64(len(repetitive)) * 0.5
	score -= (1.0 - adherenceRate) * 2.0

	if score < 0.0 {
		score = 0.0
	} else if score > 10.0 {
		score = 10.0
	}

	return StyleReport{
		TotalSentences:         len(sentences),
		TotalWords:             len(words),
		DuplicateSentenceCount: dupSentenceCount,
		RepetitivePhrases:      repetitive,
		GlossaryAdherenceRate:  adherenceRate,
		QualityScore:           score,
	}
}
