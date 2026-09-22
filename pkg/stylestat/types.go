package stylestat

// RepetitivePhrase records a phrase pattern that was overused in the text.
type RepetitivePhrase struct {
	Phrase string `json:"phrase"`
	Count  int    `json:"count"`
}

// StyleReport contains deterministic novel stylistic metrics and quality signals.
type StyleReport struct {
	TotalSentences         int                `json:"total_sentences"`
	TotalWords             int                `json:"total_words"`
	DuplicateSentenceCount int                `json:"duplicate_sentence_count"`
	RepetitivePhrases      []RepetitivePhrase `json:"repetitive_phrases"`
	GlossaryAdherenceRate  float64            `json:"glossary_adherence_rate"`
	QualityScore           float64            `json:"quality_score"`
}
